package factory

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Vitus43/tovary/pkg/events"
	"github.com/Vitus43/tovary/pkg/localisator"
	"github.com/Vitus43/tovary/pkg/subjects"
	"github.com/google/uuid"
)

const smeltPerHour = 10

func NewFactory(name string) (*localisator.Location, error) {
	l, err := localisator.NewLocation(name, localisator.Factory)
	if err != nil {
		return nil, err
	}

	go runFactory(l)

	return l, nil
}

func runFactory(l *localisator.Location) {
	copperWire := 0
	ironPlate := 0
	chip := 0
	deliveryTreshhold := 50
	receivedCw := 0
	receivedIp := 0
	firstDay := true

	l.Run = true
	close(l.Started)

	for h := range l.Time {

		if h == 23 {
			firstDay = false
		}

		if h == 1 {
			receivedCw = 0
			receivedIp = 0
		}

		if copperWire < 50 {
			e := events.ResourceEvent{Resources: map[string]int{subjects.CopperWire: 100 - copperWire, subjects.IronPlate: 100 - ironPlate}}
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

			copperWire += e.Resources[subjects.CopperWire]
			ironPlate += e.Resources[subjects.IronPlate]
			receivedCw += e.Resources[subjects.CopperWire]
			receivedIp += e.Resources[subjects.IronPlate]
		}

		if receivedCw == 0 && h == 0 && !firstDay {
			l.Broker.Publish(subjects.Error+"."+subjects.NoProducers, events.MustMarshal(l.Name, "failed to get resources"))
		}

		if receivedIp == 0 && h == 0 && !firstDay {
			l.Broker.Publish(subjects.Error+"."+subjects.NoProducers, events.MustMarshal(l.Name, "failed to get resources"))
		}

		// logika
		if copperWire > 10 {
			chip += smeltPerHour
			copperWire -= smeltPerHour
			ironPlate -= smeltPerHour
		}

		if chip >= deliveryTreshhold {
			e := events.ResourceEvent{Resources: map[string]int{subjects.Chip: chip}}
			data, err := json.Marshal(e)
			if err != nil {
				l.Broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(l.Name, "failed to marshal resource event"))
			}

			// Subject
			l.Broker.Publish(idfy(subjects.Store, l.NearestStorage), data)
			chip = 0
		}

	}
}

func idfy(s string, id uuid.UUID) string {
	return s + "." + id.String()
}
