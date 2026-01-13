package spawner

import "github.com/nats-io/nats.go"

type Spawner struct {
	Broker *nats.Conn
}

func NewSpawner() (*Spawner, error) {
	b, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}
	s := &Spawner{Broker: b}
	return s, nil
}
