package localisator

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/Vitus43/tovary/pkg/events"
	"github.com/Vitus43/tovary/pkg/subjects"
	"github.com/Vitus43/tovary/pkg/utils"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

// 1 opcja
// rola: "storage"
// mapa string -> Lokalizacja(x, y, ID)
// REQUEST(string) -> []Lokalizacji

// 2 opcja
// rola: "storage" lokalizacja: x,y

// Obliczmy jaki jest najblizszy i zwracamy tylko ID w odpowiedzi
// REQUEST(string, XY) -> ID

const (
	CopperMine        = "copperMine"
	IronMine          = "ironMine"
	IronSmelter       = "ironSmelter"
	Storage           = "storage"
	CopperSmelter     = "copperSmelter"
	SandMine          = "sandMine"
	SandSmelter       = "sandSmelter"
	OilPump           = "oilPump"
	OilRefinery       = "oilRefinery"
	Factory           = "factory"
	Assembler         = "assembler"
	HeartbeatInterval = time.Second * 2
)

type localisator struct {
	broker *nats.Conn
	data   map[string][]*EntityInfo
	lock   sync.Mutex
}

type EntityInfo struct {
	ID    uuid.UUID
	Cords Coordinates
}

func NewLocalisator() error {
	c, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return err
	}

	l := &localisator{broker: c, data: map[string][]*EntityInfo{}}

	l.broker.Subscribe(subjects.Localisation, func(msg *nats.Msg) {
		l.SaveLoc(msg)
	})

	l.broker.Subscribe(subjects.Locate, func(msg *nats.Msg) {
		l.ShareLoc(msg)
	})

	return nil
}

func (l *localisator) SaveLoc(msg *nats.Msg) {
	l.lock.Lock()
	defer l.lock.Unlock()
	e := events.LocalisationEvent{}
	err := json.Unmarshal(msg.Data, &e)
	if err != nil {
		fmt.Println("unmarshal localisation event", err)
		return
	}

	l.data[e.Type] = append(l.data[e.Type], &EntityInfo{
		ID:    e.ID,
		Cords: Coordinates{X: e.X, Y: e.Y},
	})

	fmt.Printf("Registered %s %s (%d,%d)\n", e.Type, e.ID, e.X, e.Y)
}

func (l *localisator) ShareLoc(msg *nats.Msg) {
	l.lock.Lock()
	defer l.lock.Unlock()
	e := events.LocalisationEvent{}
	err := json.Unmarshal(msg.Data, &e)
	if err != nil {
		fmt.Println("unmarshal localisation event", err)
		return
	}

	locs, ok := l.data[e.Type]
	if !ok {
		data, err := json.Marshal([]*EntityInfo{})
		if err != nil {
			fmt.Println("marshal entity info", err)
		}

		msg.Respond(data)
		return
	}

	data, err := json.Marshal(locs)
	if err != nil {
		fmt.Println("marshal entity info", err)
	}

	msg.Respond(data)
}

func RegisterLocation(b *nats.Conn, t string, ID uuid.UUID, c Coordinates) {
	ev := events.LocalisationEvent{Type: t, ID: ID, X: c.X, Y: c.Y}
	data, err := json.Marshal(ev)
	if err != nil {
		fmt.Println("error in marshal localisation", err)
	}

	b.Publish(subjects.Localisation, data)
}

type Location struct {
	Coordinates
	Name           string
	Broker         *nats.Conn
	ID             uuid.UUID
	NearestStorage uuid.UUID
	Time           chan int
	StopChan       chan struct{}
	Type           string

	Started chan struct{}
	Run     bool
}

// TODO spawner, slucha na spawn, dostaje spawnevent z typem do stworzenia i go tworzy. Odsyla inofmacje o stworzonej lokalizacji
 
// TODO moze sluchac na kanale stop.<id> i wysylac stop sygnal do kanalu StopChan gdy przyjdzie wiadomosc
func (l *Location) Stop() {
	l.StopChan <- struct{}{}
}

func (l *Location) Heartbeat() {
	ticker := time.NewTicker(HeartbeatInterval)
	for {
		select {
		case <-ticker.C:
			l.Broker.Publish(subjects.Heartbeat, nil)
		case <-l.StopChan:
			return
		}
	}
}

func (l *Location) HealthCheck(msg *nats.Msg) {
	h := events.HealthEvent{Status: "OK"}
	data, err := json.Marshal(h)
	if err != nil {
		fmt.Println("error in health check", err)
	}
	msg.Respond(data)
}

type Coordinates struct {
	X int
	Y int
}

// CalcDistance calculates distance to given points
func (c *Coordinates) CalcDistance(x, y int) float64 {
	return math.Sqrt(math.Pow(float64(x-c.X), 2.0) + math.Pow(float64(y-c.Y), 2.0))
}

func NewLocation(name, locationType string) (*Location, error) {
	c, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}

	st := &Location{Name: name, Type: locationType, Broker: c,
		Time: make(chan int), Started: make(chan struct{}), StopChan: make(chan struct{}),
		ID: uuid.New(), Coordinates: Coordinates{X: rand.IntN(21) - 10, Y: rand.IntN(21) - 10}}

	c.Subscribe(subjects.Health, func(msg *nats.Msg) {
		st.HealthCheck(msg)
	})

	c.Subscribe(subjects.Time, func(msg *nats.Msg) {
		e := events.TimeEvent{}
		err := json.Unmarshal(msg.Data, &e)
		if err != nil {
			fmt.Println("unmarshal time event", err)
		}

		select {
		case <-time.After(time.Second):
			return
		case <-st.Started:
		}

		st.Time <- e.Hour
	})

	RegisterLocation(st.Broker, locationType, st.ID, Coordinates{X: st.X, Y: st.Y})

	if err = st.LookUp(); err != nil {
		return nil, err
	}

	return st, nil
}

func (l *Location) LookUp() error {
	e := events.LocalisationEvent{Type: Storage}
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	var storages []EntityInfo

	resp, err := utils.NewRetryOpts(l.Broker).WithSubject(subjects.Locate).WithPayload(data).Run()
	if err != nil {
		return err
	}

	err = json.Unmarshal(resp.Data, &storages)
	if err != nil {
		return err
	}

	if len(storages) == 0 {
		return errors.New("no storages")
	}

	fmt.Println("storages", storages)

	if len(storages) > 0 {
		min := storages[0].Cords.CalcDistance(l.X, l.Y)
		nearest := storages[0].ID
		for _, storage := range storages {
			dist := storage.Cords.CalcDistance(l.X, l.Y)
			if dist < min {
				min = dist
				nearest = storage.ID
			}
		}

		l.NearestStorage = nearest
	}

	fmt.Printf("%s storage ->%s\n", l.Type, l.NearestStorage)

	return nil
}
