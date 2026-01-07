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
	Downloader *DownloadManager
	Uploader   *UploadManager

	OnMessage           func(senderID, text string)
	OnPeerUpdate        func(peer *Peer)
	OnConnectionRequest func(requesterID, name string)
	OnFileList          func(peerID string, files []protocol.FileInfo)
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
		Downloader: NewDownloadManager(),
		Uploader:   NewUploadManager(),
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
	case protocol.TypeListFileRequest:
		pm.handleFileListRequest(remote)
	case protocol.TypeListFileResponse:
		pm.handleFileListResponse(msg, remote)
	case protocol.TypeDownloadRequest:
		pm.handleDownloadRequest(msg, remote)
	case protocol.TypeDownloadInfo:
		pm.handleDownloadInfo(msg, remote)
	case protocol.TypeChunkRequest:
		pm.handleChunkRequest(msg, remote)
	case protocol.TypeChunkResponse:
		pm.handleChunkResponse(msg)
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

	if err := pm.Store.SaveMessage(targetID, "me", text, ts); err != nil {
		log.Printf("Error saving sent message: %v", err)
	}

}

func (pm *P2PManager) RequestFileList(targetID string) {
	peer, ok := pm.Peers[targetID]
	if !ok || peer.State != StateConnected {
		log.Printf("Cannot request files from %s: Peer not found", targetID)
		return
	}

	addr := &net.UDPAddr{IP: peer.IP, Port: peer.Port}
	pm.udp.Send(addr, protocol.UDPMessage{
		Type:    protocol.TypeListFileRequest,
		Payload: nil,
	})
}

func (pm *P2PManager) handleFileListRequest(remote *net.UDPAddr) {
	peer, ok := pm.findPeerByRemote(remote)
	if !ok {
		log.Printf("Cannot handle file list request from %s. Peer not found!", remote)
		return
	}
	log.Printf("Handling file list request from %s", peer.Name)
	files, err := ScanShareDir()
	if err != nil {
		log.Println("Failed to scan share directory", err)
		return
	}

	pm.udp.Send(remote, protocol.UDPMessage{
		Type:    protocol.TypeListFileResponse,
		Payload: protocol.FileListPayload{Files: files},
	})

}

func (pm *P2PManager) findPeerByRemote(remote *net.UDPAddr) (*Peer, bool) {
	for _, p := range pm.Peers {
		if p.IP.Equal(remote.IP) && p.Port == remote.Port {
			return p, true
		}
	}
	return nil, false
}

func (pm *P2PManager) handleFileListResponse(msg protocol.UDPMessage, remote *net.UDPAddr) {
	peer, ok := pm.findPeerByRemote(remote)
	if !ok {
		log.Printf("Cannot handle file list response from %s. Peer not found!", remote)
		return
	}
	bytes, _ := json.Marshal(msg.Payload)
	var p protocol.FileListPayload
	json.Unmarshal(bytes, &p)

	log.Printf("Received File List from %s: %d files", peer.Name, len(p.Files))
	if pm.OnFileList != nil {
		pm.OnFileList(peer.ID, p.Files)
	}

}

func (pm *P2PManager) RequestDownload(peerID string, fileName string) (string, error) {
	peer, ok := pm.Peers[peerID]
	if !ok || peer.State != StateConnected {
		return "", fmt.Errorf("peer not connected")
	}
	log.Printf("Starting download: File=%s from Peer=%s", fileName, peerID)

	requestID := fmt.Sprintf("%s_%d", peerID, time.Now().UnixNano())

	pm.Downloader.StartDownload(requestID, peerID, peer.Name, fileName)

	addr := &net.UDPAddr{IP: peer.IP, Port: peer.Port}
	req := protocol.DownloadRequest{
		FileName:  fileName,
		RequestID: requestID,
	}

	pm.udp.Send(addr, protocol.UDPMessage{
		Type:    protocol.TypeDownloadRequest,
		Payload: req,
	})

	return requestID, nil
}

func (pm *P2PManager) handleDownloadRequest(msg protocol.UDPMessage, remote *net.UDPAddr) {
	peer, ok := pm.findPeerByRemote(remote)
	if !ok {
		log.Printf("Cannot handle download request from %s. Peer not found!", remote)
		return
	}

	bytes, _ := json.Marshal(msg.Payload)
	var req protocol.DownloadRequest
	json.Unmarshal(bytes, &req)

	log.Printf("!!! RECEIVED DOWNLOAD REQUEST !!! from %s for file: %s (Peer: %s)", peer.Name, req.FileName, peer.ID)

	go func() {
		info, err := pm.Uploader.PrepareFile(req.RequestID, peer.ID, peer.Name, req.FileName)
		if err != nil {
			log.Printf("Error preparing file: %v", err)
			return
		}
		pm.udp.Send(remote, protocol.UDPMessage{
			Type:    protocol.TypeDownloadInfo,
			Payload: info,
		})
		log.Printf("Download info sent to %s: %d chunks for %s", peer.Name, info.ChunkCount, req.FileName)
	}()
}

func (pm *P2PManager) handleDownloadInfo(msg protocol.UDPMessage, remote *net.UDPAddr) {
	bytes, _ := json.Marshal(msg.Payload)
	var info protocol.DownloadInfo
	json.Unmarshal(bytes, &info)

	log.Printf("Download info received: %s (%d chunks)", info.FileName, info.ChunkCount)

	if err := pm.Downloader.HandleDownloadInfo(info); err != nil {
		log.Printf("Error handling download info: %v", err)
		return
	}

	go pm.startChunkRequests(info.RequestID)
}

func (pm *P2PManager) startChunkRequests(requestID string) {
	windowSize := 8
	for i := 0; i < windowSize; i++ {
		pm.requestNextChunk(requestID)
		time.Sleep(time.Millisecond * 10)
	}
}

func (pm *P2PManager) requestNextChunk(requestID string) {
	download, exists := pm.Downloader.GetDownload(requestID)
	if !exists || download.Status != DownloadActive {
		return
	}

	index, hasMore := pm.Downloader.GetNextChunkToRequest(requestID)
	if !hasMore {
		return
	}

	peer, ok := pm.Peers[download.PeerID]
	if !ok {
		return
	}

	addr := &net.UDPAddr{IP: peer.IP, Port: peer.Port}
	req := protocol.ChunkRequest{
		RequestID:  requestID,
		ChunkIndex: index,
	}
	log.Printf("Requesting chunk %d from %s", index, download.PeerName)
	pm.udp.Send(addr, protocol.UDPMessage{
		Type:    protocol.TypeChunkRequest,
		Payload: req,
	})
}

func (pm *P2PManager) handleChunkRequest(msg protocol.UDPMessage, remote *net.UDPAddr) {
	go func() {
		bytes, _ := json.Marshal(msg.Payload)
		var req protocol.ChunkRequest
		json.Unmarshal(bytes, &req)

		chunkResp, err := pm.Uploader.GetChunk(req.RequestID, req.ChunkIndex)
		if err != nil {
			log.Printf("Error getting chunk: %v", err)
			return
		}

		pm.udp.Send(remote, protocol.UDPMessage{
			Type:    protocol.TypeChunkResponse,
			Payload: chunkResp,
		})
	}()
}

func (pm *P2PManager) handleChunkResponse(msg protocol.UDPMessage) {
	go func() {
		bytes, _ := json.Marshal(msg.Payload)
		var resp protocol.ChunkResponse
		json.Unmarshal(bytes, &resp)

		if err := pm.Downloader.SaveChunk(resp.RequestID, resp.ChunkIndex, resp.ChunkData, resp.ChunkSHA); err != nil {
			log.Printf("Error saving chunk: %v", err)
			return
		}

		pm.requestNextChunk(resp.RequestID)
	}()
}
