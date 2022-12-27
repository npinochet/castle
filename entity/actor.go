package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/libs/bump"
	"game/utils"
)

const (
	defaultAttackPushForce                    = 100
	defaultReactForce                         = 200
	deafultDamage, defaultStaminaDamage       = 40, 20
	defaultMaxXDiv, defaultMaxXRecoverRateDiv = 1.2, 2
)

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

func (a *Actor) GetBody() *body.Comp          { return a.Body }
func (a *Actor) GetHitbox() *hitbox.Comp      { return a.Hitbox }
func (a *Actor) GetAnim() *anim.Comp          { return a.Anim }
func (a *Actor) GetStats() *stats.Comp        { return a.Stats }
func (a *Actor) GetAI() *ai.Comp              { return a.AI }
func (a *Actor) SetSpeed(speed, maxX float64) { a.speed, a.Body.MaxX = speed, maxX }

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
	actor.Hitbox.HurtFunc = func(otherHc *hitbox.Comp, col *bump.Collision, damange float64) {
		if actor.Anim.State == anim.BlockTag {
			actor.ShieldDown()
		}
		actor.Hurt(otherHc.Entity.X, damage, nil)
	}
	actor.Hitbox.BlockFunc = func(otherHc *hitbox.Comp, col *bump.Collision, damange float64) {
		actor.Block(otherHc.Entity.X, damange, nil)
	}

	return actor
}

func (a *Actor) ManageAnim(attackTags []string) {
	// TODO: Make more general, maybe add speed in the mix.
	state := a.Anim.State
	a.Body.Friction = !(state == anim.WalkTag && a.Body.Vx != 0)
	a.Stats.Pause = utils.Contains(attackTags, state) || state == anim.StaggerTag

	if state == anim.IdleTag || state == anim.WalkTag {
		nextState := anim.IdleTag
		if a.Body.Vx != 0 {
			nextState = anim.WalkTag
		}
		a.Anim.SetState(nextState)
	}
}

func (a *Actor) Attack(attackTag string) {
	if a.Stats.Stamina <= 0 || a.Anim.State == attackTag {
		return
	}
	force := a.AttackPushForce
	if a.Anim.FlipX {
		force *= -1
	}
	a.Anim.SetState(attackTag)

	once := false
	var contacted []*hitbox.Comp
	a.Anim.OnFrames(func(frame int) {
		if hitbox, err := a.Anim.GetFrameHitbox(anim.HitboxSliceName); err == nil {
			if !once {
				once = true
				a.Body.Vx += force
				a.Stats.AddStamina(-a.StaminaDamage)
			}
			var blocked bool
			blocked, contacted = a.Hitbox.HitFromHitBox(hitbox, a.Damage, contacted)
			if blocked {
				a.Anim.OnFrames(nil)
				// TODO: a.Stagger(force) when shield has too much defense?
				a.Body.Vx -= (force * 2) / float64(frame) // TODO: why divide by frame?
			}
		}
	})
}

func (a *Actor) Stagger(force float64) {
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
	blockbox, err := a.Anim.GetFrameHitbox(anim.BlockSliceName)
	if err != nil {
		panic(err)
	}
	a.Hitbox.PushHitbox(blockbox, true)
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
