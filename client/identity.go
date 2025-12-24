package main

import (
	"github.com/google/uuid"
	"github.com/rustamniraula90/gop2p/client/db"
)

type Identity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func LoadIdentity(db *db.Store, n string) (*Identity, error) {
	id, name, err := db.GetIdentity(n)
	if err != nil {
		return nil, err
	}
	if id == "" {
		id = uuid.New().String()
		err = db.SaveIdentity(id, n)
		if err != nil {
			return nil, err
		}
		return &Identity{
			ID:   id,
			Name: n,
		}, nil
	}
	return &Identity{
		ID:   id,
		Name: name,
	}, nil

}
