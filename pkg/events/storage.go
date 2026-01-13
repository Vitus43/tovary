package events

import "github.com/google/uuid"

type Storage struct {
	ID   uuid.UUID
	Name string
	Resources map[string]int
}

// report.<id>