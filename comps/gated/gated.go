package gated

import (
	"game/core"
	"game/vars"
	"log"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

type Gate interface {
	Open()
	Close()
}

type Comp struct {
	Props map[string]string
	open  bool
	gates []Gate
}

func (c *Comp) Init(_ core.Entity) {
	for count := 1; c.Props["gate"+strconv.Itoa(count)] != ""; count++ {
		prop := "gate" + strconv.Itoa(count)
		gateID, err := strconv.Atoi(c.Props[prop])
		if err != nil {
			log.Panicf("Invalid gate ID %s", c.Props[prop])
		}
		gate := vars.World.Get(uint(gateID)).(Gate)
		if gate == nil {
			log.Panicf("Gate entity %d not found", gateID)
		}
		c.gates = append(c.gates, gate)
		gate.Open()
	}
	c.open = true
}

func (c *Comp) Update(_ float64)                     {}
func (c *Comp) Remove()                              {}
func (c *Comp) Draw(_ *core.Pipeline, _ ebiten.GeoM) {}

func (c *Comp) Open() {
	if c.open {
		return
	}
	for _, gate := range c.gates {
		gate.Open()
	}
	c.open = true
}

func (c *Comp) Close() {
	if !c.open {
		return
	}
	for _, gate := range c.gates {
		gate.Close()
	}
	c.open = false
}
