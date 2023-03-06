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
	defaultMaxXDiv, defaultMaxXRecoverRateDiv = 2, 3
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
	weightSave                        float64
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

	a := &Actor{
		Entity: core.Entity{X: x, Y: y, W: w, H: h},
		Hitbox: &hitbox.Comp{},
		Body:   bodyc, Anim: animc, Stats: stat,
		BlockMaxXDiv: defaultMaxXDiv, BlockRecoverRateDiv: defaultMaxXRecoverRateDiv,
		AttackPushForce: defaultAttackPushForce,
		ReactForce:      defaultReactForce,
		AttackTags:      attackTags,
	}
	a.AddComponent(a.Body, a.Hitbox, a.Anim, a.Stats)
	a.Anim.Fsm.Exit[anim.ConsumeTag] = func(_ *anim.Comp) { a.ResetState(anim.IdleTag) }
	a.Hitbox.HurtFunc = func(otherHc *hitbox.Comp, col *bump.Collision, damage float64) {
		if a.Anim.State == anim.BlockTag || a.Anim.State == anim.ParryBlockTag {
			a.ShieldDown()
		}
		a.Hurt(otherHc.Entity, damage, nil)
	}
	a.Hitbox.BlockFunc = func(otherHc *hitbox.Comp, col *bump.Collision, damange float64) {
		a.Block(otherHc.Entity, damange, nil)
	}

	return a
}

func (a *Actor) ManageAnim() {
	// TODO: Make more general, maybe add speed in the mix
	if state := a.Anim.State; state == anim.IdleTag || state == anim.WalkTag {
		nextState := anim.IdleTag
		if a.Body.Vx != 0 {
			nextState = anim.WalkTag
		}
		a.Anim.SetState(nextState)
	}
	a.Stats.Pause = a.pausedState()
	a.Hitbox.ParryBlock = a.Anim.State == anim.ParryBlockTag
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
	a.Stats.StaminaRecoverRateMultiplier = 0
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

	var contacted []*hitbox.Comp
	var once, blockForceOnce, blocked bool
	a.Anim.OnFrames(func(frame int) {
		hitbox, err := a.Anim.GetFrameHitbox(anim.HitboxSliceName)
		if err != nil {
			contacted = nil

			return
		}

		blocked, contacted = a.Hitbox.HitFromHitBox(hitbox, damage, contacted)
		if blocked && !blockForceOnce {
			blockForceOnce = true
			blockForce := force
			if !once {
				blockForce *= 2
			}
			a.Body.Vx -= blockForce
			for _, other := range contacted {
				// TODO: If other state == parryblock then take poise damage
				// Find a better way to refactor this behaviour
				if other.ParryBlock {
					a.Stats.AddPoise(-damage)
					if a.Stats.Poise <= 0 {
						force := a.ReactForce / 2
						if a.X > other.Entity.X {
							force *= -1
						}
						a.Stagger(force * (damage / a.Stats.MaxHealth))
					}

					break
				}
			}
		}
		if !once {
			a.Stats.AddStamina(-staminaDamage)
			a.Body.Vx += force
			once = true
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
	a.weightSave = a.Body.Weight
	a.Body.Weight = -1
}

func (a *Actor) ClimbOff() {
	if a.Anim.State != anim.ClimbTag {
		return
	}
	a.Body.ClipLadder = false
	a.Body.Weight = a.weightSave
	a.Anim.Data.PlaySpeed = 1
	a.Anim.SetState(anim.IdleTag)
}

func (a *Actor) Heal(effectFrame int, amount float64) {
	if a.pausedState() || a.Stats.Heal <= 0 || !a.Body.Ground {
		return
	}
	a.ResetState(anim.ConsumeTag)
	a.Body.MaxXMultiplier = -1
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
	if a.pausedState() || a.Anim.State == anim.BlockTag || a.Anim.State == anim.ParryBlockTag || a.Stats.Stamina <= 0 {
		return
	}
	a.ResetState(anim.ParryBlockTag)
	a.Body.MaxXMultiplier = -1 / a.BlockMaxXDiv
	a.Stats.StaminaRecoverRateMultiplier = -1 / a.BlockRecoverRateDiv
	blockbox, err := a.Anim.GetFrameHitbox(anim.BlockSliceName)
	if err != nil {
		panic(err)
	}
	a.Hitbox.PushHitbox(blockbox, true)
}

func (a *Actor) ShieldDown() {
	if a.Anim.State != anim.BlockTag && a.Anim.State != anim.ParryBlockTag {
		return
	}
	a.Anim.SetState(anim.IdleTag)
	a.Body.MaxXMultiplier = 0
	a.Stats.StaminaRecoverRateMultiplier = 0
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

	if a.Anim.State == anim.ParryBlockTag {
		return
	}

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
