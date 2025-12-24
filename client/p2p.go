package main

import (
	"encoding/json"
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
	upd        *UDPManager
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
		upd:        udpManager,
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
		for packet := range pm.upd.Incoming {
			pm.handlePacket(packet)
		}
	}()
	pm.sendRegister()
}

func (pm *P2PManager) sendRegister() {
	log.Printf("Sending register to %s", pm.ServerAddr)
	payload := protocol.RegisterPayload{ID: pm.identity.ID, Name: pm.identity.Name}
	err := pm.upd.Send(pm.ServerAddr, protocol.UDPMessage{
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
	pm.handlePeerMessage(packet.Message)
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
		pm.handlePeerInfo(msg, true)
	}
}

func (pm *P2PManager) handlePeerMessage(msg protocol.UDPMessage) {

}

func (pm *P2PManager) sendHeartbeat() {
	log.Printf("Sending heartbeat to %s", pm.ServerAddr)
	payload := protocol.RegisterPayload{ID: pm.identity.ID, Name: pm.identity.Name}
	err := pm.upd.Send(pm.ServerAddr, protocol.UDPMessage{
		Type:    protocol.TypeHeartbeat,
		Payload: payload,
	})
	if err != nil {
		log.Fatalln("Failed to send heartbeat to relay server", err)
	}
}

func (pm *P2PManager) FetchPeers() {
	err := pm.upd.Send(pm.ServerAddr, protocol.UDPMessage{
		Type:    protocol.TypeListPeersRequest,
		Payload: nil,
	})
	if err != nil {
		log.Fatalln("Failed to request peers from relay server", err)
	}
}

func (pm *P2PManager) handlePeersList(msg protocol.UDPMessage) {
	bytes, _ := json.Marshal(msg.Payload)
	log.Println("Received list of peers from server", string(bytes))

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

	if err := pm.upd.Send(pm.ServerAddr, protocol.UDPMessage{
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
	if err := pm.upd.Send(pm.ServerAddr, protocol.UDPMessage{
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

func (pm *P2PManager) handlePeerInfo(msg protocol.UDPMessage, isTarget bool) {
	bytes, _ := json.Marshal(msg.Payload)
	var p protocol.PeerInfoPayload
	if err := json.Unmarshal(bytes, &p); err != nil {
		log.Println("Failed to unmarshal peer info", err)
		return
	}

	log.Printf("Received Peer Info: %s (%s) at %s:%d. IsTarget=%v", p.Name, p.ID, p.IP, p.Port, isTarget)

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
			LastUsed: peer.LastUsed,
		}
		if err := pm.Store.SavePeer(np); err != nil {
			log.Println("Failed to save peer", "id", peer.ID, err)
		}
	}
}
