package glass

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Vitus43/tovary/pkg/events"
	"github.com/Vitus43/tovary/pkg/localisator"
	"github.com/Vitus43/tovary/pkg/subjects"
)

const smeltPerHour = 8

func NewSmelter(name string) (*localisator.Location, error) {
	l, err := localisator.NewLocation(name, localisator.SandSmelter)
	if err != nil {
		return nil, err
	}

	go runSmelter(l)

	return l, nil
}

func runSmelter(l *localisator.Location) {
	glass := 0
	sand := 0
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

		if sand < 50 {
			e := events.ResourceEvent{Resources: map[string]int{subjects.Sand: 100 - sand}}
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

			sand += e.Resources[subjects.Sand]
			received += e.Resources[subjects.Iron]
		}

		if received == 0 && h == 0 && !firstDay {
			l.Broker.Publish(subjects.Error+"."+subjects.NoProducers, events.MustMarshal(l.Name, "failed to get resources"))
			fmt.Println("failed to get resources")
		}

		// logika
		if sand > 10 {
			glass += smeltPerHour
			sand -= smeltPerHour
		}

		if glass >= deliveryTreshhold {
			e := events.ResourceEvent{Resources: map[string]int{subjects.Glass: glass}}
			data, err := json.Marshal(e)
			if err != nil {
				l.Broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(l.Name, "failed to marshal resource event"))
			}

			// Subject
			l.Broker.Publish(idfy(subjects.Store, l.NearestStorage), data)
			glass = 0
		}

	}
}
