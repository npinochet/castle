package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/libs/bump"
)

const (
	defaultAttackPushForce                    = 100
	defaultReactForce                         = 200
	deafultDamage, defaultStaminaDamage       = 20, 20
	defaultMaxXDiv, defaultMaxXRecoverRateDiv = 1.2, 2
)

func (a *Actor) IsActive() bool        { return a.Active }
func (a *Actor) SetActive(active bool) { a.Active = active }

type Actor struct {
	core.Entity
	Body                                *body.Comp
	Hitbox                              *hitbox.Comp
	Anim                                *anim.Comp
	Stats                               *stats.Comp
	AI                                  *ai.Comp
	speed                               float64
	Damage, StaminaDamage               float64
	BlockMaxXDiv, BlockRecoverRateDiv   float64
	ReactForce, AttackPushForce         float64
	blockMaxXSave, blockRecoverRateSave float64
}

func NewActor(
	x, y float64,
	body *body.Comp,
	animc *anim.Comp,
	stat *stats.Comp,
	damage, staminaDamage float64,
) *Actor {
	if stat == nil {
		stat = &stats.Comp{}
	}
	if animc.Fsm == nil {
		animc.Fsm = anim.DefaultFsm()
	}
	if damage == 0 {
		damage = deafultDamage
	}
	if staminaDamage == 0 {
		staminaDamage = defaultStaminaDamage
	}

	actor := &Actor{
		Entity: core.Entity{X: x, Y: y},
		Hitbox: &hitbox.Comp{},
		Body:   body, Anim: animc, Stats: stat,
		Damage: damage, StaminaDamage: staminaDamage,
		BlockMaxXDiv: defaultMaxXDiv, BlockRecoverRateDiv: defaultMaxXRecoverRateDiv,
		AttackPushForce: defaultAttackPushForce,
		ReactForce:      defaultReactForce,
	}
	actor.AddComponent(actor.Body, actor.Hitbox, actor.Anim, actor.Stats)
	actor.Hitbox.HurtFunc = func(otherHc *hitbox.Comp, col bump.Collision, damange float64) {
		if actor.Anim.State == anim.BlockTag {
			actor.ShieldDown()
		}
		actor.Hurt(*otherHc.EntX, damage, nil)
	}
	actor.Hitbox.BlockFunc = func(otherHc *hitbox.Comp, col bump.Collision, damange float64) {
		actor.Block(*otherHc.EntX, damange, nil)
	}

	return actor
}

func (a *Actor) ManageAnim() {
	// TODO: Make more general, maybe add speed in the mix.
	state := a.Anim.State
	a.Body.Friction = !(state == anim.WalkTag && a.Body.Vx != 0)
	a.Stats.SetActive(state != anim.AttackTag && state != anim.StaggerTag)

	if state == anim.IdleTag || state == anim.WalkTag {
		nextState := anim.IdleTag
		if a.Body.Vx != 0 {
			nextState = anim.WalkTag
		}
		a.Anim.SetState(nextState)
	}
}

func (a *Actor) Attack() {
	if a.Stats.Stamina <= 0 || a.Anim.State == anim.AttackTag {
		return
	}
	force := a.AttackPushForce
	if a.Anim.FlipX {
		force *= -1
	}
	a.Anim.SetState(anim.AttackTag)

	once := false
	onceReaction := false
	a.Anim.OnFrames(func(frame int) {
		if _, hitbox, _ := a.Anim.GetFrameHitboxes(); hitbox != nil {
			if !once {
				once = true
				a.Body.Vx += force
				a.Stats.AddStamina(-a.StaminaDamage)
			}
			if a.Hitbox.Hit(hitbox.X, hitbox.Y, hitbox.W, hitbox.H, a.Damage) {
				if !onceReaction {
					onceReaction = true
					// TODO: a.Stagger(force) when shield has too much defense?
					a.Body.Vx -= (force * 2) / float64(frame)
				}
			}
		}
	})
}

func (a *Actor) Stagger(force float64) {
	if a.Anim.State == anim.StaggerTag {
		return
	}
	a.Anim.SetState(anim.StaggerTag)
	a.Body.Vx = -force
}

func (a *Actor) ShieldUp() {
	if a.Anim.State == anim.BlockTag || a.Anim.State == anim.StaggerTag {
		return
	}
	a.Anim.SetState(anim.BlockTag)
	a.blockMaxXSave = a.Body.MaxX
	a.blockRecoverRateSave = a.Stats.StaminaRecoverRate
	a.Body.MaxX /= a.BlockMaxXDiv
	a.Stats.StaminaRecoverRate /= a.BlockRecoverRateDiv
	_, _, blockbox := a.Anim.GetFrameHitboxes()
	a.Hitbox.PushHitbox(blockbox.X, blockbox.Y, blockbox.W, blockbox.H, true)
}

func (a *Actor) ShieldDown() {
	if a.Anim.State != anim.BlockTag {
		return
	}
	a.Anim.SetState(anim.IdleTag)
	a.Body.MaxX = a.blockMaxXSave
	a.Stats.StaminaRecoverRate = a.blockRecoverRateSave
	a.Hitbox.PopHitbox()
}

func (a *Actor) Hurt(otherX float64, damage float64, poiseBreak func(force, damage float64)) {
	if poiseBreak == nil {
		poiseBreak = func(force, damage float64) {
			force *= damage / a.Stats.MaxHealth
			a.Stagger(force)
		}
	}
	a.Stats.AddPoise(-damage)
	a.Stats.AddHealth(-damage)

	force := a.ReactForce / 2
	if a.X > otherX {
		force *= -1
	}

	if a.Stats.Poise <= 0 {
		poiseBreak(a.ReactForce, damage)
	} else {
		a.Body.Vx -= force
	}
}

func (a *Actor) Block(otherX float64, damage float64, blockBreak func(force, damage float64)) {
	if blockBreak == nil {
		blockBreak = func(force, damage float64) {
			a.ShieldDown()
			force *= damage / a.Stats.MaxHealth
			a.Stagger(force)
			a.Anim.Data.PlaySpeed = 0.5 // double time stagger.
		}
	}

	a.Stats.AddStamina(-damage)

	force := a.ReactForce / 4
	if a.X > otherX {
		force *= -1
	}

	if a.Stats.Stamina < 0 {
		blockBreak(a.ReactForce, damage)
	} else {
		a.Body.Vx -= force
	}
}
