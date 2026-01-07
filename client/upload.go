package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rustamniraula90/gop2p/protocol"
)

type UploadStatus string

const (
	UploadPreparing UploadStatus = "preparing"
	UploadActive    UploadStatus = "active"
	UploadComplete  UploadStatus = "complete"
	UploadError     UploadStatus = "error"
)

type Upload struct {
	RequestID      string       `json:"RequestID"`
	PeerID         string       `json:"PeerID"`
	PeerName       string       `json:"PeerName"`
	FileName       string       `json:"FileName"`
	OriginalSize   int64        `json:"OriginalSize"`
	CompressedFile string       `json:"compressed_file"`
	CompressedSize int64        `json:"compressed_size"`
	Status         UploadStatus `json:"status"`
	ChunkSize      int          `json:"chunk_size"`
	ChunkCount     int          `json:"chunk_count"`
	ChunksSent     int          `json:"chunks_sent"`
	StartTime      time.Time    `json:"start_time"`
	TmpDir         string       `json:"tmp_dir"`
	Progress       float64      `json:"progress"` // 0.0 to 1.0
	ChunkSHAs      []string
	mu             sync.RWMutex
}

type UploadManager struct {
	uploads    map[string]*Upload
	mu         sync.RWMutex
	OnProgress func(requestID, status string, progress float64, chunkIndex int, chunkStatus string)
}

func NewUploadManager() *UploadManager {
	um := &UploadManager{
		uploads: make(map[string]*Upload),
	}

	// Auto-cleanup finished uploads every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			um.ClearCompleted()
		}
	}()

	return um
}

func (um *UploadManager) PrepareFile(requestID, peerID, peerName, fileName string) (*protocol.DownloadInfo, error) {
	um.mu.Lock()
	filePath := filepath.Join("data", "share", fileName)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", fileName)
	}

	originalSize := fileInfo.Size()

	upload := &Upload{
		RequestID:    requestID,
		PeerID:       peerID,
		PeerName:     peerName,
		FileName:     fileName,
		OriginalSize: originalSize,
		Status:       UploadPreparing,
		ChunkSize:    DefaultChunkSize,
		StartTime:    time.Now(),
		TmpDir:       filepath.Join("data", "share", ".tmp", requestID),
		Progress:     0.0,
	}

	if err := os.MkdirAll(upload.TmpDir, 0755); err != nil {
		um.mu.Unlock()
		return nil, err
	}

	compressPath := filepath.Join(upload.TmpDir, "compressed.gz")
	if err := um.compressFile(filePath, compressPath); err != nil {
		um.mu.Unlock()
		return nil, err
	}

	compressInfo, err := os.Stat(compressPath)
	if err != nil {
		um.mu.Unlock()
		os.Remove(compressPath)
		return nil, err
	}
	upload.CompressedSize = compressInfo.Size()
	upload.CompressedFile = compressPath

	chunkCount := int(upload.CompressedSize) / upload.ChunkSize
	if int(upload.CompressedSize)%upload.ChunkSize != 0 {
		chunkCount++
	}
	upload.ChunkCount = chunkCount

	upload.ChunkSHAs = make([]string, upload.ChunkCount)
	f, err := os.Open(compressPath)
	if err != nil {
		um.mu.Unlock()
		os.Remove(compressPath)
		return nil, err
	}
	defer f.Close()

	for i := 0; i < chunkCount; i++ {
		buf := make([]byte, upload.ChunkSize)
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			um.mu.Unlock()
			os.Remove(compressPath)
			return nil, err
		}
		chunk := buf[:n]
		hash := sha256.Sum256(chunk)
		upload.ChunkSHAs[i] = hex.EncodeToString(hash[:])
	}

	upload.Status = UploadActive
	um.uploads[requestID] = upload
	um.mu.Unlock()

	log.Printf("File prepared for upload: %s (%d bytes -> %d bytes, %d chunks)",
		fileName, originalSize, upload.CompressedSize, chunkCount)

	if um.OnProgress != nil {
		um.OnProgress(requestID, string(UploadActive), 0, -1, "")
	}

	return &protocol.DownloadInfo{
		RequestID:      requestID,
		FileName:       fileName,
		OriginalSize:   originalSize,
		CompressedSize: upload.CompressedSize,
		ChunkSize:      upload.ChunkSize,
		ChunkCount:     upload.ChunkCount,
	}, nil
}

func (um *UploadManager) compressFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	gzipWriter := gzip.NewWriter(dstFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, srcFile)
	return err
}

func (um *UploadManager) GetChunk(requestID string, index int) (*protocol.ChunkResponse, error) {
	upload, exists := um.uploads[requestID]
	if !exists {
		return nil, fmt.Errorf("upload not found: %s", requestID)
	}

	if index < 0 || index >= upload.ChunkCount {
		return nil, fmt.Errorf("invalid chunk index: %d", index)
	}

	upload.mu.Lock()
	defer upload.mu.Unlock()

	f, err := os.Open(upload.CompressedFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	offset := int64(index * upload.ChunkSize)
	buf := make([]byte, upload.ChunkSize)
	n, err := f.ReadAt(buf, offset)
	if err != nil && err != io.EOF {
		return nil, err
	}
	data := buf[:n]

	upload.ChunksSent++
	upload.Progress = float64(upload.ChunksSent) / float64(upload.ChunkCount)

	log.Printf("Sending chunk %d/%d for %s (%.1f%%)",
		upload.ChunksSent, upload.ChunkCount, upload.FileName, upload.Progress*100)

	if um.OnProgress != nil {
		um.OnProgress(requestID, string(UploadActive), upload.Progress, index, "sent")
	}

	if upload.ChunksSent >= upload.ChunkCount {
		upload.Status = UploadComplete
		if um.OnProgress != nil {
			um.OnProgress(requestID, string(UploadComplete), 1.0, -1, "")
		}
	}

	return &protocol.ChunkResponse{
		RequestID:  requestID,
		ChunkIndex: index,
		ChunkData:  data,
		ChunkSHA:   upload.ChunkSHAs[index],
	}, nil
}

func (um *UploadManager) ClearCompleted() {
	for id, upload := range um.uploads {
		if upload.Status == UploadComplete || upload.Status == UploadError {
			os.RemoveAll(upload.TmpDir)
			delete(um.uploads, id)
		}
	}
}

func (um *UploadManager) GetUpload(requestID string) (*Upload, bool) {
	upload, exists := um.uploads[requestID]
	return upload, exists
}

func (um *UploadManager) GetAllUploads() []*Upload {
	uploads := make([]*Upload, 0, len(um.uploads))
	for _, u := range um.uploads {
		uploads = append(uploads, u)
	}
	return uploads
}
