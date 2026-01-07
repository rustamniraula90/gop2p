package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/rustamniraula90/gop2p/client/db"
)

func main() {
	var (
		serverAddr = flag.String("server", "127.0.0.1:8080", "Relay server address")
		udpPort    = flag.Int("udp", 0, "UDP port to listen on")
		apiPort    = flag.Int("api", 8081, "API port to listen on")
		name       = flag.String("name", "Test", "Display name of the client")
	)

	flag.Parse()

	os.MkdirAll("data", 0755)
	os.MkdirAll(filepath.Join("data", "download", ".tmp"), 0755)
	os.MkdirAll(filepath.Join("data", "share", ".tmp"), 0755)

	store, err := db.NewStore("data/.gop2p.db")
	if err != nil {
		log.Fatal("Failed to init database", err)
	}

	identity, err := LoadIdentity(store, *name)
	if err != nil {
		log.Fatal("Failed to load identity", err)
	}
	log.Printf("Client Identity: %s (%s)", identity.Name, identity.ID)

	um, err := NewUDPManager(*udpPort)
	if err != nil {
		log.Fatal("Failed to create UDP manager: ", err)
	}
	log.Printf("Listening on UDP %s", um.conn.LocalAddr())

	p2pm, err := NewP2PManager(identity, um, store, *serverAddr)
	if err != nil {
		log.Fatal("Failed to create P2P manager: ", err)
	}
	p2pm.Start()

	if err = p2pm.Store.SaveServerConfig(*serverAddr); err != nil {
		log.Fatal("Failed to save server config", err)
	}

	api := NewAPIServer(p2pm, *apiPort)
	api.Start()

}
