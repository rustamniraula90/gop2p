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
