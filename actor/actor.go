package actor

import (
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
	core.CoreEntity
	Anim
	Body
	Hitbox
	Stats
	AI
	AttackTags                        []string
	Speed                             float64
	ParryBlock                        bool
	BlockMaxXDiv, BlockRecoverRateDiv float64
	ReactForce, AttackPushForce       float64
	weightSave                        float64
}

func (a *Actor) GetEntity() *core.CoreEntity { return &a.CoreEntity }

func NewActor(x, y, w, h float64, attackTags []string) Actor {
	return Actor{
		CoreEntity:          core.CoreEntity{X: x, Y: y, W: w, H: h},
		BlockMaxXDiv:        defaultMaxXDiv,
		BlockRecoverRateDiv: defaultMaxXRecoverRateDiv,
		AttackPushForce:     defaultAttackPushForce,
		ReactForce:          defaultReactForce,
		AttackTags:          attackTags,
	}
}

func (a *Actor) Init() {
	a.Anim.Init(a)
	a.Anim.Fsm.Exit[ConsumeTag] = func(actor *Actor) { actor.ResetState(IdleTag) }
	a.Body.Init(a)
	a.Hitbox.Init(a)
	a.Hitbox.HurtFunc = func(other *Actor, col *bump.Collision, damage float64) {
		if a.Anim.State == BlockTag || a.Anim.State == ParryBlockTag {
			a.ShieldDown()
		}
		a.Hurt(other, damage, nil)
	}
	a.Hitbox.BlockFunc = func(other *Actor, col *bump.Collision, damage float64) {
		a.Block(other, damage, nil)
	}
	hurtbox, err := a.Anim.GetFrameSlice(AHurtboxSliceName)
	if err != nil {
		panic(err)
	}
	a.Hitbox.PushHitbox(a, hurtbox, false)
	a.Stats.Init()
	a.AI.Init(a)
}

func (a *Actor) Update(dt float64) {
	a.Anim.Update(a, dt)
	a.Body.Update(a, dt)
	a.Hitbox.Update(a)
	a.Stats.Update(dt)
	a.AI.Update(dt)
}

func (a *Actor) ManageAnim() {
	// TODO: Make more general, maybe add speed in the mix
	if state := a.Anim.State; state == IdleTag || state == WalkTag {
		nextState := IdleTag
		if a.Body.Vx != 0 {
			nextState = WalkTag
		}
		a.Anim.SetState(a, nextState)
	}
	a.Stats.Pause = a.PausedState()
	a.ParryBlock = a.Anim.State == ParryBlockTag
}

func (a *Actor) BasicUpdate(dt float64) {
	a.ManageAnim()

	if target := a.AI.Target; target != nil {
		if a.Anim.State == WalkTag || a.Anim.State == IdleTag {
			a.Anim.FlipX = target.X > a.X
		}
	}

	if !a.PausedState() {
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

func (a *Actor) ResetState(state string) {
	a.ClimbOff()
	a.ShieldDown()
	a.Body.MaxXMultiplier = 0
	a.Stats.StaminaRecoverRateMultiplier = 0
	a.Anim.SetState(a, state)
}

func (a *Actor) Attack(attackTag string, damage, staminaDamage float64) {
	if a.PausedState() || a.Stats.Stamina <= 0 {
		return
	}
	force := -a.AttackPushForce
	if a.Anim.FlipX {
		force *= -1
	}
	a.ResetState(attackTag)

	var contacted []*Actor
	var once, blockForceOnce, blocked bool
	a.Anim.OnFrames(func(frame int) {
		hitbox, err := a.Anim.GetFrameSlice(AHitboxSliceName)
		if err != nil {
			contacted = nil

			return
		}

		blocked, contacted = a.Hitbox.Hit(a, hitbox, damage, contacted)
		if blocked && !blockForceOnce {
			blockForceOnce = true
			blockForce := force
			if !once {
				blockForce *= 2
			}
			a.Body.Vx -= blockForce
			for _, other := range contacted {
				// TODO: Find a better way to refactor this behaviour
				if other.ParryBlock {
					a.Stats.AddPoise(-damage)
					if a.Stats.Poise <= 0 {
						force := a.ReactForce / 2
						if a.X > other.X {
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
	if a.PausedState() {
		return
	}
	a.Body.ClipLadder = a.Body.ClipLadder || goingDown
	if !a.Body.OnLadder || a.Anim.State == ClimbTag {
		return
	}
	a.ResetState(ClimbTag)
	a.weightSave = a.Body.Weight
	a.Body.Weight = -1
}

func (a *Actor) ClimbOff() {
	if a.Anim.State != ClimbTag {
		return
	}
	a.Body.ClipLadder = false
	a.Body.Weight = a.weightSave
	a.Anim.Data.PlaySpeed = 1
	a.Anim.SetState(a, IdleTag)
}

func (a *Actor) Heal(effectFrame int, amount float64) {
	if a.PausedState() || a.Stats.Heal <= 0 || !a.Body.Ground {
		return
	}
	a.ResetState(ConsumeTag)
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
	a.ResetState(StaggerTag)
	a.Body.Vx = -force
}

func (a *Actor) ShieldUp() {
	if a.PausedState() || a.Anim.State == BlockTag || a.Anim.State == ParryBlockTag || a.Stats.Stamina <= 0 {
		return
	}
	a.ResetState(ParryBlockTag)
	a.Body.MaxXMultiplier = -1 / a.BlockMaxXDiv
	a.Stats.StaminaRecoverRateMultiplier = -1 / a.BlockRecoverRateDiv
	blockbox, err := a.Anim.GetFrameSlice(ABlockSliceName)
	if err != nil {
		panic(err)
	}
	a.Hitbox.PushHitbox(a, blockbox, true)
}

func (a *Actor) ShieldDown() {
	if a.Anim.State != BlockTag && a.Anim.State != ParryBlockTag {
		return
	}
	a.Anim.SetState(a, IdleTag)
	a.Body.MaxXMultiplier = 0
	a.Stats.StaminaRecoverRateMultiplier = 0
	a.Hitbox.PopHitbox()
}

func (a *Actor) Hurt(other *Actor, damage float64, poiseBreak func(force, damage float64)) {
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
		if a.Anim.State == ConsumeTag {
			a.Stagger(force)
		} else {
			a.Body.Vx -= force
		}
	}

	if a.AI.Target == nil {
		a.AI.Target = other
	}
}

func (a *Actor) Block(other *Actor, damage float64, blockBreak func(force, damage float64)) {
	if blockBreak == nil {
		blockBreak = func(force, damage float64) {
			a.ShieldDown()
			force *= damage / a.Stats.MaxHealth
			a.Stagger(force)
			a.Anim.Data.PlaySpeed = 0.5 // double time stagger.
		}
	}

	a.Stats.AddStamina(-damage)

	if a.Anim.State == ParryBlockTag {
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
	a.World.Space.Remove(a)
	for a.Hitbox.PopHitbox() != nil {
	}
	a.World.RemoveEntity(a.ID)
}

func (a *Actor) PausedState() bool {
	return utils.Contains(append(a.AttackTags, StaggerTag, ConsumeTag), a.Anim.State)
}
