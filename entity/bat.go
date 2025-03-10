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
	"math"
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
}

//nolint:mnd
func (r *Bat) aiScript(view *bump.Rect) {
	idle := actor.IdleAction(r.Control, view)
	idle.Exit = func() { r.awake = true }
	r.ai.Add(0, idle)
	ai.Choices{
		{0.3, func() { r.ai.Add(3, r.Stalk(20)) }},
		{0.3, func() { r.ai.Add(5, r.Stalk(25)) }},
		{0.3, func() { r.ai.Add(4, r.Stalk(28)) }},
	}.Play()
	r.ai.Add(5, r.Attack())
}

func (r *Bat) Attack() *ai.Action {
	// TODO: If the bat hits or reaches it's target, it end the attack
	action := actor.AttackAction(r.Control, "Attack", batDamage)
	originalEntry := action.Entry
	originalNext := action.Next
	action.Entry = func() {
		originalEntry()
		r.anim.OnFrame(1, func() { r.body.Vx, r.body.Vy = r.targetAngleComps() })
	}
	action.Next = func(dt float64) bool {
		if r.ai.Target == nil {
			return true
		}

		compX, compY := r.targetAngleComps()
		r.body.Vx += compX * dt
		r.body.Vy += compY * dt

		tx, _, tw, _ := r.ai.Target.Rect()
		r.anim.FlipX = tx+tw/2 > r.X+r.W/2

		return originalNext(dt)
	}

	return action
}

func (r *Bat) Stalk(rangeAdjustment float64) *ai.Action {
	return &ai.Action{
		Name:  "Stalk",
		Entry: func() { r.body.MaxX, r.body.MaxY = batMaxSpeed, batMaxSpeed },
		Next: func(dt float64) bool {
			if r.ai.Target == nil {
				return true
			}
			if r.PausingState() {
				return false
			}

			compX, compY := r.targetAngleComps()
			backUp := r.ai.InTargetRange(0, rangeAdjustment)
			if backUp {
				r.body.Vx -= compX * dt
				r.body.Vy -= (batSpeed / 4) * dt
			} else {
				r.body.Vx += compX * dt
				if _, ty, _, _ := r.ai.Target.Rect(); r.Y+r.H > ty {
					r.body.Vy -= batSpeed * 2 * dt
				} else {
					r.body.Vy += compY * dt
				}
			}

			return false
		},
	}
}

func (r *Bat) targetAngleComps() (float64, float64) {
	if r.ai.Target == nil {
		return 0, 0
	}
	tx, ty, tw, _ := r.ai.Target.Rect()
	angle := math.Atan2((ty)-(r.Y+r.H), (tx+tw/2)-(r.X+r.W/2))

	return batSpeed * math.Cos(angle), batSpeed * math.Sin(angle)
}
