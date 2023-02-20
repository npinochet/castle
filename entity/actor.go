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
	defaultMaxXDiv, defaultMaxXRecoverRateDiv = 1.2, 2
)

type Actor struct {
	core.Entity
	Body                              *body.Comp
	Hitbox                            *hitbox.Comp
	Anim                              *anim.Comp
	Stats                             *stats.Comp
	AI                                *ai.Comp
	AttackTags                        []string
	Speed                             float64
	BlockMaxXDiv, BlockRecoverRateDiv float64
	ReactForce, AttackPushForce       float64
	blockRecoverRateSave              float64
}

func (a *Actor) GetBody() *body.Comp          { return a.Body }
func (a *Actor) GetHitbox() *hitbox.Comp      { return a.Hitbox }
func (a *Actor) GetAnim() *anim.Comp          { return a.Anim }
func (a *Actor) GetStats() *stats.Comp        { return a.Stats }
func (a *Actor) GetAI() *ai.Comp              { return a.AI }
func (a *Actor) SetSpeed(speed, maxX float64) { a.Speed, a.Body.MaxX = speed, maxX }

func NewActor(x, y, w, h float64, attackTags []string, animc *anim.Comp, bodyc *body.Comp, stat *stats.Comp) *Actor {
	if bodyc == nil {
		bodyc = &body.Comp{}
	}
	if stat == nil {
		stat = &stats.Comp{}
	}
	if animc.Fsm == nil {
		animc.Fsm = anim.DefaultFsm()
	}

	actor := &Actor{
		Entity: core.Entity{X: x, Y: y, W: w, H: h},
		Hitbox: &hitbox.Comp{},
		Body:   bodyc, Anim: animc, Stats: stat,
		BlockMaxXDiv: defaultMaxXDiv, BlockRecoverRateDiv: defaultMaxXRecoverRateDiv,
		AttackPushForce: defaultAttackPushForce,
		ReactForce:      defaultReactForce,
		AttackTags:      attackTags,
	}
	actor.AddComponent(actor.Body, actor.Hitbox, actor.Anim, actor.Stats)
	actor.Hitbox.HurtFunc = func(otherHc *hitbox.Comp, col *bump.Collision, damage float64) {
		if actor.Anim.State == anim.BlockTag {
			actor.ShieldDown()
		}
		actor.Hurt(otherHc.Entity, damage, nil)
	}
	actor.Hitbox.BlockFunc = func(otherHc *hitbox.Comp, col *bump.Collision, damange float64) {
		actor.Block(otherHc.Entity, damange, nil)
	}

	return actor
}

func (a *Actor) ManageAnim() {
	// TODO: Make more general, maybe add speed in the mix
	state := a.Anim.State
	a.Body.Friction = !(state == anim.WalkTag && a.Body.Vx != 0)
	a.Stats.Pause = a.pausedState()

	if state == anim.IdleTag || state == anim.WalkTag {
		nextState := anim.IdleTag
		if a.Body.Vx != 0 {
			nextState = anim.WalkTag
		}
		a.Anim.SetState(nextState)
	}
}

func (a *Actor) SimpleUpdate(dt float64) {
	a.ManageAnim()
	if target := a.AI.Target; target != nil {
		if a.Anim.State == anim.WalkTag || a.Anim.State == anim.IdleTag {
			a.Anim.FlipX = target.X > a.X
		}
	}

	if !a.pausedState() {
		if a.Anim.FlipX {
			a.Body.Vx += a.Speed * dt
		} else {
			a.Body.Vx -= a.Speed * dt
		}
	}

	if a.Stats.Health <= 0 {
		a.Remove()
	}
}

func (a *Actor) pausedState() bool {
	return utils.Contains(append(a.AttackTags, anim.StaggerTag, anim.ConsumeTag), a.Anim.State)
}

func (a *Actor) ResetState(state string) {
	a.ClimbOff()
	a.ShieldDown()
	a.Body.MaxXMultiplier = 0
	a.Anim.SetState(state)
}

func (a *Actor) Attack(attackTag string, damage, staminaDamage float64) {
	if a.pausedState() || a.Stats.Stamina <= 0 {
		return
	}
	force := -a.AttackPushForce
	if a.Anim.FlipX {
		force *= -1
	}
	a.ResetState(attackTag)

	once := false
	var contacted []*hitbox.Comp
	a.Anim.OnFrames(func(frame int) {
		if hitbox, err := a.Anim.GetFrameHitbox(anim.HitboxSliceName); err == nil {
			if !once {
				once = true
				a.Body.Vx += force
				a.Stats.AddStamina(-staminaDamage)
			}
			var blocked bool
			blocked, contacted = a.Hitbox.HitFromHitBox(hitbox, damage, contacted)
			if blocked {
				// a.Anim.OnFrames(nil) // TODO: This negates skeleman long attack, can't be blocked multiple times. This can be solved maybe with next TODO
				// TODO: a.Stagger(force) when shield has too much defense?
				a.Body.Vx -= (force * 2) / float64(frame) // TODO: why divide by frame?
			}
		} else {
			// TODO: Hitbox does not work with skeleman combo attack, this fixes it, it feels really hacky
			contacted = nil
		}
	})
}

func (a *Actor) ClimbOn(goingDown bool) {
	if a.pausedState() {
		return
	}
	a.Body.ClipLadder = a.Body.ClipLadder || goingDown
	if !a.Body.OnLadder || a.Anim.State == anim.ClimbTag {
		return
	}
	a.ResetState(anim.ClimbTag)
	a.Body.Weight = -1 // TODO:add weight save?? uuuhhh
}

func (a *Actor) ClimbOff() {
	if a.Anim.State != anim.ClimbTag {
		return
	}
	a.Body.ClipLadder = false
	a.Body.Weight = 0
	a.Anim.Data.PlaySpeed = 1
	a.Anim.SetState(anim.IdleTag)
}

func (a *Actor) Heal(effectFrame int, amount float64) {
	if a.pausedState() || a.Stats.Heal <= 0 {
		return
	}
	a.ResetState(anim.ConsumeTag)
	a.Body.MaxXMultiplier = (1 / a.BlockMaxXDiv) - 1
	a.Anim.OnFrames(func(frame int) {
		if frame == effectFrame {
			a.Anim.OnFrames(nil)
			a.Stats.AddHeal(-1)
			a.Stats.AddHealth(amount)
		}
	})
}

func (a *Actor) Stagger(force float64) {
	a.ResetState(anim.StaggerTag)
	a.Body.Vx = -force
}

func (a *Actor) ShieldUp() {
	if a.pausedState() || a.Anim.State == anim.BlockTag {
		return
	}
	a.ResetState(anim.BlockTag)
	a.Body.MaxXMultiplier = (1 / a.BlockMaxXDiv) - 1
	a.blockRecoverRateSave = a.Stats.StaminaRecoverRate
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
	a.Body.MaxXMultiplier = 0
	a.Stats.StaminaRecoverRate = a.blockRecoverRateSave
	a.Hitbox.PopHitbox()
}

func (a *Actor) Hurt(other *core.Entity, damage float64, poiseBreak func(force, damage float64)) {
	a.Stats.AddPoise(-damage)
	a.Stats.AddHealth(-damage)

	force := a.ReactForce / 2
	if a.X > other.X {
		force *= -1
	}

	if a.Stats.Poise <= 0 {
		if poiseBreak == nil {
			poiseBreak = func(force, damage float64) {
				force *= damage / a.Stats.MaxHealth
				a.Stagger(force)
			}
		}
		poiseBreak(a.ReactForce, damage)
	} else {
		if a.Anim.State == anim.ConsumeTag {
			a.Stagger(force)
		} else {
			a.Body.Vx -= force
		}
	}

	if a.AI != nil && a.AI.Target == nil {
		a.AI.Target = other
	}
}

func (a *Actor) Block(other *core.Entity, damage float64, blockBreak func(force, damage float64)) {
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
	if a.X > other.X {
		force *= -1
	}

	if a.Stats.Stamina < 0 {
		blockBreak(a.ReactForce, damage)
	} else {
		a.Body.Vx -= force
	}
}

func (a *Actor) Remove() {
	if a.Body != nil {
		a.World.Space.Remove(a.Body)
	}
	if a.Hitbox != nil {
		for a.Hitbox.PopHitbox() != nil {
		}
	}
	a.World.RemoveEntity(a.ID)
}
