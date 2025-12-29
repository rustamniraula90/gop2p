package db

import (
	"net"
	"time"
)

type PeerEntity struct {
	ID       string
	Name     string
	IP       net.IP
	Port     int
	State    int
	LastUsed time.Time
}

func (s *Store) SavePeer(peer PeerEntity) error {
	_, err := s.db.Exec(`INSERT OR REPLACE INTO peers (id, name, ip, port, state, last_used) 
		VALUES (?, ?, ?, ?, ?, ?)`, peer.ID, peer.Name, peer.IP.String(), peer.Port, peer.State, peer.LastUsed.Unix())
	return err
}
