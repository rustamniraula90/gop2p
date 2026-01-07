package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rustamniraula90/gop2p/protocol"
)

const (
	DefaultChunkSize = 16 * 1024 // 16 KB for best balance of reliability and fewer roundtrips
)

type DownloadStatus string

const (
	DownloadPending    DownloadStatus = "pending"
	DownloadRequesting DownloadStatus = "requesting"
	DownloadActive     DownloadStatus = "active"
	DownloadComplete   DownloadStatus = "complete"
	DownloadError      DownloadStatus = "error"
)

type ChunkStatus string

const (
	ChunkPending    ChunkStatus = "pending"
	ChunkRequesting ChunkStatus = "requesting"
	ChunkReceived   ChunkStatus = "received"
	ChunkError      ChunkStatus = "error"
)

type Chunk struct {
	Index      int         `json:"index"`
	Status     ChunkStatus `json:"status"`
	SHA        string      `json:"sha"`
	RetryCount int         `json:"retry_count"`
}

type Download struct {
	RequestID      string         `json:"requestId"`
	PeerID         string         `json:"peerId"`
	PeerName       string         `json:"peerName"`
	FileName       string         `json:"fileName"`
	OriginalSize   int64          `json:"original_size"`
	CompressedSize int64          `json:"compressed_size"`
	ChunkSize      int            `json:"chunk_size"`
	ChunkCount     int            `json:"chunk_count"`
	Chunks         []*Chunk       `json:"chunks"`
	Status         DownloadStatus `json:"status"`
	StartTime      time.Time      `json:"start_time"`
	TempDir        string         `json:"temp_dir"`
	Progress       float64        `json:"progress"` // 0.0 to 1.0
	mu             sync.RWMutex
}

type DownloadManager struct {
	Downloads  map[string]*Download
	mu         sync.RWMutex
	OnProgress func(requestID, status string, progress float64, chunkIndex int, chunkStatus string)
}

func NewDownloadManager() *DownloadManager {
	return &DownloadManager{
		Downloads: make(map[string]*Download),
	}
}

func (dm *DownloadManager) StartDownload(requestID, peerID, peerName, fileName string) *Download {
	download := &Download{
		RequestID: requestID,
		PeerID:    peerID,
		PeerName:  peerName,
		FileName:  fileName,
		Status:    DownloadPending,
		StartTime: time.Now(),
		TempDir:   filepath.Join("data", "download", ".tmp", requestID),
	}

	dm.Downloads[requestID] = download

	os.MkdirAll(download.TempDir, 0755)

	if dm.OnProgress != nil {
		dm.OnProgress(requestID, string(DownloadPending), 0, -1, "")
	}

	return download
}

func (dm *DownloadManager) HandleDownloadInfo(info protocol.DownloadInfo) error {
	download, ok := dm.Downloads[info.RequestID]
	if !ok {
		return fmt.Errorf("download not found: %s", info.RequestID)
	}

	download.mu.Lock()
	defer download.mu.Unlock()

	download.OriginalSize = info.OriginalSize
	download.CompressedSize = info.CompressedSize
	download.ChunkSize = info.ChunkSize
	download.ChunkCount = info.ChunkCount
	download.Status = DownloadActive

	download.Chunks = make([]*Chunk, info.ChunkCount)
	for i := 0; i < len(download.Chunks); i++ {
		download.Chunks[i] = &Chunk{
			Index:  i,
			Status: ChunkPending,
		}
	}
	log.Printf("Download info received for %s: %d chunks", info.FileName, info.ChunkCount)

	if dm.OnProgress != nil {
		dm.OnProgress(info.RequestID, string(DownloadActive), 0, -1, "")
	}

	return nil
}

func (dm *DownloadManager) GetDownload(requestID string) (*Download, bool) {
	download, exists := dm.Downloads[requestID]
	return download, exists
}

func (dm *DownloadManager) GetNextChunkToRequest(requestID string) (int, bool) {
	dm.mu.RLock()
	download, exists := dm.Downloads[requestID]
	dm.mu.RUnlock()
	if !exists {
		return -1, false
	}

	download.mu.Lock()
	defer download.mu.Unlock()

	for i, chunk := range download.Chunks {
		if chunk.Status == ChunkPending {
			chunk.Status = ChunkRequesting
			if dm.OnProgress != nil {
				dm.OnProgress(requestID, string(DownloadActive), download.Progress, i, string(ChunkRequesting))
			}
			return i, true
		}
	}

	return -1, false
}

func (dm *DownloadManager) SaveChunk(requestID string, index int, data []byte, sha string) error {
	download, exists := dm.Downloads[requestID]
	if !exists {
		return fmt.Errorf("download not found: %s", requestID)
	}

	if index >= len(download.Chunks) {
		return fmt.Errorf("invalid chunk index: %d", index)
	}

	download.mu.Lock()
	defer download.mu.Unlock()

	hash := sha256.Sum256(data)
	computedSHA := hex.EncodeToString(hash[:])
	if computedSHA != sha {
		download.Chunks[index].Status = ChunkError
		download.Chunks[index].RetryCount++
		log.Printf("Chunk %d validation failed for %s", index, requestID)
		return fmt.Errorf("chunk SHA mismatch")
	}

	chunkPath := filepath.Join(download.TempDir, fmt.Sprintf("chunk_%d", index))
	if err := ioutil.WriteFile(chunkPath, data, 0644); err != nil {
		return err
	}

	download.Chunks[index].Status = ChunkReceived
	download.Chunks[index].SHA = sha

	received := 0
	for _, chunk := range download.Chunks {
		if chunk.Status == ChunkReceived {
			received++
		}
	}

	download.Progress = float64(received) / float64(download.ChunkCount)

	log.Printf("Chunk %d/%d received for %s (%.1f%%)", received, download.ChunkCount, download.FileName, download.Progress*100)

	if received == download.ChunkCount {
		go dm.AssembleFile(requestID)
	}

	if dm.OnProgress != nil {
		dm.OnProgress(requestID, string(DownloadActive), download.Progress, index, string(ChunkReceived))
	}

	return nil
}

func (dm *DownloadManager) AssembleFile(requestID string) error {
	download, exists := dm.Downloads[requestID]
	if !exists {
		return fmt.Errorf("download not found: %s", requestID)
	}

	download.mu.Lock()
	defer download.mu.Unlock()

	log.Printf("Assembling file: %s", download.FileName)

	compressedPath := filepath.Join(download.TempDir, "compressed.gz")
	outFile, err := os.Create(compressedPath)
	if err != nil {
		download.Status = DownloadError
		return err
	}

	for i := 0; i < download.ChunkCount; i++ {
		chunkPath := filepath.Join(download.TempDir, fmt.Sprintf("chunk_%d", i))
		chunkData, err := ioutil.ReadFile(chunkPath)
		if err != nil {
			outFile.Close()
			download.Status = DownloadError
			return err
		}
		outFile.Write(chunkData)
	}
	outFile.Close()

	finalPath := filepath.Join("data", "download", download.FileName)
	if err := dm.decompressFile(compressedPath, finalPath); err != nil {
		download.Status = DownloadError
		os.RemoveAll(download.TempDir) // Cleanup on decompression failure
		return err
	}
	os.RemoveAll(download.TempDir)

	download.Status = DownloadComplete
	download.Progress = 1.0

	log.Printf("Download complete: %s -> %s", download.FileName, finalPath)

	if dm.OnProgress != nil {
		dm.OnProgress(requestID, string(DownloadComplete), 1.0, -1, "")
	}

	return nil
}

func (dm *DownloadManager) decompressFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	gzReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, gzReader)
	return err
}

func (dm *DownloadManager) GetAllDownloads() []*Download {
	downloads := make([]*Download, 0, len(dm.Downloads))
	for _, d := range dm.Downloads {
		downloads = append(downloads, d)
	}
	return downloads
}

func (dm *DownloadManager) ClearCompleted() {
	for id, download := range dm.Downloads {
		if download.Status == DownloadComplete || download.Status == DownloadError {
			os.RemoveAll(download.TempDir)
			delete(dm.Downloads, id)
		}
	}
}
