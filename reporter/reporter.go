package reporter

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Vitus43/tovary/pkg/events"
	"github.com/Vitus43/tovary/pkg/localisator"
	"github.com/Vitus43/tovary/pkg/storage"
	"github.com/Vitus43/tovary/pkg/subjects"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type Reporter struct {
	Broker   *nats.Conn
	Storages []*localisator.EntityInfo
}

type StorageReport struct {
	StorageID   uuid.UUID
	StorageName string
	Resources   map[string]int
}

func (sr *StorageReport) printDiff(new, old map[string]int) {
	dif := make(map[string]int)
	for resourceName, newValue := range new {
		oldValue, ok := old[resourceName]
		if !ok {
			dif[resourceName] = newValue
			continue
		}
		if newValue == oldValue {
			continue
		}
		dif[resourceName] = newValue - oldValue
	}

	fmt.Printf("Resource difference in %s: %v\n", sr.StorageName, dif)
}

func NewReporter() (*Reporter, error) {
	b, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}

	r := &Reporter{Broker: b, Storages: make([]*localisator.EntityInfo, 0)}

	return r, nil
}

func (r *Reporter) WaitForTovary() {
	sub, _ := r.Broker.Subscribe(subjects.Time, func(msg *nats.Msg) {
		r.Run()
	})

	sub.AutoUnsubscribe(1)
}

func (r *Reporter) Run() {
	l := events.LocalisationEvent{Type: "storage"}
	data, err := json.Marshal(l)
	if err != nil {
		fmt.Println(err)
	}

	msg, err := r.Broker.Request(subjects.Locate, data, time.Second*3)
	if err != nil {
		fmt.Println(err)
	}

	if msg == nil {
		fmt.Println("msg empty")
	}

	if r.Storages == nil {
		r.Storages = make([]*localisator.EntityInfo, 0)
	}

	err = json.Unmarshal(msg.Data, &r.Storages)
	if err != nil {
		fmt.Println(err)
	}

	for _, s := range r.Storages {
		fmt.Printf("Found storage: %s\n", s.ID.String())
	}


		r.Broker.Subscribe(subjects.Report, func(msg *nats.Msg) {
			var s events.Storage
			err := json.Unmarshal(msg.Data, &s)
			if err != nil {
				fmt.Println(err)
			}

			report := StorageReport{StorageID: s.ID, StorageName: s.Name}
			if report.Resources == nil {
				report.Resources = make(map[string]int)
				for _, res := range storage.SupportedResources {
					report.Resources[res] = 0
				}
			}

			report.printDiff(s.Resources, report.Resources)
			report.Resources = s.Resources
		})
	

}
