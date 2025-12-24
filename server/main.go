package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"

	"github.com/rustamniraula90/gop2p/protocol"
)

func main() {
	var (
		udpPort = flag.Int("udp", 8080, "Relay server port")
	)
	flag.Parse()

	addr := net.UDPAddr{Port: *udpPort}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	log.Printf("Started relay server on UDP port %d", *udpPort)
	registry := NewRegistry()

	buf := make([]byte, 4096)
	for {
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("failed to read from UDP: %v", err)
			continue
		}

		go handlePacket(conn, remote, registry, buf[:n])
	}
}

func handlePacket(conn *net.UDPConn, remoteAddr *net.UDPAddr, registry *Registry, buf []byte) {
	var msg protocol.UDPMessage
	if err := json.Unmarshal(buf, &msg); err != nil {
		log.Printf("failed to unmarshal UDP message: %v", err)
		return
	}

	switch msg.Type {
	case protocol.TypeRegister, protocol.TypeHeartbeat:
		handleRegister(conn, registry, remoteAddr, msg)
	case protocol.TypeListPeersRequest:
		handlePeerRequest(conn, registry, remoteAddr)
	case protocol.TypeConnectRequest:
		handleConnectRequest(conn, registry, remoteAddr, msg)
	case protocol.TypeConnectAccept:
		handleConnectAccept(conn, registry, remoteAddr, msg)
	}
}

func sendJSON(conn *net.UDPConn, remoteAddr *net.UDPAddr, message protocol.UDPMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("failed to marshal message: %v", err)
		return
	}
	_, err = conn.WriteToUDP(data, remoteAddr)
	if err != nil {
		log.Printf("failed to write to UDP: %v", err)
		return
	}
}
