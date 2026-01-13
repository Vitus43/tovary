package copper

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Vitus43/tovary/pkg/events"
	"github.com/Vitus43/tovary/pkg/localisator"
	"github.com/Vitus43/tovary/pkg/subjects"
)

const smeltPerHour = 10

func NewSmelter(name string) (*localisator.Location, error) {
	l, err := localisator.NewLocation(name, localisator.CopperSmelter)
	if err != nil {
		return nil, err
	}

	go runSmelter(l)

	return l, nil
}

func runSmelter(l *localisator.Location) {
	copperWire := 0
	copper := 0
	deliveryTreshhold := 50
	received := 0

	firstDay := true
	l.Run = true
	close(l.Started)

	for h := range l.Time {
		
		if h == 23 {
			firstDay = false
		}

		if h == 1 {
			received = 0
		}

		if copper < 50 {
			e := events.ResourceEvent{Resources: map[string]int{subjects.Copper: 100 - copper}}
			data, err := json.Marshal(e)
			if err != nil {
				l.Broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(l.Name, "failed to marshal resource event"))
			}

			msg, err := l.Broker.Request(subjects.Request+"."+l.NearestStorage.String(), data, time.Second*3)
			if err != nil {
				fmt.Println(err)

				continue
			}

			if msg == nil {
				fmt.Println("nil msg")

				continue
			}

			e = events.ResourceEvent{}

			err = json.Unmarshal(msg.Data, &e)
			if err != nil {
				l.Broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(l.Name, "failed to unmarshal resource event"))
			}

			copper += e.Resources[subjects.Copper]
			received += e.Resources[subjects.Copper]
		}

		if received == 0 && h == 0 && !firstDay{
				l.Broker.Publish(subjects.Error+"."+subjects.NoProducers, events.MustMarshal(l.Name, "failed to get resources"))
				fmt.Println("failed to get resources")
			}

		// logika
		if copper > 10 {
			copperWire += smeltPerHour
			copper -= smeltPerHour
		}

		if copperWire >= deliveryTreshhold {
			e := events.ResourceEvent{Resources: map[string]int{subjects.CopperWire: copperWire}}
			data, err := json.Marshal(e)
			if err != nil {
				l.Broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(l.Name, "failed to marshal resource event"))
			}

			// Subject
			l.Broker.Publish(idfy(subjects.Store, l.NearestStorage), data)
			copperWire = 0
		}

	}
}
