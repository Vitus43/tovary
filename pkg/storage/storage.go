package storage

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"slices"
	"sync"

	"github.com/Vitus43/tovary/pkg/events"
	"github.com/Vitus43/tovary/pkg/localisator"
	"github.com/Vitus43/tovary/pkg/subjects"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

var SupportedResources = []string{
	subjects.Copper,
	subjects.Iron,
	subjects.CopperWire,
	subjects.Sand,
	subjects.Oil,
	subjects.Plastic,
	subjects.IronPlate,
	subjects.Glass,
	subjects.Chip,
	subjects.Phone,
}

type Storage struct {
	broker         *nats.Conn
	name           string
	resources      map[string]int
	resourcesMutex sync.Mutex
	toReport       map[string]*Counter
	id             uuid.UUID
	x              int
	y              int
}

type Counter struct {
	mutex sync.Mutex
	count int
}

func (c *Counter) Update(count int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.count += count
}

func (c *Counter) Value() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.count
}

func (c *Counter) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.count = 0
}

func NewStorage(name string) (*Storage, error) {
	c, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}

	s := &Storage{broker: c, name: name, id: uuid.New(), x: rand.IntN(21) - 10, y: rand.IntN(21) - 10, resources: map[string]int{}, toReport: map[string]*Counter{}}
	s.initMaps()

	c.Subscribe(subjects.Store+"."+s.id.String(), func(msg *nats.Msg) {
		s.handleStore(msg)
	})

	c.Subscribe(subjects.Request+"."+s.id.String(), func(msg *nats.Msg) {
		s.handleRequest(msg)
	})

	c.Subscribe(subjects.Time, func(msg *nats.Msg) {
		s.dumpState(msg)
	})

	localisator.RegisterLocation(s.broker, localisator.Storage, s.id, localisator.Coordinates{X: s.x, Y: s.y})

	return s, nil

}

func (s *Storage) initMaps() {
	for _, res := range SupportedResources {
		s.resources[res] = 0
		s.toReport[res] = &Counter{}
	}
}

func (s *Storage) handleStore(msg *nats.Msg) {
	e := events.ResourceEvent{}
	err := json.Unmarshal(msg.Data, &e)
	if err != nil {
		s.broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(s.name, "failed to unmarshal resource event"))
	}

	for name, value := range e.Resources {
		ok := slices.Contains(SupportedResources, name)
		if !ok {
			s.broker.Publish(subjects.Error+"."+subjects.BadRequest,
				events.MustMarshal(s.name, fmt.Sprintf("unsupported resource detected: %s\n", name)))

			continue
		}

		if value <= 0 {
			s.broker.Publish(subjects.Error+"."+subjects.BadRequest,
				events.MustMarshal(s.name, fmt.Sprintf("negative value detected for resource %s: %d\n", name, value)))

			continue
		}

		s.resourcesMutex.Lock()
		s.resources[name] += value
		s.toReport[name].Update(value)
		s.resourcesMutex.Unlock()
	}
}

func (s *Storage) dumpState(msg *nats.Msg) {
	e := events.TimeEvent{}
	err := json.Unmarshal(msg.Data, &e)
	if err != nil {
		s.broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(s.name, "failed to unmarshal time event"))
	}

	if e.Hour != 0 {
		return
	}

	se := events.Storage{ID: s.id, Name: s.name, Resources: s.resources}
	data, err := json.Marshal(se)
	if err != nil {
		s.broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(s.name, "failed to marshal storage event"))
	}

	s.broker.Publish(subjects.Report, data)
}

func (s *Storage) handleRequest(msg *nats.Msg) {
	e := events.ResourceEvent{}
	err := json.Unmarshal(msg.Data, &e)
	if err != nil {
		s.broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(s.name, "failed to unmarshal resource event"))
		return
	}

	s.resourcesMutex.Lock()
	defer s.resourcesMutex.Unlock()

	canHandle := true

	for name, value := range e.Resources {
		ok := slices.Contains(SupportedResources, name)
		if !ok {
			s.broker.Publish(subjects.Error+"."+subjects.BadRequest,
				events.MustMarshal(s.name, fmt.Sprintf("unsupported resource detected: %s\n", name)))
			canHandle = false

			break
		}

		if value <= 0 {
			s.broker.Publish(subjects.Error+"."+subjects.BadRequest,
				events.MustMarshal(s.name, fmt.Sprintf("negative value detected for resource %s: %d\n", name, value)))
			canHandle = false

			break
		}

		c, ok := s.resources[name]
		if !ok {
			s.resources[name] = 0
		}

		if value > c {
			// fmt.Printf("not enough %s in store: want:%d stored: %d\n", name, value, c)
			canHandle = false

			break
		}
	}

	if !canHandle {
		data, err := json.Marshal(events.ResourceEvent{Resources: map[string]int{}})
		if err != nil {
			s.broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(s.name, "failed to marshal resource event"))
		}

		msg.Respond(data)

		return
	}

	for name, value := range e.Resources {
		s.resources[name] -= value
	}

	data, err := json.Marshal(e)
	if err != nil {
		s.broker.Publish(subjects.Error+"."+subjects.JSON, events.MustMarshal(s.name, "failed to marshal resource event"))
	}

	msg.Respond(data)
}
