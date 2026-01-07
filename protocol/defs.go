package protocol

import "net"

type PacketType byte

const (
	TypeRegister    PacketType = 0x01
	TypeRegisterAck PacketType = 0x02
	TypeHeartbeat   PacketType = 0x03

	TypeListPeersRequest  PacketType = 0x04
	TypeListPeersResponse PacketType = 0x04
	TypeConnectRequest    PacketType = 0x05
	TypeConnectAccept     PacketType = 0x06
	TypeConnectForward    PacketType = 0x07
	TypePeerInfo          PacketType = 0x08

	TypePunch    PacketType = 0x09
	TypePunchAck PacketType = 0x10

	TypeChat             PacketType = 0x11
	TypeListFileRequest  PacketType = 0x12
	TypeListFileResponse PacketType = 0x13
	TypeDownloadRequest  PacketType = 0x14
	TypeDownloadInfo     PacketType = 0x15
	TypeChunkRequest     PacketType = 0x16
	TypeChunkResponse    PacketType = 0x17
)

type UDPMessage struct {
	Type    PacketType  `json:"type"`
	Payload interface{} `json:"payload"`
}

type RegisterPayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PeerInfoPayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	IP   net.IP `json:"ip"`
	Port int    `json:"port"`
}

type ListPeersPayload struct {
	Peers []PeerInfoPayload `json:"peers"`
}

type ConnectRequest struct {
	TargetID string `json:"target_id"`
	SenderID string `json:"sender_id"`
}

type ConnectAccept struct {
	RequesterID string `json:"requester_id"`
	SenderID    string `json:"sender_id"`
}

type ListPeersResponse struct {
	Peers []PeerInfoPayload `json:"peers"`
}

type ConnectForward struct {
	RequesterID   string `json:"requester_id"`
	RequesterName string `json:"requester_name"`
}

type ChatMessage struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	SenderID  string `json:"sender_id"`
	Text      string `json:"text"`
}

type FileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"type"` // "file" or "dir"
}

type FileListPayload struct {
	Files []FileInfo `json:"files"`
}

type DownloadRequest struct {
	FileName  string `json:"file_name"`
	RequestID string `json:"request_id"` // Unique ID to track this download
}

type DownloadInfo struct {
	RequestID      string `json:"request_id"`
	FileName       string `json:"file_name"`
	OriginalSize   int64  `json:"original_size"`
	CompressedSize int64  `json:"compressed_size"`
	ChunkCount     int    `json:"chunk_count"`
	ChunkSize      int    `json:"chunk_size"`
}

type ChunkRequest struct {
	RequestID  string `json:"request_id"`
	ChunkIndex int    `json:"chunk_index"`
}

type ChunkResponse struct {
	RequestID  string `json:"request_id"`
	ChunkIndex int    `json:"chunk_index"`
	ChunkData  []byte `json:"chunk_data"`
	ChunkSHA   string `json:"chunk_sha"` // SHA256 hex string
}
