package copper

import (
	"encoding/json"

	"github.com/Vitus43/tovary/pkg/events"
	"github.com/Vitus43/tovary/pkg/localisator"
	"github.com/Vitus43/tovary/pkg/subjects"
	"github.com/google/uuid"
)

const (
	copperDeliveryInterval = 4
	copperPerHour          = 5
)

func NewCopperMine(name string) (*localisator.Location, error) {
	l, err := localisator.NewLocation(name, localisator.CopperMine)
	if err != nil {
		return nil, err
	}

	go runMine(l)

	return l, nil
}

func runMine(l *localisator.Location) {
	hourCounter := 0
	copper := 0

	l.Run = true
	close(l.Started)

	//TODO zamienic w reszcie producerow for range na for select tak jak nizej
	for {
		select {
		case <- l.StopChan:
			return
		case <-l.Time:
			copper += copperPerHour

			if hourCounter == copperDeliveryInterval {
				e := events.ResourceEvent{Resources: map[string]int{subjects.Copper: copper}}
				data, err := json.Marshal(e)
				if err != nil {
					l.Broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(l.Name, "failed to marshal resource event"))
				}

				l.Broker.Publish(idfy(subjects.Store, l.NearestStorage), data)
				hourCounter = 0
				copper = 0
			}

			hourCounter++

		}
	}
}

func idfy(s string, id uuid.UUID) string {
	return s + "." + id.String()
}
