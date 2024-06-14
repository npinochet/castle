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
	"game/vars"
	"time"
)

const (
	skelemanAnimFile                                     = "skeleman"
	skelemanWidth, skelemanHeight                        = 8, 12
	skelemanOffsetX, skelemanOffsetY, skelemanOffsetFlip = -12, -5, 20
	skelemanSpeed, skelemanMaxSpeed                      = 100, 60
	skelemanHealth                                       = 110
	skelemanDamage                                       = 18
	skelemanExp                                          = 25
	skelemanPoise                                        = 30
)

type Skeleman struct {
	*core.BaseEntity
	*actor.Control
	anim   *anim.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	stats  *stats.Comp
	ai     *ai.Comp
}

func NewSkeleman(x, y, _, _ float64, props *core.Properties) *Skeleman {
	skeleman := &Skeleman{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: skelemanWidth, H: skelemanHeight},
		anim: &anim.Comp{
			FilesName: skelemanAnimFile,
			OX:        skelemanOffsetX, OY: skelemanOffsetY,
			OXFlip: skelemanOffsetFlip,
			FlipX:  props.FlipX,
		},
		body:   &body.Comp{},
		hitbox: &hitbox.Comp{},
		stats:  &stats.Comp{MaxHealth: skelemanHealth, MaxPoise: skelemanPoise, Exp: skelemanExp},
		ai:     &ai.Comp{},
	}
	skeleman.Add(skeleman.anim, skeleman.body, skeleman.hitbox, skeleman.stats, skeleman.ai)
	skeleman.Control = actor.NewControl(skeleman)

	var view *bump.Rect
	if props.View != nil {
		viewRect := bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
		view = &viewRect
	}
	skeleman.ai.SetAct(func() { skeleman.aiScript(view) })

	return skeleman
}

func (s *Skeleman) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return s.anim, s.body, s.hitbox, s.stats, s.ai
}

func (s *Skeleman) Update(dt float64) {
	s.SimpleUpdate(dt)
}

func (s *Skeleman) jumpAttackAction() *ai.Action {
	speed := float64(skelemanSpeed)

	return &ai.Action{
		Name: "JumpAttack",
		Entry: func() {
			if s.PausingState() {
				return
			}
			s.body.MaxX = skelemanMaxSpeed * 2
			time.AfterFunc(1*time.Millisecond, func() { s.Control.Attack("AttackShort", skelemanDamage, 0, 10, 10) })
			s.body.Vy = -skelemanSpeed
			s.body.Ground = false
			if s.anim.FlipX {
				s.body.Vx += skelemanMaxSpeed * 2
			} else {
				s.body.Vx -= skelemanMaxSpeed * 2
				speed *= -1
			}
		},
		Next: func(dt float64) bool {
			if !s.PausingState() {
				s.body.Vx += speed * dt
			}

			return s.body.Ground && s.anim.State != "AttackShort"
		},
		Exit: func() { s.body.MaxX = skelemanMaxSpeed },
	}
}

// nolint: nolintlint, gomnd
func (s *Skeleman) aiScript(view *bump.Rect) {
	s.ai.Add(0, actor.IdleAction(s.Control, view))
	s.ai.Add(0, actor.ApproachAction(s.Control, skelemanSpeed, vars.DefaultMaxX))
	s.ai.Add(0.1, actor.WaitAction())

	ai.Choices{
		{2, func() { s.ai.Add(5, actor.AttackAction(s.Control, "AttackShort", skelemanDamage)) }},
		{2, func() { s.ai.Add(5, actor.AttackAction(s.Control, "AttackLong", skelemanDamage)) }},
		{1, func() { s.ai.Add(10, s.jumpAttackAction()) }},
		{0.5, func() { s.ai.Add(1, actor.BackUpAction(s.Control, skelemanSpeed, 0)) }},
		{1, func() { s.ai.Add(1, actor.WaitAction()) }},
	}.Play()
}
