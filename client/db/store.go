package db

import (
	"database/sql"
	"fmt"

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
