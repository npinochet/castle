package entity

import (
	"game/comp"
	"game/core"
	"game/libs/bump"
)

const (
	defaultAttackPushForce                    = 100
	defaultReactForce                         = 200
	deafultDamage, defaultStaminaDamage       = 20, 20
	defaultMaxXDiv, defaultMaxXRecoverRateDiv = 1.2, 2
	animIdleTag                               = "Idle"
	animWalkTag                               = "Walk"
	animAttackTag                             = "Attack"
	animBlockTag                              = "Block"
	animStaggerTag                            = "Stagger"
)

var DefaultAnimFsm = &comp.AnimFsm{
	Transitions: map[string]string{animWalkTag: animIdleTag, animAttackTag: animIdleTag, animStaggerTag: animIdleTag},
	ExitCallbacks: map[string]func(*comp.AsepriteComponent){
		animStaggerTag: func(ac *comp.AsepriteComponent) { ac.Data.PlaySpeed = 1 },
	},
}

func (a *Actor) IsActive() bool        { return a.Active }
func (a *Actor) SetActive(active bool) { a.Active = active }

type Actor struct {
	core.Entity
	Body                                *comp.BodyComponent
	Hitbox                              *comp.HitboxComponent
	Anim                                *comp.AsepriteComponent
	Stats                               *comp.StatsComponent
	AI                                  *comp.AIComponent
	Damage, StaminaDamage               float64
	BlockMaxXDiv, BlockRecoverRateDiv   float64
	ReactForce, AttackPushForce         float64
	blockMaxXSave, blockRecoverRateSave float64
}

func NewActor(
	x, y float64,
	body *comp.BodyComponent,
	anim *comp.AsepriteComponent,
	stats *comp.StatsComponent,
	damage, staminaDamage float64,
) *Actor {
	if stats == nil {
		stats = &comp.StatsComponent{}
	}
	if anim.Fsm == nil {
		anim.Fsm = DefaultAnimFsm
	}
	if damage == 0 {
		damage = deafultDamage
	}
	if staminaDamage == 0 {
		staminaDamage = defaultStaminaDamage
	}

	actor := &Actor{
		Entity: core.Entity{X: x, Y: y},
		Hitbox: &comp.HitboxComponent{},
		Body:   body, Anim: anim, Stats: stats,
		Damage: damage, StaminaDamage: staminaDamage,
		BlockMaxXDiv: defaultMaxXDiv, BlockRecoverRateDiv: defaultMaxXRecoverRateDiv,
		AttackPushForce: defaultAttackPushForce,
		ReactForce:      defaultReactForce,
	}
	actor.AddComponent(actor.Body, actor.Hitbox, actor.Anim, actor.Stats)
	actor.Hitbox.HurtFunc = func(otherHc *comp.HitboxComponent, col bump.Collision, damange float64) {
		if actor.Anim.State == animBlockTag {
			actor.ShieldDown()
		}
		actor.Hurt(*otherHc.EntX, damage, nil)
	}
	actor.Hitbox.BlockFunc = func(otherHc *comp.HitboxComponent, col bump.Collision, damange float64) {
		actor.Block(*otherHc.EntX, damange, nil)
	}

	return actor
}

func (a *Actor) ManageAnim() {
	// TODO: Make more general.
	state := a.Anim.State
	a.Body.Friction = !(state == animWalkTag && a.Body.Vx != 0)
	a.Stats.SetActive(state != animAttackTag && state != animStaggerTag)

	if state == animIdleTag || state == animWalkTag {
		nextState := animIdleTag
		if a.Body.Vx != 0 {
			nextState = animWalkTag
		}
		a.Anim.SetState(nextState)
	}
}

func (a *Actor) Attack() {
	if a.Stats.Stamina <= 0 || a.Anim.State == animAttackTag {
		return
	}
	force := a.AttackPushForce
	if a.Anim.FlipX {
		force *= -1
	}
	a.Anim.SetState(animAttackTag)

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
	if a.Anim.State == animStaggerTag {
		return
	}
	a.Anim.SetState(animStaggerTag)
	a.Body.Vx = -force
}

func (a *Actor) ShieldUp() {
	if a.Anim.State == animBlockTag {
		return
	}
	a.Anim.SetState(animBlockTag)
	a.blockMaxXSave = a.Body.MaxX
	a.blockRecoverRateSave = a.Stats.StaminaRecoverRate
	a.Body.MaxX /= a.BlockMaxXDiv
	a.Stats.StaminaRecoverRate /= a.BlockRecoverRateDiv
	_, _, blockbox := a.Anim.GetFrameHitboxes()
	a.Hitbox.PushHitbox(blockbox.X, blockbox.Y, blockbox.W, blockbox.H, true)
}

func (a *Actor) ShieldDown() {
	if a.Anim.State != animBlockTag {
		return
	}
	a.Anim.SetState(animIdleTag)
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
