package control

import (
	"game/comps/basic/anim"
	"game/comps/basic/body"
	"game/comps/basic/hitbox"
	"game/comps/basic/stats"
	"game/core"
	"game/libs/bump"
	"game/utils"
)

type Comp struct {
	*core.Entity
	Body                              *body.Comp
	Hitbox                            *hitbox.Comp
	Anim                              *anim.Comp
	Stats                             *stats.Comp
	AttackTags                        []string
	Speed                             float64
	BlockMaxXDiv, BlockRecoverRateDiv float64
	ReactForce, AttackPushForce       float64
	weightSave                        float64
}

func (c *Comp) SetSpeed(speed, maxX float64) { c.Speed, c.Body.MaxX = speed, maxX }

func (c *Comp) Init(e *core.Entity) {
	c.Entity = e
	c.Body = core.GetComponent[*body.Comp](e)
	c.Body = core.GetComponent[*body.Comp](e)
	c.Hitbox = core.GetComponent[*hitbox.Comp](e)
	c.Anim = core.GetComponent[*anim.Comp](e)
	c.Stats = core.GetComponent[*stats.Comp](e)
}

func (c *Comp) ManageAnim() {
	// TODO: Make more general, maybe add speed in the mix.
	if state := c.Anim.State; state == anim.IdleTag || state == anim.WalkTag {
		nextState := anim.IdleTag
		if c.Body.Vx != 0 {
			nextState = anim.WalkTag
		}
		c.Anim.SetState(nextState)
	}
	c.Stats.Pause = c.PausedState()
}

func (c *Comp) SimpleUpdate(dt float64, target *core.Entity) {
	c.ManageAnim()
	if target != nil {
		if c.Anim.State == anim.WalkTag || c.Anim.State == anim.IdleTag {
			c.Anim.FlipX = target.X > c.X
		}
	}

	if !c.PausedState() {
		if c.Anim.FlipX {
			c.Body.Vx += c.Speed * dt
		} else {
			c.Body.Vx -= c.Speed * dt
		}
	}

	if c.Stats.Health <= 0 {
		c.Remove()
	}
}

func (c *Comp) PausedState() bool {
	return utils.Contains(append(c.AttackTags, anim.StaggerTag, anim.ConsumeTag), c.Anim.State)
}

func (c *Comp) ResetState(state string) { // TODO: I hate ResetState, kill it somehow, fucking control comp
	c.ClimbOff()
	c.ShieldDown()
	c.Body.MaxXMultiplier = 0
	c.Stats.StaminaRecoverRateMultiplier = 0
	c.Anim.SetState(state)
}

// TODO: This is a mess...
func (c *Comp) Attack(attackTag string, damage, staminaDamage float64) {
	if c.PausedState() || c.Stats.Stamina <= 0 {
		return
	}
	c.ResetState(attackTag)

	var contacted []*hitbox.Comp
	var once bool
	c.Anim.OnHitboxUpdate(anim.HitboxSliceName, func(hitboxRect bump.Rect, newHitbox bool) {
		if newHitbox {
			contacted = nil
		}
		var blocked hitbox.BlockType
		blocked, contacted = c.Hitbox.HitFromHitBox(hitboxRect, damage, contacted)
		/*if len(contacted) > 0 && c.Entity == entity.PlayerRef.Entity && !freezeOnce {
			freezeOnce = true
			c.World.Freeze(0.1)
			c.World.Camera.Shake(0.1, 1)
		}*/
		if blocked != hitbox.NoBlock {
			blockForce := c.ReactForce / 4
			if c.Anim.FlipX {
				blockForce *= -1
			}
			if !once {
				blockForce *= 2
			}
			c.Body.Vx += blockForce
			if blocked == hitbox.ParryBlock {
				c.Stats.AddPoise(-damage)
				if c.Stats.Poise <= 0 {
					force := -c.ReactForce / 2
					if c.Anim.FlipX {
						force *= -1
					}
					c.Stagger(force * (damage / c.Stats.MaxHealth))
				}
			}
		}
		if !once {
			once = true
			c.Stats.AddStamina(-staminaDamage)
			force := c.AttackPushForce
			if c.Anim.FlipX {
				force *= -1
			}
			c.Body.Vx -= force
		}
	})
}

func (c *Comp) ClimbOn(goingDown bool) {
	if c.PausedState() {
		return
	}
	c.Body.ClipLadder = c.Body.ClipLadder || goingDown
	if !c.Body.OnLadder || c.Anim.State == anim.ClimbTag {
		return
	}
	c.ResetState(anim.ClimbTag)
	c.weightSave = c.Body.Weight
	c.Body.Weight = -1
}

func (c *Comp) ClimbOff() {
	if c.Anim.State != anim.ClimbTag {
		return
	}
	c.Body.ClipLadder = false
	c.Body.Weight = c.weightSave
	c.Anim.Data.PlaySpeed = 1
	c.Anim.SetState(anim.IdleTag)
}

func (c *Comp) Heal(effectFrame int, amount float64) {
	if c.PausedState() || c.Stats.Heal <= 0 || !c.Body.Ground {
		return
	}
	c.ResetState(anim.ConsumeTag)
	c.Body.MaxXMultiplier = -1
	c.Anim.OnFrame(effectFrame, func() {
		c.Stats.AddHeal(-1)
		c.Stats.AddHealth(amount)
	})
}

func (c *Comp) Stagger(force float64) {
	c.ResetState(anim.StaggerTag)
	c.Body.Vx = force
}

func (c *Comp) ShieldUp() {
	if c.PausedState() || c.Anim.State == anim.BlockTag || c.Anim.State == anim.ParryBlockTag || c.Stats.Stamina <= 0 {
		return
	}
	c.ResetState(anim.ParryBlockTag)
	c.Body.MaxXMultiplier = -1 / c.BlockMaxXDiv
	c.Stats.StaminaRecoverRateMultiplier = -1 / c.BlockRecoverRateDiv
	blockbox, err := c.Anim.GetFrameHitbox(anim.BlockSliceName)
	if err != nil {
		panic(err)
	}
	c.Hitbox.PushHitbox(blockbox, true)
}

func (c *Comp) ShieldDown() {
	if c.Anim.State != anim.BlockTag && c.Anim.State != anim.ParryBlockTag {
		return
	}
	c.Anim.SetState(anim.IdleTag)
	c.Body.MaxXMultiplier = 0
	c.Stats.StaminaRecoverRateMultiplier = 0
	c.Hitbox.PopHitbox()
}

func (c *Comp) Remove() { // TODO: Donde mover esto?
	if c.Body != nil {
		c.World.Space.Remove(c.Body)
	}
	if c.Hitbox != nil {
		for c.Hitbox.PopHitbox() != nil {
		}
	}
	c.World.RemoveEntity(c.ID)
}
