package ai

import (
	"fmt"
	"game/assets"
	"game/core"
	"game/utils"

	"github.com/hajimehoshi/ebiten/v2"
)

type Comp struct {
	Fsm            *Fsm
	Target, entity *core.Entity
	actions        []WeightedState
}

func (c *Comp) State() State                    { return c.Fsm.State }
func (c *Comp) SetState(states []WeightedState) { c.Fsm.setState(states) }

func (c *Comp) Init(entity *core.Entity) {
	c.entity = entity
	if c.Fsm != nil { // TODO: What happens here?
		c.Fsm.Comp = c
	}
}

func (c *Comp) Update(dt float64) {
	c.Fsm.update(dt)
}

func (c *Comp) DebugDraw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-5, -10)
	utils.DrawText(screen, fmt.Sprintf(`AI:%s`, c.Fsm.State), assets.BittyFont, op)
}

func (c *Comp) InTargetRange(minDist, maxDist float64) bool {
	dist := c.distFromTarget()
	if dist < 0 {
		return false
	}
	in := dist >= minDist
	out := true
	if maxDist > 0 {
		out = dist <= maxDist
	}

	return in && out
}

func (c *Comp) distFromTarget() float64 {
	if c.Target == nil {
		return -1
	}
	tx, ty := c.Target.Position()
	x, y := c.entity.Position()

	return utils.Distante(x, y, tx, ty)
}
