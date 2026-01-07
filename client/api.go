package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rustamniraula90/gop2p/client/db"
	"github.com/rustamniraula90/gop2p/protocol"
	"github.com/rustamniraula90/gop2p/web"
)

type APIServer struct {
	p2p     *P2PManager
	address string
	clients map[*websocket.Conn]bool
}

func NewAPIServer(p2p *P2PManager, port int) *APIServer {
	api := &APIServer{
		p2p:     p2p,
		address: fmt.Sprintf(":%d", port),
		clients: make(map[*websocket.Conn]bool),
	}
	p2p.OnPeerUpdate = api.BroadcastPeerUpdate
	p2p.OnConnectionRequest = api.BroadcastConnectionRequest
	p2p.OnMessage = func(senderID, text string) {
		api.BroadcastMessage(senderID, text, "")
	}
	p2p.OnFileList = api.BroadcastFileList
	p2p.Downloader.OnProgress = api.BroadcastDownloadProgress
	p2p.Uploader.OnProgress = api.BroadcastUploadProgress

	return api
}

func (api *APIServer) Start() {
	http.Handle("/", http.FileServer(http.FS(web.GetAssets())))
	http.HandleFunc("/ws", api.handleWS)
	http.HandleFunc("/api/identity", api.handleGetIdentity)
	http.HandleFunc("/api/server/config", api.handleServerConfigSave)
	http.HandleFunc("/api/server/config/list", api.handleServerConfigList)
	http.HandleFunc("/api/server/config/current", api.handleCurrentServerConfig)
	http.HandleFunc("/api/peers", api.handleFetchPeers)
	http.HandleFunc("/api/peers/remove", api.handleRemovePeer)
	http.HandleFunc("/api/connect/request", api.handleConnectRequest)
	http.HandleFunc("/api/connect/accept", api.handleConnectAccept)
	http.HandleFunc("/api/messages", api.handleGetMessages)
	http.HandleFunc("/api/message", api.handleSendMessage)
	http.HandleFunc("/api/files/request", api.handleRequestFiles)
	http.HandleFunc("/api/download/start", api.handleStartFileDownload)
	http.HandleFunc("/api/downloads", api.handleGetDownloads)
	http.HandleFunc("/api/uploads", api.handleGetUploads)
	http.HandleFunc("/api/downloads/clear", api.handleClearDownloads)
	http.HandleFunc("/api/uploads/clear", api.handleClearUploads)

	log.Printf("UI accessible at http://localhost%s", api.address)
	if err := http.ListenAndServe(api.address, nil); err != nil {
		log.Fatal("Failed to start API server", err)
	}
}

func (api *APIServer) handleGetIdentity(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(api.p2p.identity)
}

func (api *APIServer) handleFetchPeers(w http.ResponseWriter, r *http.Request) {
	go api.p2p.FetchPeers()
	w.WriteHeader(200)
}

func (api *APIServer) BroadcastPeerUpdate(peer *Peer) {
	msg := map[string]interface{}{
		"type": "peer_update",
		"data": map[string]interface{}{
			"id":        peer.ID,
			"name":      peer.Name,
			"status":    peer.State,
			"last_used": peer.LastUsed,
		},
	}
	api.broadcast(msg)
}

func (api *APIServer) BroadcastConnectionRequest(id string, name string) {
	msg := map[string]interface{}{
		"type": "connection_request",
		"data": map[string]interface{}{
			"id":   id,
			"name": name,
		},
	}
	api.broadcast(msg)
}

func (api *APIServer) BroadcastMessage(senderID, text, receiverID string) {
	msg := map[string]interface{}{
		"type": "message",
		"data": map[string]string{
			"sender_id":   senderID,
			"receiver_id": receiverID,
			"text":        text,
			"timestamp":   fmt.Sprintf("%d", time.Now().Unix()),
		},
	}
	api.broadcast(msg)
}

func (api *APIServer) BroadcastFileList(peerID string, files []protocol.FileInfo) {
	msg := map[string]interface{}{
		"type": "file_list",
		"data": map[string]interface{}{
			"peer_id": peerID,
			"files":   files,
		},
	}
	api.broadcast(msg)
}

func (api *APIServer) BroadcastDownloadProgress(requestID string, status string, progress float64, chunkIndex int, chunkStatus string) {
	download, exists := api.p2p.Downloader.GetDownload(requestID)
	if !exists {
		return
	}
	msg := map[string]interface{}{
		"type": "download_progress",
		"data": map[string]interface{}{
			"request_id":      download.RequestID,
			"peer_id":         download.PeerID,
			"peer_name":       download.PeerName,
			"file_name":       download.FileName,
			"status":          download.Status,
			"progress":        download.Progress,
			"original_size":   download.OriginalSize,
			"compressed_size": download.CompressedSize,
			"chunk_count":     download.ChunkCount,
			"last_chunk_idx":  chunkIndex,
			"last_chunk_stat": chunkStatus,
		},
	}
	api.broadcast(msg)
}

func (api *APIServer) BroadcastUploadProgress(requestID, status string, progress float64, chunkIndex int, chunkStatus string) {
	upload, exists := api.p2p.Uploader.GetUpload(requestID)
	if !exists {
		return
	}
	msg := map[string]interface{}{
		"type": "upload_progress",
		"data": map[string]interface{}{
			"request_id":      upload.RequestID,
			"peer_id":         upload.PeerID,
			"peer_name":       upload.PeerName,
			"file_name":       upload.FileName,
			"status":          upload.Status,
			"progress":        upload.Progress,
			"original_size":   upload.OriginalSize,
			"compressed_size": upload.CompressedSize,
			"chunk_count":     upload.ChunkCount,
			"last_chunk_idx":  chunkIndex,
			"last_chunk_stat": chunkStatus,
		},
	}
	api.broadcast(msg)
}

func (api *APIServer) broadcast(msg interface{}) {
	clients := make([]*websocket.Conn, 0, len(api.clients))
	for conn := range api.clients {
		clients = append(clients, conn)
	}

	for _, conn := range clients {
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := conn.WriteJSON(msg); err != nil {
			log.Println("Failed to broadcast peer update. Removing client.", err)
			conn.Close()
			delete(api.clients, conn)
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (api *APIServer) handleWS(writer http.ResponseWriter, request *http.Request) {
	ws, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println("Failed to upgrade websocket connection", err)
		return
	}
	api.clients[ws] = true

	api.sendIdentity(ws)
	api.sendPeers(ws)

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			delete(api.clients, ws)
			break
		}
	}
}

func (api *APIServer) sendIdentity(ws *websocket.Conn) {
	ws.WriteJSON(map[string]interface{}{
		"type": "identity",
		"data": api.p2p.identity,
	})
}

func (api *APIServer) sendPeers(ws *websocket.Conn) {
	peers := make([]interface{}, 0)
	for _, p := range api.p2p.Peers {
		peers = append(peers, map[string]interface{}{
			"id":        p.ID,
			"name":      p.Name,
			"status":    p.State,
			"last_used": p.LastUsed,
		})
	}

	ws.WriteJSON(map[string]interface{}{
		"type": "peers",
		"data": peers,
	})
}

func (api *APIServer) handleConnectRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TargetID string `json:"target_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	api.p2p.RequestConnection(req.TargetID)
	w.WriteHeader(200)
}

func (api *APIServer) handleConnectAccept(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RequesterID string `json:"requester_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	api.p2p.AcceptConnection(req.RequesterID)
	w.WriteHeader(200)
}

func (api *APIServer) handleServerConfigSave(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ServerAddr string `json:"server_addr"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := api.p2p.Store.SaveServerConfig(req.ServerAddr); err != nil {
		log.Printf("Error saving config: %v", err)
	}

	err := api.p2p.SetServer(req.ServerAddr)
	if err != nil {
		log.Printf("Error saving config: %v", err)
	}
	w.WriteHeader(200)

}

func (api *APIServer) handleServerConfigList(w http.ResponseWriter, r *http.Request) {
	configs, err := api.p2p.Store.GetServerConfigs()
	if err != nil {
		log.Printf("Error saving config: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(configs)
}

func (api *APIServer) handleCurrentServerConfig(writer http.ResponseWriter, request *http.Request) {
	config := db.ServerConfig{
		Address:  api.p2p.ServerAddr.String(),
		LastUsed: time.Now(),
	}
	json.NewEncoder(writer).Encode(config)
}

func (api *APIServer) handleRemovePeer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
	}
	api.p2p.RemovePeer(req.ID)
	w.WriteHeader(200)
}

func (api *APIServer) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TargetID string `json:"target_id"`
		Text     string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	api.p2p.SendMessage(req.TargetID, req.Text)
	api.BroadcastMessage("me", req.Text, req.TargetID)
	w.WriteHeader(200)

}

func (api *APIServer) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	peerID := r.URL.Query().Get("peer_id")
	if peerID == "" {
		http.Error(w, "peer_id required", 400)
		return
	}
	msgs, err := api.p2p.Store.LoadMessageByPeerID(peerID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(msgs)
}

func (api *APIServer) handleRequestFiles(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PeerID string `json:"peer_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	api.p2p.RequestFileList(req.PeerID)
	w.WriteHeader(200)
}

func (api *APIServer) handleStartFileDownload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PeerID   string `json:"peer_id"`
		FileName string `json:"file_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error unmarshaling download request: %v", err)
		http.Error(w, err.Error(), 400)
		return
	}

	if req.PeerID == "" || req.FileName == "" {
		log.Printf("Invalid download request: PeerID or FileName is empty")
		http.Error(w, "missing peer_id or file_name", 400)
		return
	}

	requestID, err := api.p2p.RequestDownload(req.PeerID, req.FileName)
	if err != nil {
		log.Printf("RequestDownload failed: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"request_id": requestID,
	})

}

func (api *APIServer) handleGetDownloads(w http.ResponseWriter, r *http.Request) {
	downloads := api.p2p.Downloader.GetAllDownloads()
	json.NewEncoder(w).Encode(downloads)
}

func (api *APIServer) handleGetUploads(w http.ResponseWriter, r *http.Request) {
	uploads := api.p2p.Uploader.GetAllUploads()
	json.NewEncoder(w).Encode(uploads)
}

func (api *APIServer) handleClearDownloads(w http.ResponseWriter, r *http.Request) {
	api.p2p.Downloader.ClearCompleted()
	w.WriteHeader(200)
}

func (api *APIServer) handleClearUploads(w http.ResponseWriter, r *http.Request) {
	api.p2p.Uploader.ClearCompleted()
	w.WriteHeader(200)
}
