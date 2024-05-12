package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/gated"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/entity/actor"
	"game/libs/bump"
	"game/vars"
)

const (
	knightAnimFile                                 = "knight"
	knightWidth, knightHeight                      = 8, 11
	knightOffsetX, knightOffsetY, knightOffsetFlip = -10, -3, 17
	knightMaxPosie                                 = 25
	knightExp                                      = 20
)

type Knight struct {
	*core.BaseEntity
	*actor.Control
	anim   *anim.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	stats  *stats.Comp
	ai     *ai.Comp
	gates  *gated.Comp
}

func NewKnight(x, y, _, _ float64, props *core.Properties) *Knight {
	knight := &Knight{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: knightWidth, H: knightHeight},
		anim: &anim.Comp{
			FilesName: knightAnimFile,
			OX:        knightOffsetX, OY: knightOffsetY,
			OXFlip: knightOffsetFlip,
			FlipX:  props.FlipX,
		},
		body:   &body.Comp{},
		hitbox: &hitbox.Comp{},
		stats:  &stats.Comp{MaxPoise: knightMaxPosie, Exp: knightExp},
		ai:     &ai.Comp{},
		gates:  &gated.Comp{Props: props.Custom},
	}
	knight.Add(knight.anim, knight.body, knight.hitbox, knight.stats, knight.ai, knight.gates)
	knight.Control = actor.NewControl(knight)

	var view *bump.Rect
	if props.View != nil {
		viewRect := bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
		view = &viewRect
	}
	knight.ai.SetAct(func() { knight.aiScript(view) })

	return knight
}

func (k *Knight) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return k.anim, k.body, k.hitbox, k.stats, k.ai
}

func (k *Knight) Update(dt float64) {
	if k.stats.Health <= 0 {
		k.gates.Open()
	}
	k.SimpleUpdate(dt)
}

// nolint: nolintlint, gomnd
func (k *Knight) aiScript(view *bump.Rect) {
	speed := 100.0
	k.ai.Add(0, actor.IdleAction(k.Control, view))
	k.ai.Add(1, actor.EntryAction(func() { k.gates.Close() }))
	k.ai.Add(0, actor.ApproachAction(k.Control, speed, vars.DefaultMaxX))
	k.ai.Add(0.1, actor.WaitAction())

	ai.Choice{
		{2, func() { k.ai.Add(5, actor.AttackAction(k.Control, "Attack", playerDamage)) }},
		{0.5, func() { k.ai.Add(1, actor.BackUpAction(k.Control, speed, 0)) }},
		{1, func() { k.ai.Add(1, actor.WaitAction()) }},
		{1, func() { k.ai.Add(1, actor.ShieldAction(k.Control)) }},
	}.Play()
}
