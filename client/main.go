package main

import (
	"flag"
	"log"
)

func main() {
	var (
		serverAddr = flag.String("server", "localhost:8080", "Relay server address")
		udpPort    = flag.Int("udp", 0, "UDP port to listen on")
		apiPort    = flag.Int("api", 8081, "API port to listen on")
		name       = flag.String("name", "Test", "Display name of the client")
	)

	flag.Parse()

	identity := LoadIdentity(*name)
	log.Printf("Client Identity: %s (%s)", identity.Name, identity.ID)

	um, err := NewUDPManager(*udpPort)
	if err != nil {
		log.Fatal("Failed to create UDP manager: ", err)
	}

	p2pm, err := NewP2PManager(identity, um, *serverAddr)
	if err != nil {
		log.Fatal("Failed to create P2P manager: ", err)
	}
	p2pm.Start()

	api := NewAPIServer(p2pm, *apiPort)
	api.Start()

}
