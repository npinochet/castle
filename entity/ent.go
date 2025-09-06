package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/entity/actor"
	"game/libs/bump"
)

const (
	entAnimFile                           = "ent"
	entWidth, entHeight                   = 12, 16
	entOffsetX, entOffsetY, entOffsetFlip = -8, -4, 17
	entSpeed, entMaxSpeed                 = 80, 30
	entHealth                             = 100
	entDamage, entPoise                   = 40, 41
	entExp                                = 40
)

type Ent struct {
	*core.BaseEntity
	*actor.Control
	anim   *anim.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	stats  *stats.Comp
	ai     *ai.Comp
}

func NewEnt(x, y, _, _ float64, props *core.Properties) *Ent {
	ent := &Ent{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: entWidth, H: entHeight},
		anim:       &anim.Comp{FilesName: entAnimFile, OX: entOffsetX, OY: entOffsetY, OXFlip: entOffsetFlip, FlipX: props.FlipX},
		body:       &body.Comp{MaxX: entMaxSpeed},
		hitbox:     &hitbox.Comp{},
		stats:      &stats.Comp{MaxHealth: entHealth, MaxPoise: entPoise, Exp: entExp},
		ai:         &ai.Comp{},
	}
	ent.Add(ent.anim, ent.body, ent.hitbox, ent.stats, ent.ai)
	ent.Control = actor.NewControl(ent)

	var view *bump.Rect
	if props.View != nil {
		viewRect := bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
		view = &viewRect
	}
	ent.ai.SetAct(func() { ent.aiScript(view) })

	return ent
}

func (g *Ent) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return g.anim, g.body, g.hitbox, g.stats, g.ai
}

func (g *Ent) Update(dt float64) { g.SimpleUpdate(dt) }

//nolint:mnd
func (g *Ent) aiScript(view *bump.Rect) {
	g.ai.Add(0, actor.IdleAction(g.Control, view))
	ai.Choices{
		{0.2, func() {
			g.ai.Add(0, actor.ApproachAction(g.Control, entSpeed, entMaxSpeed, 10))
			g.ai.Add(0.2, actor.WaitAction())
		}},
		{1, func() {
			g.ai.Add(0, actor.ApproachAction(g.Control, entSpeed, entMaxSpeed, 0))
			g.ai.Add(0.1, actor.WaitAction())
			ai.Choices{
				{2, func() { g.ai.Add(5, actor.AttackAction(g.Control, "Attack", entDamage)) }},
				{0.5, func() { g.ai.Add(1, actor.BackUpAction(g.Control, entSpeed, 0)) }},
				{1, func() { g.ai.Add(0.1, actor.WaitAction()) }},
			}.Play()
		}},
	}.Play()
}
