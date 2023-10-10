package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/control"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/libs/bump"
)

const (
	defaultAttackPushForce                    = 100
	defaultReactForce                         = 100
	defaultMaxXDiv, defaultMaxXRecoverRateDiv = 2, 3
)

type ActorControl struct {
	Control *control.Comp
	Body    *body.Comp
	Hitbox  *hitbox.Comp
	Anim    *anim.Comp
	Stats   *stats.Comp
	AI      *ai.Comp
}

func (a *ActorControl) BindControl(entity *core.Entity) {
	a.Control = core.GetComponent[*control.Comp](entity)
	a.Body = core.GetComponent[*body.Comp](entity)
	a.Hitbox = core.GetComponent[*hitbox.Comp](entity)
	a.Anim = core.GetComponent[*anim.Comp](entity)
	a.Stats = core.GetComponent[*stats.Comp](entity)

	a.AI = core.GetComponent[*ai.Comp](entity)
}

func NewActorControl(x, y, w, h float64, attackTags []string, animc *anim.Comp, bodyc *body.Comp, stat *stats.Comp) *core.Entity {
	if bodyc == nil {
		bodyc = &body.Comp{}
	}
	if stat == nil {
		stat = &stats.Comp{}
	}
	if animc.Fsm == nil {
		animc.Fsm = anim.DefaultFsm()
	}
	hitboxc := &hitbox.Comp{}
	controlc := &control.Comp{
		BlockMaxXDiv: defaultMaxXDiv, BlockRecoverRateDiv: defaultMaxXRecoverRateDiv,
		AttackPushForce: defaultAttackPushForce,
		ReactForce:      defaultReactForce,
		AttackTags:      attackTags,
	}
	aic := &ai.Comp{}

	a := &core.Entity{X: x, Y: y, W: w, H: h}
	a.AddComponent(hitboxc, bodyc, animc, stat, controlc, aic)
	aic.Entity = a

	animc.Fsm.Exit[anim.ConsumeTag] = func(_ *anim.Comp) { controlc.ResetState(anim.IdleTag) }
	animc.Fsm.Exit[anim.ConsumeTag] = func(_ *anim.Comp) { controlc.ResetState(anim.IdleTag) }
	hitboxc.BlockFunc = func(other *core.Entity, col *bump.Collision, damage float64) { Block(a, other, damage, nil) }
	hitboxc.HurtFunc = func(other *core.Entity, col *bump.Collision, damage float64) { Hurt(a, other, damage, nil) }

	return a
}

func Hurt(a, other *core.Entity, damage float64, poiseBreak func(force, damage float64)) {
	controlc := core.GetComponent[*control.Comp](a)
	if controlc.Anim.State == anim.BlockTag || controlc.Anim.State == anim.ParryBlockTag {
		controlc.ShieldDown()
	}
	controlc.Stats.AddPoise(-damage)
	controlc.Stats.AddHealth(-damage)

	force := controlc.ReactForce / 2
	if controlc.X > other.X {
		force *= -1
	}

	if controlc.Stats.Poise <= 0 {
		if poiseBreak == nil {
			poiseBreak = func(force, damage float64) {
				force *= damage / controlc.Stats.MaxHealth
				controlc.Stagger(force)
			}
		}
		poiseBreak(controlc.ReactForce, damage)
	} else {
		if controlc.Anim.State == anim.ConsumeTag {
			controlc.Stagger(force)
		} else {
			controlc.Body.Vx -= force
		}
	}

	if aic := core.GetComponent[*ai.Comp](a); aic != nil {
		if aic != nil && aic.Target == nil {
			aic.Target = other
		}
	}
}

func Block(a, other *core.Entity, damage float64, blockBreak func(force, damage float64)) {
	controlc := core.GetComponent[*control.Comp](a)
	controlc.Stats.AddStamina(-damage)

	if controlc.Anim.State == anim.ParryBlockTag {
		return
	}

	force := controlc.ReactForce / 4
	if controlc.X > other.X {
		force *= -1
	}

	if controlc.Stats.Stamina < 0 {
		if blockBreak == nil {
			blockBreak = func(force, damage float64) {
				controlc.ShieldDown()
				force *= damage / controlc.Stats.MaxHealth
				controlc.Stagger(force)
				controlc.Anim.Data.PlaySpeed = 0.5 // double time stagger.
			}
		}
		blockBreak(controlc.ReactForce, damage)
	} else {
		controlc.Body.Vx -= force
	}
}
