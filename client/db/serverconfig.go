package db

import "time"

type ServerConfig struct {
	Address  string    `json:"address"`
	LastUsed time.Time `json:"last_used"`
}

func (s *Store) SaveServerConfig(addr string) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO server_configs (address, last_used) VALUES (?, ?)", addr, time.Now().Unix())
	return err
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
