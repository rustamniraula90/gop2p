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
