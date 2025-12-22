package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
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

	return api
}

func (api *APIServer) Start() {
	http.Handle("/", http.FileServer(http.FS(web.GetAssets())))
	http.HandleFunc("/ws", api.handleWS)
	http.HandleFunc("/api/identity", api.handleGetIdentity)
	http.HandleFunc("/api/peers", api.handleFetchPeers)
	http.HandleFunc("/api/connect/request", api.handleConnectRequest)
	http.HandleFunc("/api/connect/accept", api.handleConnectAccept)

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
			"last_seen": peer.LastSeen,
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
			"last_seen": p.LastSeen,
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
