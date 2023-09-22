package actor

import (
	"fmt"
	"game/assets"
	"game/utils"

	"github.com/hajimehoshi/ebiten/v2"
)

const AICombatState AIState = "Combat"

var AIDebugDraw = false

type AI struct {
	Fsm           *AIFsm
	Actor, Target *Actor
}

func (c *AI) State() AIState                    { return c.Fsm.State }
func (c *AI) SetState(states []AIWeightedState) { c.Fsm.setState(states) }
func (c *AI) SetCombatOptions(combatOptions []AIWeightedState) {
	c.Fsm.Actions[AICombatState] = &AIAction{Next: func() []AIWeightedState { return combatOptions }}
}

func (c *AI) Init(a *Actor) {
	c.Actor = a
}

func (c *AI) Update(dt float64) {
	if c.Fsm != nil {
		c.Fsm.update(dt)
	}
}

func (c *AI) Draw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	if !AIDebugDraw {
		return
	}
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-5, -10)
	utils.DrawText(screen, fmt.Sprintf(`AI:%s`, c.Fsm.State), assets.TinyFont, op)
}

func (c *AI) InTargetRange(minDist, maxDist float64) bool {
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

func (c *AI) InRangeFunc(maxDist float64) func() bool {
	return func() bool { return c.InTargetRange(0, maxDist) }
}

func (c *AI) OutRangeFunc(maxDist float64) func() bool {
	return func() bool { return !c.InTargetRange(0, maxDist) }
}

func (c *AI) EnoughStamina(minStamina float64) func() bool {
	return func() bool { return c.Actor.Stats.Stamina >= minStamina }
}

func (c *AI) SetSpeedFunc(speed, maxSpeed float64) func() {
	return func() { c.Actor.Speed, c.Actor.Body.MaxX = speed, maxSpeed }
}

func (c *AI) distFromTarget() float64 {
	if c.Target == nil {
		return -1
	}
	tx, ty := c.Target.Position()
	x, y := c.Actor.Position()

	return utils.Distante(x, y, tx, ty)
}
