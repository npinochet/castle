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
	batWidth, batHeight                   = 5, 5
	batOffsetX, batOffsetY, batOffsetFlip = -10, -9, 9
	batSpeed, batMaxSpeed                 = 60.0, 40
	batHealth                             = 25
	batDamage                             = 15
	batExp                                = 15
	batPoise                              = 10
)

type Bat struct {
	*core.BaseEntity
	*actor.Control
	anim   *anim.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	stats  *stats.Comp
	ai     *ai.Comp
	awake  bool
}

// TODO: Bats do not attack yet
func NewBat(x, y, _, _ float64, props *core.Properties) *Bat {
	bat := &Bat{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: batWidth, H: batHeight},
		anim: &anim.Comp{
			FilesName: "bat",
			OX:        batOffsetX, OY: batOffsetY,
			OXFlip: batOffsetFlip,
			FlipX:  props.FlipX,
		},
		body:   &body.Comp{},
		hitbox: &hitbox.Comp{},
		stats:  &stats.Comp{MaxHealth: batHealth, MaxPoise: batPoise, Exp: batExp},
		ai:     &ai.Comp{},
	}
	bat.Add(bat.anim, bat.body, bat.hitbox, bat.stats, bat.ai)
	bat.Control = actor.NewControl(bat)

	var view *bump.Rect
	if props.View != nil {
		viewRect := bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
		view = &viewRect
	}
	bat.ai.SetAct(func() { bat.aiScript(view) })

	return bat
}

func (r *Bat) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return r.anim, r.body, r.hitbox, r.stats, r.ai
}

func (r *Bat) Init() {
	r.body.Weight = 0
	r.Control.Init()
}

func (r *Bat) Update(dt float64) {
	if !r.awake {
		return
	}
	if r.SimpleUpdate(dt); r.awake && r.anim.State == vars.IdleTag {
		r.anim.SetState(vars.WalkTag)
	}
	// TODO: make better movement AI
	if r.ai != nil && r.ai.Target != nil {
		_, ty, _, th := r.ai.Target.Rect()
		_, y, _, h := r.Rect()
		if ty+th/2 > y+h/2 {
			r.body.Vy += batSpeed * dt
		} else {
			r.body.Vy -= batSpeed * dt
		}
	}
}

//nolint:mnd
func (r *Bat) aiScript(view *bump.Rect) {
	rangeAdjustment := 10.0
	idle := actor.IdleAction(r.Control, view)
	idle.Exit = func() { r.awake = true }
	r.ai.Add(0, idle)
	r.ai.Add(0, actor.ApproachAction(r.Control, batSpeed, batMaxSpeed, rangeAdjustment))
	r.ai.Add(0.1, actor.WaitAction())

	ai.Choices{
		{0.5, func() { r.ai.Add(1, actor.BackUpAction(r.Control, batSpeed, -rangeAdjustment)) }},
		{1, func() { r.ai.Add(0.5, actor.WaitAction()) }},
	}.Play()
}
