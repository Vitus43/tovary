package plastic

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Vitus43/tovary/pkg/events"
	"github.com/Vitus43/tovary/pkg/localisator"
	"github.com/Vitus43/tovary/pkg/subjects"
)

const refinePerHour = 12

func NewRefinery(name string) (*localisator.Location, error) {
	l, err := localisator.NewLocation(name, localisator.OilRefinery)
	if err != nil {
		return nil, err
	}

	go runRefinery(l)

	return l, nil
}

func runRefinery(l *localisator.Location) {
	plastic := 0
	oil := 0
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

		if oil < 50 {
			e := events.ResourceEvent{Resources: map[string]int{subjects.Oil: 100 - oil}}
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

			oil += e.Resources[subjects.Oil]
			received += e.Resources[subjects.Oil]
		}
		
		if received == 0 && h == 0 && !firstDay {
			l.Broker.Publish(subjects.Error+"."+subjects.NoProducers, events.MustMarshal(l.Name, "failed to get resources"))
		}

		// logika
		if oil > 10 {
			plastic += refinePerHour
			oil -= refinePerHour
		}

		if plastic >= deliveryTreshhold {
			e := events.ResourceEvent{Resources: map[string]int{subjects.Plastic: plastic}}
			data, err := json.Marshal(e)
			if err != nil {
				l.Broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(l.Name, "failed to marshal resource event"))
			}

			// Subject
			l.Broker.Publish(idfy(subjects.Store, l.NearestStorage), data)
			plastic = 0
		}

	}
}
