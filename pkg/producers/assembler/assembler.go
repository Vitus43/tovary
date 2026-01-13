package assembler

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

func NewAssembler(name string) (*localisator.Location, error) {
	l, err := localisator.NewLocation(name, localisator.Assembler)
	if err != nil {
		return nil, err
	}

	go runAssembler(l)

	return l, nil
}

func runAssembler(l *localisator.Location) {
	glass := 0
	plastic := 0
	chip := 0
	phone := 0
	deliveryTreshhold := 50
	receivedC, receivedP, receivedG := 0, 0, 0
	firstDay := true

	l.Run = true
	close(l.Started)

	for h := range l.Time {

		if h == 23 {
			firstDay = false
		}

		if h == 1 {
			receivedC = 0
			receivedP = 0
			receivedG = 0
		}

		if chip < 50 {
			e := events.ResourceEvent{Resources: map[string]int{
				subjects.Chip:    100 - chip,
				subjects.Glass:   100 - glass,
				subjects.Plastic: 100 - plastic}}

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

			chip += e.Resources[subjects.Chip]
			glass += e.Resources[subjects.Glass]
			plastic += e.Resources[subjects.Plastic]
			receivedC += e.Resources[subjects.Chip]
			receivedG += e.Resources[subjects.Glass]
			receivedP += e.Resources[subjects.Plastic]
		}

		if receivedC == 0 && h == 0 && !firstDay {
			l.Broker.Publish(subjects.Error+"."+subjects.NoProducers, events.MustMarshal(l.Name, "failed to get resources"))
		}

		if receivedG == 0 && h == 0 && !firstDay {
			l.Broker.Publish(subjects.Error+"."+subjects.NoProducers, events.MustMarshal(l.Name, "failed to get resources"))
		}

		if receivedP == 0 && h == 0 && !firstDay {
			l.Broker.Publish(subjects.Error+"."+subjects.NoProducers, events.MustMarshal(l.Name, "failed to get resources"))
		}

		// logika
		if chip > 10 {
			phone += smeltPerHour
			chip -= smeltPerHour
			glass -= smeltPerHour
			plastic -= smeltPerHour
		}

		if phone >= deliveryTreshhold {
			e := events.ResourceEvent{Resources: map[string]int{subjects.Phone: phone}}
			data, err := json.Marshal(e)
			if err != nil {
				l.Broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(l.Name, "failed to marshal resource event"))
			}

			// Subject
			l.Broker.Publish(idfy(subjects.Store, l.NearestStorage), data)
			phone = 0
		}

	}
}

func idfy(s string, id uuid.UUID) string {
	return s + "." + id.String()
}
