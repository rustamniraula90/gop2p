package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/rustamniraula90/gop2p/client/db"
	"github.com/rustamniraula90/gop2p/protocol"
)

type PeerState int

const (
	StateDisconnected PeerState = iota
	StatePunching
	StateConnected
)

type Peer struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	State    PeerState `json:"state"`
	IP       net.IP    `json:"ip"`
	Port     int       `json:"port"`
	LastUsed time.Time `json:"last_used"`
}

type P2PManager struct {
	identity   *Identity
	udp        *UDPManager
	ServerAddr *net.UDPAddr
	Store      *db.Store
	Peers      map[string]*Peer

	OnMessage           func(senderID, text string)
	OnPeerUpdate        func(peer *Peer)
	OnConnectionRequest func(requesterID, name string)
}

func NewP2PManager(identity *Identity, udpManager *UDPManager, store *db.Store, server string) (*P2PManager, error) {
	sAddr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		return nil, err
	}
	return &P2PManager{
		identity:   identity,
		udp:        udpManager,
		ServerAddr: sAddr,
		Store:      store,
		Peers:      make(map[string]*Peer),
	}, nil
}

func (pm *P2PManager) Start() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			pm.sendHeartbeat()
		}
	}()

	go func() {
		for packet := range pm.udp.Incoming {
			pm.handlePacket(packet)
		}
	}()
	pm.sendRegister()
}

func (pm *P2PManager) sendRegister() {
	log.Printf("Sending register to %s", pm.ServerAddr)
	payload := protocol.RegisterPayload{ID: pm.identity.ID, Name: pm.identity.Name}
	err := pm.udp.Send(pm.ServerAddr, protocol.UDPMessage{
		Type:    protocol.TypeRegister,
		Payload: payload,
	})
	if err != nil {
		log.Fatalln("Failed to register client to relay server", err)
	}
}

func (pm *P2PManager) handlePacket(packet PacketWrapper) {
	if packet.RemoteAddr.String() == pm.ServerAddr.String() {
		pm.handleServerMessage(packet.Message)
		return
	}
	pm.handlePeerMessage(packet.RemoteAddr, packet.Message)
}

func (pm *P2PManager) handleServerMessage(msg protocol.UDPMessage) {
	switch msg.Type {
	case protocol.TypeRegisterAck:
		log.Println("Registered to relay server successfully")
	case protocol.TypeListPeersResponse:
		pm.handlePeersList(msg)
	case protocol.TypeConnectForward:
		pm.handleConnectForward(msg)
	case protocol.TypePeerInfo:
		pm.handlePeerInfo(msg)
	}
}

func (pm *P2PManager) handlePeerMessage(remote *net.UDPAddr, msg protocol.UDPMessage) {
	switch msg.Type {
	case protocol.TypePunch, protocol.TypePunchAck:
		pm.handlePunch(remote, msg)
	case protocol.TypeChat:
		pm.handleChatMessage(msg)

	}
}

func (pm *P2PManager) handlePunch(remote *net.UDPAddr, msg protocol.UDPMessage) {
	bytes, _ := json.Marshal(msg.Payload)
	var p map[string]string
	json.Unmarshal(bytes, &p)
	senderID := p["from"]
	if senderID == "" {
		return
	}
	peer, ok := pm.Peers[senderID]
	if !ok {
		return
	}
	peer.LastUsed = time.Now()
	peer.State = StateConnected
	peer.IP = remote.IP
	peer.Port = remote.Port

	if msg.Type == protocol.TypePunchAck {
		log.Printf("Received PUNCH_ACK from %s. Connected!", senderID)
		return
	} else {
		log.Printf("Received PUNCH from %s. Replying with ACK.", senderID)
		err := pm.udp.Send(remote, protocol.UDPMessage{
			Type: protocol.TypePunchAck,
			Payload: map[string]string{
				"from": pm.identity.ID,
			},
		})
		if err != nil {
			log.Printf("Failed to send PUNCH_ACK to %s. Error: %v", senderID, err)
		}
	}

	if pm.OnPeerUpdate != nil {
		pm.OnPeerUpdate(peer)
	}
}

func (pm *P2PManager) sendHeartbeat() {
	payload := protocol.RegisterPayload{ID: pm.identity.ID, Name: pm.identity.Name}
	err := pm.udp.Send(pm.ServerAddr, protocol.UDPMessage{
		Type:    protocol.TypeHeartbeat,
		Payload: payload,
	})
	if err != nil {
		log.Fatalln("Failed to send heartbeat to relay server", err)
	}
}

func (pm *P2PManager) FetchPeers() {
	err := pm.udp.Send(pm.ServerAddr, protocol.UDPMessage{
		Type:    protocol.TypeListPeersRequest,
		Payload: nil,
	})
	if err != nil {
		log.Fatalln("Failed to request peers from relay server", err)
	}
}

func (pm *P2PManager) handlePeersList(msg protocol.UDPMessage) {
	bytes, _ := json.Marshal(msg.Payload)

	var resp protocol.ListPeersPayload
	json.Unmarshal(bytes, &resp)

	for _, p := range resp.Peers {
		if p.ID == pm.identity.ID {
			continue
		}

		existing, ok := pm.Peers[p.ID]
		if ok && existing.State == StateConnected {
			continue
		}

		peer := &Peer{
			ID:    p.ID,
			Name:  p.Name,
			State: StateDisconnected,
		}

		if pm.OnPeerUpdate != nil {
			pm.OnPeerUpdate(peer)
		}

	}
}

func (pm *P2PManager) RequestConnection(targetID string) {
	req := protocol.ConnectRequest{
		TargetID: targetID,
		SenderID: pm.identity.ID,
	}

	if err := pm.udp.Send(pm.ServerAddr, protocol.UDPMessage{
		Type:    protocol.TypeConnectRequest,
		Payload: &req,
	}); err != nil {
		log.Fatalln("Failed to send connect request", err)
	}
}

func (pm *P2PManager) AcceptConnection(requesterID string) {
	req := protocol.ConnectAccept{
		RequesterID: requesterID,
		SenderID:    pm.identity.ID,
	}
	if err := pm.udp.Send(pm.ServerAddr, protocol.UDPMessage{
		Type:    protocol.TypeConnectAccept,
		Payload: &req,
	}); err != nil {
		log.Fatalln("Failed to accept connect request", err)
	}
}

func (pm *P2PManager) SetServer(addr string) error {
	rAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	pm.ServerAddr = rAddr
	// Re-register
	pm.sendRegister()
	return nil
}

func (pm *P2PManager) RemovePeer(id string) {
	delete(pm.Peers, id)
}

func (pm *P2PManager) handlePeerInfo(msg protocol.UDPMessage) {
	bytes, _ := json.Marshal(msg.Payload)
	var p protocol.PeerInfoPayload
	if err := json.Unmarshal(bytes, &p); err != nil {
		log.Println("Failed to unmarshal peer info", err)
		return
	}

	log.Printf("Received Peer Info: %s (%s) at %s:%d.", p.Name, p.ID, p.IP, p.Port)

	peer := &Peer{
		ID:    p.ID,
		Name:  p.Name,
		IP:    p.IP,
		Port:  p.Port,
		State: StatePunching,
	}
	pm.Peers[peer.ID] = peer
	pm.SavePeers()

	if pm.OnPeerUpdate != nil {
		pm.OnPeerUpdate(peer)
	}

	go pm.startPunching(peer)
}

func (pm *P2PManager) startPunching(peer *Peer) {
	addr := &net.UDPAddr{IP: peer.IP, Port: peer.Port}
	// send packet and pray
	for i := 0; i < 10; i++ {
		if pm.Peers[peer.ID].State == StateConnected {
			return
		}

		log.Printf("Punching peer %s(%s) at %s", peer.Name, peer.ID, addr)
		pm.udp.Send(addr, protocol.UDPMessage{
			Type: protocol.TypePunch,
			Payload: map[string]string{
				"from": pm.identity.ID,
			},
		})
		time.Sleep(500 * time.Millisecond)
	}
}

func (pm *P2PManager) handleConnectForward(msg protocol.UDPMessage) {
	bytes, _ := json.Marshal(msg.Payload)
	var req protocol.ConnectForward
	if err := json.Unmarshal(bytes, &req); err != nil {
		log.Println("Failed to unmarshal connect forward message", err)
		return
	}
	log.Printf("Received connection request from %s (%s)", req.RequesterName, req.RequesterID)

	if pm.OnConnectionRequest != nil {
		pm.OnConnectionRequest(req.RequesterID, req.RequesterName)
	}
}

func (pm *P2PManager) SavePeers() {
	for _, peer := range pm.Peers {
		np := db.PeerEntity{
			ID:       peer.ID,
			Name:     peer.Name,
			IP:       peer.IP,
			Port:     peer.Port,
			State:    int(peer.State),
			LastUsed: peer.LastUsed,
		}
		if err := pm.Store.SavePeer(np); err != nil {
			log.Println("Failed to save peer", "id", peer.ID, err)
		}
	}
}

func (pm *P2PManager) handleChatMessage(msg protocol.UDPMessage) {
	bytes, _ := json.Marshal(msg.Payload)
	var p protocol.ChatMessage
	json.Unmarshal(bytes, &p)

	if p.SenderID == "" {
		return
	}
	peer, ok := pm.Peers[p.SenderID]
	if !ok {
		return
	}
	log.Printf("Chat from %s: %s", peer.Name, p.Text)

	if err := pm.Store.SaveMessage(peer.ID, peer.ID, p.Text, p.Timestamp); err != nil {
		log.Printf("Error saving message: %v", err)
	}

	if pm.OnMessage != nil {
		pm.OnMessage(peer.ID, p.Text)
	}
}

func (pm *P2PManager) SendMessage(targetID string, text string) {
	peer, ok := pm.Peers[targetID]
	if !ok {
		log.Printf("Cannot send to %s: Peer not found", targetID)
		return
	}

	if peer.State != StateConnected {
		log.Printf("Cannot send to %s: Peer state is %v (not Connected)", targetID, peer.State)
		return
	}

	addr := &net.UDPAddr{IP: peer.IP, Port: peer.Port}
	ts := time.Now().Unix()
	msg := protocol.ChatMessage{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()), // simple ID
		Timestamp: ts,
		SenderID:  pm.identity.ID,
		Text:      text,
	}

	err := pm.udp.Send(addr, protocol.UDPMessage{
		Type:    protocol.TypeChat,
		Payload: msg,
	})
	if err != nil {
		log.Println("Failed to send chat message", err)
		return
	}

	// Save my own message
	if err := pm.Store.SaveMessage(targetID, "me", text, ts); err != nil {
		log.Printf("Error saving sent message: %v", err)
	}

}
