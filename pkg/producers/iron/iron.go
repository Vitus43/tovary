package iron

import (
	"encoding/json"

	"github.com/Vitus43/tovary/pkg/events"
	"github.com/Vitus43/tovary/pkg/localisator"
	"github.com/Vitus43/tovary/pkg/subjects"
	"github.com/google/uuid"
)

const (
	ironDeliveryInterval = 5
	ironPerHour          = 6
)

func NewIronMine(name string) (*localisator.Location, error) {
	l, err := localisator.NewLocation(name, localisator.IronMine)
	if err != nil {
		return nil, err
	}

	go runMine(l)

	return l, nil
}

func runMine(l *localisator.Location) {
	hourCounter := 0
	iron := 0

	l.Run = true
	close(l.Started)

	for range l.Time {
		iron += ironPerHour

		if hourCounter == ironDeliveryInterval {
			e := events.ResourceEvent{Resources: map[string]int{subjects.Iron: iron}}
			data, err := json.Marshal(e)
			if err != nil {
				l.Broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(l.Name, "failed to marshal resource event"))
			}

			l.Broker.Publish(idfy(subjects.Store, l.NearestStorage), data)
			hourCounter = 0
			iron = 0
		}

		hourCounter++
	}
}

func idfy(s string, id uuid.UUID) string {
	return s + "." + id.String()
}
