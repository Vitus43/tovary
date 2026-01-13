package events

import "github.com/google/uuid"

type RaportEvent struct {
	ID    uuid.UUID
	Name  string
	Day   int
	Value int
}
