package main

import "github.com/google/uuid"

type Identity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func LoadIdentity(name string) *Identity {
	return &Identity{
		ID:   uuid.NewString(),
		Name: name,
	}

}
