package ai

import (
	"fmt"
	"game/assets"
	"game/comps/control"
	"game/comps/stats"
	"game/core"
	"game/utils"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	Tag               = "AI"
	CombatState State = "Combat"
)

var DebugDraw = false

type Comp struct {
	*core.Entity
	Fsm    *Fsm
	Target *core.Entity
}

func (c *Comp) Tag() string { return Tag }

func (c *Comp) State() State                    { return c.Fsm.State }
func (c *Comp) SetState(states []WeightedState) { c.Fsm.setState(states) }
func (c *Comp) SetCombatOptions(combatOptions []WeightedState) {
	c.Fsm.Actions[CombatState] = &Action{Next: func() []WeightedState { return combatOptions }}
}

func (c *Comp) Init(entity *core.Entity) {
	c.Entity = entity
}

func (c *Comp) Update(dt float64) {
	if fsm := c.Fsm; fsm != nil {
		fsm.update(dt)
	}
}

func (c *Comp) Draw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	if !DebugDraw || c.Fsm == nil {
		return
	}
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-5, -10)
	utils.DrawText(screen, fmt.Sprintf(`AI:%s`, c.Fsm.State), assets.TinyFont, op)
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

func (c *Comp) InRangeFunc(maxDist float64) func() bool {
	return func() bool { return c.InTargetRange(0, maxDist) }
}

func (c *Comp) OutRangeFunc(maxDist float64) func() bool {
	return func() bool { return !c.InTargetRange(0, maxDist) }
}

func (c *Comp) EnoughStamina(minStamina float64) func() bool {
	return func() bool { return core.GetComponent[*stats.Comp](c.Entity).Stamina >= minStamina }
}

func (c *Comp) SetSpeedFunc(speed, maxSpeed float64) func() {
	return func() {
		if control := core.GetComponent[*control.Comp](c.Entity); control != nil {
			control.SetSpeed(speed, maxSpeed)
		}
	}
}

func (c *Comp) distFromTarget() float64 {
	if c.Target == nil {
		return -1
	}
	tx, ty := c.Target.Position()
	x, y := c.Entity.Position()

	return utils.Distante(x, y, tx, ty)
}
