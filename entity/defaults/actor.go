package defaults

import (
	"game/comps/ai"
	"game/comps/basic/anim"
	"game/comps/basic/body"
	"game/comps/basic/hitbox"
	"game/comps/basic/stats"
	"game/comps/control"
	"game/core"
	"game/libs/bump"
)

const (
	defaultAttackPushForce                    = 100
	defaultReactForce                         = 100
	defaultMaxXDiv, defaultMaxXRecoverRateDiv = 2, 3
)

type Actor struct {
	*core.Entity
	Control *control.Comp
	Body    *body.Comp
	Hitbox  *hitbox.Comp
	Anim    *anim.Comp
	Stats   *stats.Comp
	AI      *ai.Comp
}

func NewActor(x, y, w, h float64, attackTags []string) *Actor {
	a := &Actor{
		Entity: &core.Entity{X: x, Y: y, W: w, H: h},
		Control: &control.Comp{
			BlockMaxXDiv: defaultMaxXDiv, BlockRecoverRateDiv: defaultMaxXRecoverRateDiv,
			AttackPushForce: defaultAttackPushForce,
			ReactForce:      defaultReactForce,
			AttackTags:      attackTags,
		},
		Body:   &body.Comp{},
		Hitbox: &hitbox.Comp{},
		Anim:   &anim.Comp{Fsm: anim.DefaultFsm()},
		Stats:  &stats.Comp{},
		AI:     &ai.Comp{},
	}

	return a
}

func (a *Actor) SetupComponents() {
	a.AddComponent(a.Hitbox, a.Body, a.Anim, a.Stats, a.Control, a.AI)

	if a.Anim != nil && a.Anim.Fsm == nil {
		a.Anim.Fsm = anim.DefaultFsm()
		a.Anim.Fsm.Exit[anim.ConsumeTag] = func(_ *anim.Comp) { a.Control.ResetState(anim.IdleTag) }
	}
	a.Hitbox.BlockFunc = func(other *core.Entity, col *bump.Collision, damage float64) { Block(a.Entity, other, damage, nil) }
	a.Hitbox.HurtFunc = func(other *core.Entity, col *bump.Collision, damage float64) { Hurt(a.Entity, other, damage, nil) }
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
