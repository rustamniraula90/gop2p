package main

import (
	"net"
	"time"
)

type ClientInfo struct {
	ID       string
	Name     string
	IP       net.IP
	Port     int
	LastSeen time.Time
}

type Registry struct {
	clients map[string]*ClientInfo
}

func NewRegistry() *Registry {
	return &Registry{
		clients: make(map[string]*ClientInfo),
	}
}

func (r Registry) Register(id string, name string, addr *net.UDPAddr) {
	r.clients[id] = &ClientInfo{
		ID:       id,
		Name:     name,
		IP:       addr.IP,
		Port:     addr.Port,
		LastSeen: time.Now(),
	}
}

func (r Registry) List() []ClientInfo {
	list := make([]ClientInfo, 0, len(r.clients))
	for _, c := range r.clients {
		list = append(list, *c)
	}
	return list
}

func (r Registry) Get(id string) (*ClientInfo, bool) {
	client, ok := r.clients[id]
	if !ok {
		return nil, false
	}
	return client, true
}
