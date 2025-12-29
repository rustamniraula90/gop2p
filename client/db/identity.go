package db

import (
	"database/sql"
	"errors"
)

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
