package ai

import (
	"fmt"
	"game/assets"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/utils"

	"github.com/hajimehoshi/ebiten/v2"
)

type Actor interface {
	GetBody() *body.Comp
	GetHitbox() *hitbox.Comp
	GetAnim() *anim.Comp
	GetStats() *stats.Comp
	GetAI() *Comp
	SetSpeed(speed, maxSpeed float64)
}

const CombatState State = "Combat"

var DebugDraw = false

type Comp struct {
	Fsm            *Fsm
	Actor          Actor
	Target, entity *core.Entity
}

func (c *Comp) State() State                    { return c.Fsm.State }
func (c *Comp) SetState(states []WeightedState) { c.Fsm.setState(states) }
func (c *Comp) SetCombatOptions(combatOptions []WeightedState) {
	c.Fsm.Actions[CombatState] = &Action{Next: func() []WeightedState { return combatOptions }}
}

func (c *Comp) Init(entity *core.Entity) {
	c.entity = entity
}

func (c *Comp) Update(dt float64) {
	c.Fsm.update(dt)
}

func (c *Comp) Draw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	if !DebugDraw {
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
	return func() bool { return c.Actor.GetStats().Stamina >= minStamina }
}

func (c *Comp) SetSpeedFunc(speed, maxSpeed float64) func() {
	return func() { c.Actor.SetSpeed(speed, maxSpeed) }
}

func (c *Comp) distFromTarget() float64 {
	if c.Target == nil {
		return -1
	}
	tx, ty := c.Target.Position()
	x, y := c.entity.Position()

	return utils.Distante(x, y, tx, ty)
}
