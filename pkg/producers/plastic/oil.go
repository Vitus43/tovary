package plastic

import (
	"encoding/json"

	"github.com/Vitus43/tovary/pkg/events"
	"github.com/Vitus43/tovary/pkg/localisator"
	"github.com/Vitus43/tovary/pkg/subjects"
	"github.com/google/uuid"
)

const (
	oilDeliveryInterval = 3
	oilPerHour          = 3
)

func NewOilPump(name string) (*localisator.Location, error) {
	l, err := localisator.NewLocation(name, localisator.OilPump)
	if err != nil {
		return nil, err
	}

	go runPump(l)

	return l, nil
}

func runPump(l *localisator.Location) {
	hourCounter := 0
	oil := 0

	l.Run = true
	close(l.Started)

	for range l.Time {
		oil += oilPerHour

		if hourCounter == oilDeliveryInterval {
			e := events.ResourceEvent{Resources: map[string]int{subjects.Oil: oil}}
			data, err := json.Marshal(e)
			if err != nil {
				l.Broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(l.Name, "failed to marshal resource event"))
			}

			l.Broker.Publish(idfy(subjects.Store, l.NearestStorage), data)
			hourCounter = 0
			oil = 0
		}

		hourCounter++
	}
}

func idfy(s string, id uuid.UUID) string {
	return s + "." + id.String()
}
