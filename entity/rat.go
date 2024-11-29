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
)

const (
	ratWidth, ratHeight                   = 11, 7
	ratOffsetX, ratOffsetY, ratOffsetFlip = -13, -17, 21
	ratSpeed, ratMaxSpeed                 = 60.0, 40
	ratWeight                             = 0.6
	ratHealth                             = 25
	ratDamage                             = 15
	ratExp                                = 15
	ratPoise                              = 10
)

type Rat struct {
	*core.BaseEntity
	*actor.Control
	anim      *anim.Comp
	body      *body.Comp
	hitbox    *hitbox.Comp
	stats     *stats.Comp
	ai        *ai.Comp
	jumpFrame int
}

func NewRat(x, y, _, _ float64, props *core.Properties) *Rat {
	rat := &Rat{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: ratWidth, H: ratHeight},
		anim: &anim.Comp{
			FilesName: "rat",
			OX:        ratOffsetX, OY: ratOffsetY,
			OXFlip: ratOffsetFlip,
			FlipX:  props.FlipX,
		},
		body:   &body.Comp{Weight: ratWeight},
		hitbox: &hitbox.Comp{},
		stats:  &stats.Comp{MaxHealth: ratHealth, MaxPoise: ratPoise, Exp: ratExp},
		ai:     &ai.Comp{},
	}
	rat.Add(rat.anim, rat.body, rat.hitbox, rat.stats, rat.ai)
	rat.Control = actor.NewControl(rat)

	var view *bump.Rect
	if props.View != nil {
		viewRect := bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
		view = &viewRect
	}
	rat.ai.SetAct(func() { rat.aiScript(view) })

	return rat
}

func (r *Rat) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return r.anim, r.body, r.hitbox, r.stats, r.ai
}

func (r *Rat) Init() {
	r.Control.Init()
	r.jumpFrame = r.anim.Data.Animation("Jump").From - r.anim.Data.Animation("Attack").From
}

func (r *Rat) Update(dt float64) {
	r.SimpleUpdate(dt)
	if !r.body.Ground && (r.anim.State == vars.IdleTag || r.anim.State == vars.WalkTag) {
		r.anim.SetState("Jump")
	}
}

func (r *Rat) jumpAttackAction() *ai.Action {
	speed := ratSpeed
	jumped, lifted := false, false

	return &ai.Action{
		Name: "JumpAttack",
		Entry: func() {
			if r.PausingState() {
				return
			}
			r.Control.Attack("Attack", ratDamage, ratDamage, 10, 10)
			r.anim.OnFrame(r.jumpFrame, func() {
				jumped = true
				r.body.MaxX = ratMaxSpeed * 2
				r.body.Vy = -ratSpeed
				r.body.Ground = false
				if r.anim.FlipX {
					r.body.Vx += ratMaxSpeed * 2
				} else {
					r.body.Vx -= ratMaxSpeed * 2
					speed *= -1
				}
			})
		},
		Next: func(dt float64) bool {
			if !r.body.Ground {
				lifted = true
			}
			if !jumped || !lifted {
				return false
			}
			if !r.PausingState() {
				r.body.Vx += speed * dt
			}
			if r.body.Ground {
				r.anim.SetState(vars.IdleTag)
			}

			return r.anim.State != "Attack"
		},
		Exit: func() { r.body.MaxX = ratMaxSpeed },
	}
}

//nolint:mnd
func (r *Rat) aiScript(view *bump.Rect) {
	rangeAdjustment := 10.0
	r.ai.Add(0, actor.IdleAction(r.Control, view))
	r.ai.Add(0, actor.ApproachAction(r.Control, ratSpeed, ratMaxSpeed, rangeAdjustment))
	r.ai.Add(0.1, actor.WaitAction())

	ai.Choices{
		{2, func() {
			r.ai.Add(10, r.jumpAttackAction())
			r.ai.Add(0.5, actor.WaitAction())
		}},
		{0.5, func() { r.ai.Add(1, actor.BackUpAction(r.Control, ratSpeed, -rangeAdjustment)) }},
		{1, func() { r.ai.Add(0.5, actor.WaitAction()) }},
	}.Play()
}
