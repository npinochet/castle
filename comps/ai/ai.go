package ai

import (
	"fmt"
	"game/assets"
	"game/core"
	"game/utils"

	"github.com/hajimehoshi/ebiten/v2"
)

const CombatState State = "Combat"

var DebugDraw = false

type Comp struct {
	*core.Entity
	Target *core.Entity
	Fsm    *Fsm
}

func (c *Comp) Init(entity *core.Entity) {
	c.Entity = entity
}

func (c *Comp) Update(dt float64) {
	if fsm := c.Fsm; fsm != nil {
		fsm.update(dt)
	}
}

func (c *Comp) SetCombatOptions(combatOptions []WeightedState) {
	c.Fsm.Actions[CombatState] = &Action{Next: func() []WeightedState { return combatOptions }}
}

func (c *Comp) Draw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	if !DebugDraw || c.Fsm == nil {
		return
	}
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-5, -10)
	utils.DrawText(screen, fmt.Sprintf(`AI:%s`, c.Fsm.State), assets.TinyFont, op)
}
