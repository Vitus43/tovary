package main

import (
	"fmt"
	"time"

	"github.com/Vitus43/tovary/pkg/localisator"
	"github.com/Vitus43/tovary/pkg/producers/assembler"
	"github.com/Vitus43/tovary/pkg/producers/copper"
	"github.com/Vitus43/tovary/pkg/producers/factory"
	"github.com/Vitus43/tovary/pkg/producers/glass"
	"github.com/Vitus43/tovary/pkg/producers/iron"
	"github.com/Vitus43/tovary/pkg/producers/plastic"
	"github.com/Vitus43/tovary/pkg/storage"
	"github.com/Vitus43/tovary/timekeeper"
)

const (
	copperMines    = 5
	copperSmelters = 5
	ironMines      = 5
	ironSmelters   = 5
	sandMines      = 5
	sandSmelters   = 5
	oilPumps       = 5
	oilRefineries  = 5
	storages       = 5
	factories      = 5
	assemblers     = 2
)

func main() {
	// c, err := nats.Connect(nats.DefaultURL)
	// if err != nil {
	// 	panic(err)
	// }

	// c.Subscribe(">", func(msg *nats.Msg) {
	// 	fmt.Println(msg.Subject, string(msg.Data))
	// })

	err := localisator.NewLocalisator()
	if err != nil {
		panic(err)
	}

	for i := range storages {
		_, err = storage.NewStorage(fmt.Sprintf("storage-%d", i))
		if err != nil {
			panic(err)
		}
	}

	for i := range copperMines {
		_, err := copper.NewCopperMine(fmt.Sprintf("copper-mine-%d", i))
		if err != nil {
			panic(err)
		}
	}

	for i := range ironMines {
		_, err = iron.NewIronMine(fmt.Sprintf("iron-mine-%d", i))
		if err != nil {
			panic(err)
		}
	}

	for i := range sandMines {
		_, err = glass.NewSandMine(fmt.Sprintf("sand-mine-%d", i))
		if err != nil {
			panic(err)
		}
	}

	for i := range oilPumps {
		_, err = plastic.NewOilPump(fmt.Sprintf("oil-pump-%d", i))
		if err != nil {
			panic(err)
		}
	}

	for i := range copperSmelters {
		_, err = copper.NewSmelter(fmt.Sprintf("copper-smelter-%d", i))
		if err != nil {
			panic(err)
		}
	}

	for i := range ironSmelters {
		_, err = iron.NewSmelter(fmt.Sprintf("iron-smelter-%d", i))
		if err != nil {
			panic(err)
		}
	}

	for i := range sandSmelters {
		_, err = glass.NewSmelter(fmt.Sprintf("glass-smelter-%d", i))
		if err != nil {
			panic(err)
		}
	}

	for i := range oilRefineries {
		_, err = plastic.NewRefinery(fmt.Sprintf("oil-refinery-%d", i))
		if err != nil {
			panic(err)
		}
	}

	for i := range factories {
		_, err = factory.NewFactory(fmt.Sprintf("factory-%d", i))
		if err != nil {
			panic(err)
		}
	}

	for i := range assemblers {
		_, err = assembler.NewAssembler(fmt.Sprintf("assembler-%d", i))
		if err != nil {
			panic(err)
		}
	}

	_, err = timekeeper.NewTimekeeper()
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 7)
}
