package actor

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/entity/particle"
	"game/libs/bump"
	"game/utils"
	"game/vars"
	"image/color"
	"log"
)

const dieSeconds = 1

type Actor interface {
	core.Entity
	Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp)
}

type Control struct {
	actor    Actor
	anim     *anim.Comp
	body     *body.Comp
	hitbox   *hitbox.Comp
	stats    *stats.Comp
	ai       *ai.Comp
	paused   bool
	dieTimer float64
}

func NewControl(a Actor) *Control {
	c := &Control{actor: a, dieTimer: dieSeconds}
	c.anim, c.body, c.hitbox, c.stats, c.ai = a.Comps()

	return c
}

func (c *Control) Init() {
	hurtbox, err := c.anim.FrameSlice(vars.HurtboxSliceName)
	if err != nil {
		log.Panicf("actor: no hurtbox found: %s", err)
	}
	c.hitbox.PushHitbox(hurtbox, hitbox.Hit, nil)
	c.hitbox.HitFunc = func(other core.Entity, _ *bump.Collision, damage float64, contactType hitbox.ContactType) {
		switch contactType {
		case hitbox.Hit:
			c.Hurt(other, damage, vars.DefaultReactForce)
		case hitbox.Block, hitbox.ParryBlock:
			c.Block(other, damage, vars.DefaultReactForce, contactType)
		}
	}
}

func (c *Control) SimpleUpdate(dt float64) {
	if c.stats.Health <= 0 {
		c.Die(dt)

		return
	}
	c.stats.Pause = c.PausingState()
	if state := c.anim.State; state == vars.IdleTag || state == vars.WalkTag {
		nextState := vars.IdleTag
		if c.body.Vx != 0 {
			nextState = vars.WalkTag
		}
		c.anim.SetState(nextState)
		if c.ai != nil && c.ai.Target != nil {
			tx, _, tw, _ := c.ai.Target.Rect()
			x, _, w, _ := c.actor.Rect()
			c.anim.FlipX = tx+tw/2 > x+w/2
		}
	}
}

func (c *Control) Hurt(other core.Entity, damage, reactForce float64) {
	// TODO: Figure out force stuff here. For Block() too.
	c.ShieldDown()
	c.stats.AddPoise(-damage)
	c.stats.AddHealth(-damage)

	force := reactForce
	ax, _ := c.actor.Position()
	if ox, _ := other.Position(); ax < ox {
		force *= -1
	}

	c.body.Vx += force
	if c.stats.Poise <= 0 || c.anim.State == vars.ConsumeTag {
		force *= 2 * (damage / c.stats.MaxHealth)
		c.anim.SetState(vars.StaggerTag)
		c.body.Vx += force
	}

	if c.ai != nil && c.ai.Target == nil {
		c.ai.Target = other
	}
}

func (c *Control) Block(other core.Entity, damage, reactForce float64, contactType hitbox.ContactType) {
	c.stats.AddHealth(-damage / 10)
	c.stats.AddStamina(-damage)
	if contactType == hitbox.ParryBlock {
		return
	}

	force := reactForce / 2
	ax, _ := c.actor.Position()
	if ox, _ := other.Position(); ax < ox {
		force *= -1
	}

	c.body.Vx += force
	if c.stats.Stamina < 0 {
		c.ShieldDown()
		c.anim.SetState(vars.StaggerTag)
		prevPlaySpeed := c.anim.Data.PlaySpeed
		c.anim.Data.PlaySpeed = 0.5 // double time stagger.
		c.anim.SetExitCallback(func() { c.anim.Data.PlaySpeed = prevPlaySpeed }, nil)
		force *= 2 * (damage / c.stats.MaxHealth)
		c.body.Vx += force
	}
}

func (c *Control) Die(dt float64) {
	c.paused = true
	c.anim.SetState(vars.StaggerTag)
	if c.dieTimer -= dt; c.dieTimer > 0 {
		alpha := uint8(50 + 205*float32(c.dieTimer)/dieSeconds)
		c.anim.ColorScale = color.RGBA{alpha, alpha, alpha, alpha}

		return
	}

	vars.World.Remove(c.actor)
	for i := 0; i < c.stats.Exp; i++ {
		vars.World.Add(particle.NewFlake(c.actor)) // TODO: keep actor package isolated somehow, move Flake to entity package.
	}
}

// TODO: This is a mess...
func (c *Control) Attack(attackTag string, damage, staminaDamage, reactForce, pushForce float64) {
	if c.PausingState() || c.stats.Stamina <= 0 {
		return
	}
	damage *= 1 + c.stats.AttackMult
	c.anim.SetState(attackTag)
	c.paused = true
	c.anim.SetExitCallback(func() { c.paused = false }, nil)

	var contactType hitbox.ContactType
	var contacted []*hitbox.Comp
	var once bool
	c.anim.OnSlicePresent(vars.HitboxSliceName, func(slice bump.Rect, segmented bool) {
		if segmented {
			contacted = nil
		}
		contactType, contacted = c.hitbox.HitFromHitBox(slice, damage, contacted)
		/*if len(contacted) > 0 && c.Entity == vars.Player && !freezeOnce { // TODO: This is a mess...
			freezeOnce = true
			c.World.Freeze(0.1)
			c.World.Camera.Shake(0.1, 1)
		}*/
		if contactType != hitbox.Hit {
			blockForce := reactForce / 2
			if !once {
				blockForce *= 2
			}
			if c.anim.FlipX {
				blockForce *= -1
			}
			c.body.Vx += blockForce
			if contactType == hitbox.ParryBlock {
				if c.stats.AddPoise(-damage); c.stats.Poise <= 0 {
					c.anim.SetState(vars.StaggerTag)
					force := reactForce
					if c.anim.FlipX {
						force *= -1
					}
					c.body.Vx += force * (damage / c.stats.MaxHealth)
				}
			}
		}
		if !once {
			once = true
			c.stats.AddStamina(-staminaDamage)
			force := pushForce
			if c.anim.FlipX {
				force *= -1
			}
			c.body.Vx -= force
		}
	})
}

func (c *Control) ShieldUp() {
	if c.PausingState() || c.BlockingState() || c.stats.Stamina <= 0 {
		return
	}
	c.anim.SetState(vars.ParryBlockTag)
	prevMaxX, prevStaminaRecoverRate := c.body.MaxX, c.stats.StaminaRecoverRate
	c.body.MaxX /= 2
	c.stats.StaminaRecoverRate /= 3
	c.anim.SetExitCallback(func() {
		c.body.MaxX = prevMaxX
		c.stats.StaminaRecoverRate = prevStaminaRecoverRate
	}, func() bool { return !c.BlockingState() })
	blockSlice, err := c.anim.FrameSlice(vars.BlockSliceName)
	if err != nil {
		panic(err)
	}
	c.hitbox.PushHitbox(blockSlice, hitbox.ParryBlock, func() hitbox.ContactType {
		if c.anim.State == vars.ParryBlockTag {
			return hitbox.ParryBlock
		}

		return hitbox.Block
	})
}

func (c *Control) ShieldDown() {
	if !c.BlockingState() {
		return
	}
	c.anim.SetState(vars.IdleTag)
	c.hitbox.PopHitbox()
}

func (c *Control) PausingState() bool {
	return c.paused || utils.Contains([]string{vars.StaggerTag, vars.ConsumeTag}, c.anim.State)
}

func (c *Control) BlockingState() bool {
	return c.anim.State == vars.BlockTag || c.anim.State == vars.ParryBlockTag
}

func (c *Control) CanJump() bool {
	return (c.anim.State == vars.ClimbTag || c.body.Ground) &&
		!c.BlockingState() && c.anim.State != vars.ConsumeTag
}

func (c *Control) ClimbOn(pressedDown bool) {
	if c.PausingState() || c.anim.State == vars.ClimbTag {
		return
	}
	if pressedDown {
		if (c.body.Ground && c.body.InsidePassThrough) || !c.body.DropThrough() {
			return
		}
	}
	if !c.body.InsidePassThrough && !c.body.DropThrough() {
		return
	}
	c.ShieldDown()
	c.anim.SetState(vars.ClimbTag)
	prevWeight := c.body.Weight
	c.body.Weight = 0
	c.anim.SetExitCallback(func() { c.body.Weight = prevWeight }, nil)
}

func (c *Control) ClimbOff() {
	if c.anim.State != vars.ClimbTag {
		return
	}
	c.anim.SetState(vars.IdleTag)
}

func (c *Control) Heal(effectFrame int) {
	if c.PausingState() || c.stats.Heal <= 0 || !c.body.Ground {
		return
	}
	c.anim.SetState(vars.ConsumeTag)
	prevMaxX := c.body.MaxX
	c.body.MaxX /= 3
	c.anim.SetExitCallback(func() { c.body.MaxX = prevMaxX }, nil)
	c.anim.OnFrame(effectFrame, func() { // TODO: Can be replaced with OnSlicePresent.
		c.stats.AddHeal(-1)
	})
}
