package main

import (
	"encoding/json"
	"log"
	"net"

	"github.com/rustamniraula90/gop2p/protocol"
)

type UDPManager struct {
	conn     *net.UDPConn
	Incoming chan PacketWrapper
}

type PacketWrapper struct {
	RemoteAddr *net.UDPAddr
	Message    protocol.UDPMessage
}

func NewUDPManager(port int) (*UDPManager, error) {
	addr := &net.UDPAddr{Port: port}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	um := &UDPManager{
		conn:     conn,
		Incoming: make(chan PacketWrapper, 100),
	}

	go um.readLoop()
	return um, nil
}

func (um *UDPManager) readLoop() {
	buf := make([]byte, 65535)
	for {
		n, remote, err := um.conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		var msg protocol.UDPMessage
		if err := json.Unmarshal(buf[:n], &msg); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		um.Incoming <- PacketWrapper{
			RemoteAddr: remote,
			Message:    msg,
		}
	}
}

func (um *UDPManager) Send(addr *net.UDPAddr, msg protocol.UDPMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = um.conn.WriteToUDP(data, addr)
	return err
}
