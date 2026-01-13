package events

import (
	"github.com/google/uuid"
)

type LocalisationEvent struct {
	Type  string
	ID    uuid.UUID
	X int
	Y int
}
