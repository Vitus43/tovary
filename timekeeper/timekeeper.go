package timekeeper

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Vitus43/tovary/pkg/events"
	"github.com/Vitus43/tovary/pkg/subjects"
	"github.com/nats-io/nats.go"
)

type Timekeeper struct {
	broker *nats.Conn
}

const (
	hourDuration = time.Second / 8
)

func NewTimekeeper() (*Timekeeper, error) {
	c, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}

	t := &Timekeeper{broker: c}

	go t.Run()

	return t, nil
}

func (tk *Timekeeper) Run() {
	ticker := time.NewTicker(hourDuration)
	h := 0
	d := 0
	for range ticker.C {
		e := events.TimeEvent{Day: d, Hour: h}
		data, err := json.Marshal(e)
		if err != nil {
			fmt.Println("error in marshal time", err)
		}
		if d == 0 && h == 0 {
			tk.broker.Publish(subjects.NewCycle, data)
		}
		tk.broker.Publish(subjects.Time, data)

		h++

		if h == 24 {
			h = 0
			d++
		}
	}
}
