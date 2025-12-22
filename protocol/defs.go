package protocol

import "net"

type PacketType byte

const (
	TypeRegister    PacketType = 0x01
	TypeRegisterAck PacketType = 0x02
	TypeHeartbeat   PacketType = 0x03

	TypeListPeersRequest  PacketType = 0x04
	TypeListPeersResponse PacketType = 0x04

	TypeConnectRequest PacketType = 0x05
	TypeConnectAccept  PacketType = 0x06
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
