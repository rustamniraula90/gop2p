package main

import (
	"encoding/json"
	"log"
	"net"

	"github.com/rustamniraula90/gop2p/protocol"
)

func handleRegister(conn *net.UDPConn, registry *Registry, remoteAddr *net.UDPAddr, msg protocol.UDPMessage) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var p protocol.RegisterPayload
	if err := json.Unmarshal(payloadBytes, &p); err != nil {
		log.Printf("failed to unmarshal RegisterPayload: %v", err)
		return
	}

	registry.Register(p.ID, p.Name, remoteAddr)

	if msg.Type == protocol.TypeRegister {
		log.Printf("Registered client: %s (%s) at %s", p.Name, p.ID, remoteAddr)
		sendJSON(conn, remoteAddr, protocol.UDPMessage{
			Type:    protocol.TypeRegisterAck,
			Payload: map[string]interface{}{"status": "ok"},
		})
	} else {
		log.Printf("Received heartbeat: %s (%s) at %s", p.Name, p.ID, remoteAddr)
	}
}

func handlePeerRequest(conn *net.UDPConn, registry *Registry, remoteAddr *net.UDPAddr) {
	log.Printf("Handling peer request from %s", remoteAddr)
	clients := registry.List()
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
	sendJSON(conn, remoteAddr, protocol.UDPMessage{
		Type:    protocol.TypeListPeersResponse,
		Payload: resp,
	})
}

func handleConnectAccept(conn *net.UDPConn, registry *Registry, addr *net.UDPAddr, msg protocol.UDPMessage) {
	bytes, _ := json.Marshal(msg.Payload)
	var acc protocol.ConnectAccept
	if err := json.Unmarshal(bytes, &acc); err != nil {
		log.Printf("failed to unmarshal ConnectAccept: %v", err)
		return
	}

	requester, ok := registry.Get(acc.RequesterID)
	if !ok {
		return
	}
	accepter, ok := registry.Get(acc.SenderID)
	if !ok {
		return
	}

	requesterAddr := &net.UDPAddr{IP: requester.IP, Port: requester.Port}
	sendJSON(conn, requesterAddr, protocol.UDPMessage{
		Type: protocol.TypePeerInfo,
		Payload: protocol.PeerInfoPayload{
			ID:   accepter.ID,
			Name: accepter.Name,
			IP:   accepter.IP,
			Port: accepter.Port,
		},
	})

	accepterAddr := &net.UDPAddr{IP: accepter.IP, Port: accepter.Port}
	sendJSON(conn, accepterAddr, protocol.UDPMessage{
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

func handleConnectRequest(conn *net.UDPConn, registry *Registry, addr *net.UDPAddr, msg protocol.UDPMessage) {
	bytes, _ := json.Marshal(msg.Payload)
	var req protocol.ConnectRequest
	if err := json.Unmarshal(bytes, &req); err != nil {
		log.Printf("failed to unmarshal ConnectRequest: %v", err)
		return
	}
	target, ok := registry.Get(req.TargetID)
	if !ok {
		return
	}
	requester, ok := registry.Get(req.SenderID)
	if !ok {
		return
	}

	fwd := protocol.ConnectForward{
		RequesterID:   requester.ID,
		RequesterName: requester.Name,
	}

	targetAddr := &net.UDPAddr{IP: target.IP, Port: target.Port}

	sendJSON(conn, targetAddr, protocol.UDPMessage{
		Type:    protocol.TypeConnectForward,
		Payload: fwd,
	})
}
