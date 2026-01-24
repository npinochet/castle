package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/entity/actor"
	"game/vars"
	"math"
)

const (
	vargWidth, vargHeight                    = 13, 16
	vargOffsetX, vargOffsetY, vargOffsetFlip = -3, -1, 10 // TODO: Tune
	vargSpeed, vargMaxSpeed                  = 60.0, 40
	vargHealth                               = 80
	vargDamage                               = 15
	vargExp                                  = 80
	vargPoise                                = 50
)

type Varg struct {
	*core.BaseEntity
	*actor.Control
	anim   *anim.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	stats  *stats.Comp
	ai     *ai.Comp
}

func NewVarg(x, y, _, _ float64, props *core.Properties) *Varg {
	varg := &Varg{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: vargWidth, H: vargHeight},
		anim: &anim.Comp{
			FilesName: "varg",
			OX:        vargOffsetX, OY: vargOffsetY,
			OXFlip: vargOffsetFlip,
		},
		body:   &body.Comp{},
		hitbox: &hitbox.Comp{},
		stats:  &stats.Comp{MaxHealth: vargHealth, MaxPoise: vargPoise, MaxStamina: vargPoise, Exp: vargExp},
		ai:     &ai.Comp{},
	}
	varg.Add(varg.anim, varg.body, varg.hitbox, varg.stats, varg.ai)
	varg.Control = actor.NewControl(varg)
	varg.ai.SetAct(func() { varg.ai.Add(10, varg.StalkAttack()) })

	return varg
}

func (r *Varg) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return r.anim, r.body, r.hitbox, r.stats, r.ai
}

func (r *Varg) Init() {
	r.body.Weight = 0
	r.Control.Init()
	r.ai.Target = vars.Player
}

func (v *Varg) Update(dt float64) {
	if v.stats.Health <= 0 {
		v.Die(dt)
	}
}

func (r *Varg) StalkAttack() *ai.Action {
	return &ai.Action{
		Name: "StalkAttack",
		Entry: func() {
			r.body.MaxX, r.body.MaxY = vargMaxSpeed, vargMaxSpeed
			r.anim.SetState(vars.WalkTag)
			mult := 0.0
			r.SetAttack(vargDamage, vargDamage, 0, 0, &mult)
		},
		Next: func(dt float64) bool {
			if r.ai.Target == nil || r.anim.State == vars.IdleTag {
				return true
			}
			compX, compY := r.targetAngleComps()
			r.body.Vx += compX * dt
			if _, ty, _, _ := r.ai.Target.Rect(); r.Y+r.H > ty {
				r.body.Vy -= vargSpeed * 2 * dt
			} else {
				r.body.Vy += compY * dt
			}

			return false
		},
	}
}

func (r *Varg) targetAngleComps() (float64, float64) {
	if r.ai.Target == nil {
		return 0, 0
	}
	tx, ty, tw, _ := r.ai.Target.Rect()
	angle := math.Atan2((ty)-(r.Y+r.H), (tx+tw/2)-(r.X+r.W/2))

	return vargSpeed * math.Cos(angle), vargSpeed * math.Sin(angle)
}
