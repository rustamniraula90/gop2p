package main

import (
	"flag"
	"log"
	"net"
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
	handler := NewHandler(conn, registry)

	buf := make([]byte, 4096)
	for {
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("failed to read from UDP: %v", err)
			continue
		}
		go handler.HandlePacket(remote, buf[:n])
	}
}
