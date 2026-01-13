package listener

import (
	"encoding/json"
	"fmt"

	"github.com/Vitus43/tovary/pkg/events"
	"github.com/Vitus43/tovary/pkg/subjects"
	"github.com/nats-io/nats.go"
)

type Listener struct {
	Broker *nats.Conn
}

func NewListener() (*Listener, error) {
	b, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}

	l := &Listener{Broker: b}

	return l, nil
}

func (l *Listener) Run() {

	l.Broker.Subscribe(subjects.Error+"."+subjects.NoProducers, func(msg *nats.Msg) {
		l.HandleNoProducers(msg)
	})

	l.Broker.Subscribe(subjects.Error+"."+subjects.JSON, func(msg *nats.Msg) {
		l.HandleJSON(msg)
	})

	l.Broker.Subscribe(subjects.Error+"."+subjects.BadRequest, func(msg *nats.Msg) {
		l.HandleBadRequest(msg)
	})
}

func (l *Listener) HandleNoProducers(msg *nats.Msg) {
	var e events.ErrorEvent
	err := json.Unmarshal(msg.Data, &e)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("No Producers: %s %s\n", e.Location, e.Msg)
}

func (l *Listener) HandleJSON(msg *nats.Msg) {
	var e events.ErrorEvent
	err := json.Unmarshal(msg.Data, &e)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("JSON: %s: %s\n", e.Location, e.Msg)
}

func (l *Listener) HandleBadRequest(msg *nats.Msg) {
	var e events.ErrorEvent
	err := json.Unmarshal(msg.Data, &e)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Bad Request: %s: %s\n", e.Location, e.Msg)
}
