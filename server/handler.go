package main

import (
	"encoding/json"
	"log"
	"net"

	"github.com/rustamniraula90/gop2p/protocol"
)

type IHandler interface {
	HandlePacket(remoteAddr *net.UDPAddr, buf []byte)
	handleRegister(remoteAddr *net.UDPAddr, msg protocol.UDPMessage)
	handlePeerRequest(remoteAddr *net.UDPAddr)
	handleConnectRequest(msg protocol.UDPMessage)
	handleConnectAccept(msg protocol.UDPMessage)
	sendJSON(remoteAddr *net.UDPAddr, message protocol.UDPMessage)
}

type Handler struct {
	conn     *net.UDPConn
	registry *Registry
}

func NewHandler(conn *net.UDPConn, registry *Registry) IHandler {
	return &Handler{
		conn:     conn,
		registry: registry,
	}
}

func (h *Handler) HandlePacket(remoteAddr *net.UDPAddr, buf []byte) {
	var msg protocol.UDPMessage
	if err := json.Unmarshal(buf, &msg); err != nil {
		log.Printf("failed to unmarshal UDP message: %v", err)
		return
	}
	switch msg.Type {
	case protocol.TypeRegister, protocol.TypeHeartbeat:
		h.handleRegister(remoteAddr, msg)
	case protocol.TypeListPeersRequest:
		h.handlePeerRequest(remoteAddr)
	case protocol.TypeConnectRequest:
		h.handleConnectRequest(msg)
	case protocol.TypeConnectAccept:
		h.handleConnectAccept(msg)
	}
}

func (h *Handler) handleRegister(remoteAddr *net.UDPAddr, msg protocol.UDPMessage) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var p protocol.RegisterPayload
	if err := json.Unmarshal(payloadBytes, &p); err != nil {
		log.Printf("failed to unmarshal RegisterPayload: %v", err)
		return
	}

	h.registry.Register(p.ID, p.Name, remoteAddr)

	if msg.Type == protocol.TypeRegister {
		log.Printf("Registered client: %s (%s) at %s", p.Name, p.ID, remoteAddr)
		h.sendJSON(remoteAddr, protocol.UDPMessage{
			Type:    protocol.TypeRegisterAck,
			Payload: map[string]interface{}{"status": "ok"},
		})
	}
}

func (h *Handler) handlePeerRequest(remoteAddr *net.UDPAddr) {
	log.Printf("Handling peer request from %s", remoteAddr)
	clients := h.registry.List()
	peers := make([]protocol.PeerInfoPayload, 0, len(clients))
	for _, c := range clients {
		peers = append(peers, protocol.PeerInfoPayload{
			ID:   c.ID,
			Name: c.Name,
			IP:   nil,
			Port: 0,
		})
	}
	resp := protocol.ListPeersResponse{
		Peers: peers,
	}
	h.sendJSON(remoteAddr, protocol.UDPMessage{
		Type:    protocol.TypeListPeersResponse,
		Payload: resp,
	})
}

func (h *Handler) handleConnectRequest(msg protocol.UDPMessage) {
	bytes, _ := json.Marshal(msg.Payload)
	var req protocol.ConnectRequest
	if err := json.Unmarshal(bytes, &req); err != nil {
		log.Printf("failed to unmarshal ConnectRequest: %v", err)
		return
	}
	target, ok := h.registry.Get(req.TargetID)
	if !ok {
		return
	}
	requester, ok := h.registry.Get(req.SenderID)
	if !ok {
		return
	}

	fwd := protocol.ConnectForward{
		RequesterID:   requester.ID,
		RequesterName: requester.Name,
	}

	targetAddr := &net.UDPAddr{IP: target.IP, Port: target.Port}

	h.sendJSON(targetAddr, protocol.UDPMessage{
		Type:    protocol.TypeConnectForward,
		Payload: fwd,
	})
}

func (h *Handler) handleConnectAccept(msg protocol.UDPMessage) {
	bytes, _ := json.Marshal(msg.Payload)
	var acc protocol.ConnectAccept
	if err := json.Unmarshal(bytes, &acc); err != nil {
		log.Printf("failed to unmarshal ConnectAccept: %v", err)
		return
	}

	requester, ok := h.registry.Get(acc.RequesterID)
	if !ok {
		return
	}
	accepter, ok := h.registry.Get(acc.SenderID)
	if !ok {
		return
	}

	requesterAddr := &net.UDPAddr{IP: requester.IP, Port: requester.Port}
	h.sendJSON(requesterAddr, protocol.UDPMessage{
		Type: protocol.TypePeerInfo,
		Payload: protocol.PeerInfoPayload{
			ID:   accepter.ID,
			Name: accepter.Name,
			IP:   accepter.IP,
			Port: accepter.Port,
		},
	})

	accepterAddr := &net.UDPAddr{IP: accepter.IP, Port: accepter.Port}
	h.sendJSON(accepterAddr, protocol.UDPMessage{
		Type: protocol.TypePeerInfo,
		Payload: protocol.PeerInfoPayload{
			ID:   requester.ID,
			Name: requester.Name,
			IP:   requester.IP,
			Port: requester.Port,
		},
	})

	log.Printf("Handshake accepted: %s <-> %s", requester.Name, accepter.Name)
}

func (h *Handler) sendJSON(remoteAddr *net.UDPAddr, message protocol.UDPMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("failed to marshal message: %v", err)
		return
	}
	_, err = h.conn.WriteToUDP(data, remoteAddr)
	if err != nil {
		log.Printf("failed to write to UDP: %v", err)
		return
	}
}
