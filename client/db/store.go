package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.Init(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Init() error {
	for _, q := range migrationQueries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("error executing query %q: %v", q, err)
		}
	}
	return nil
}

func (s *Store) GetIdentity(n string) (string, string, error) {
	var id, name string
	err := s.db.QueryRow("SELECT id, name FROM identity WHERE name = ?", n).Scan(&id, &name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	return id, name, err
}

func (s *Store) SaveIdentity(id string, name string) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO identity (id, name) VALUES (?, ?)", id, name)
	return err
}

func (s *Store) SaveServerConfig(addr string) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO server_configs (address, last_used) VALUES (?, ?)", addr, time.Now().Unix())
	return err
}

type ServerConfig struct {
	Address  string    `json:"address"`
	LastUsed time.Time `json:"last_used"`
}

func (s *Store) GetServerConfigs() ([]ServerConfig, error) {
	rows, err := s.db.Query("SELECT address, last_used FROM server_configs ORDER BY last_used DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []ServerConfig
	for rows.Next() {
		var c ServerConfig
		var lastUsed int64
		if err := rows.Scan(&c.Address, &lastUsed); err != nil {
			return nil, err
		}
		c.LastUsed = time.Unix(lastUsed, 0)
		configs = append(configs, c)
	}
	return configs, nil
}

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
