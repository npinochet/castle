package actor

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/libs/bump"
	"game/vars"
	"image/color"
	"log"
	"slices"
	"strings"
)

const (
	dieSeconds   = 1
	flashSeconds = 0.05
)

var DieParticle func(core.Entity) core.Entity

type Actor interface {
	core.Entity
	Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp)
}

type Control struct {
	actor                Actor
	anim                 *anim.Comp
	body                 *body.Comp
	hitbox               *hitbox.Comp
	stats                *stats.Comp
	ai                   *ai.Comp
	paused               bool
	dieTimer, flashTimer float64
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
	c.anim.ColorScale = color.White
	if c.flashTimer -= dt; c.flashTimer > 0 {
		c.anim.ColorScale = anim.WhiteScalerColor
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
		c.Stagger(force, false, 1)
	}
	c.flashTimer = flashSeconds

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
		force *= 2 * (damage / c.stats.MaxHealth)
		c.Stagger(force, false, 2)
	}
}

func (c *Control) Die(dt float64) {
	c.paused = true
	c.Stagger(0, false, 1)
	if c.dieTimer -= dt; c.dieTimer > 0 {
		alpha := uint8(50 + 205*float32(c.dieTimer)/dieSeconds)
		c.anim.ColorScale = color.RGBA{alpha, alpha, alpha, alpha}

		return
	}

	vars.World.Remove(c.actor)
	if DieParticle != nil {
		for i := 0; i < c.stats.Exp; i++ {
			vars.World.Add(DieParticle(c.actor))
		}
	}
}

// TODO: This is a mess...
func (c *Control) Attack(attackTag string, damage, staminaDamage, reactForce, pushForce float64) {
	if c.PausingState() || c.stats.Stamina <= 0 {
		return
	}
	damage *= 1 + c.stats.AttackMult
	if strings.HasPrefix(c.anim.State, attackTag) && c.anim.Data.Animation(c.anim.State+"C") != nil {
		attackTag = c.anim.State + "C"
	}

	c.ShieldDown()
	c.anim.SetState(attackTag)
	c.anim.SetStateEffect(func() func() {
		c.paused = true

		return func() { c.paused = false }
	})

	if c.anim.Data.Animation(attackTag+"C") != nil {
		lastFrame := c.anim.Data.CurrentAnimation.To - c.anim.Data.CurrentAnimation.From
		c.anim.OnFrame(lastFrame, func() { c.paused = false })
	}

	var contactType hitbox.ContactType
	var contacted []*hitbox.Comp
	var shakeNum int
	var once bool
	c.anim.OnSlicePresent(vars.HitboxSliceName, func(slice bump.Rect, segmented bool) {
		if segmented {
			contacted = nil
		}
		contactType, contacted = c.hitbox.HitFromHitBox(slice, damage, contacted)
		if c.actor == vars.Player && shakeNum != len(contacted) { // TODO: This is an ugly hack
			shakeNum = len(contacted)
			vars.World.Camera.Shake(0.1, 0.5)
		}
		if contactType == hitbox.ParryBlock {
			if c.stats.AddPoise(-damage); c.stats.Poise <= 0 {
				c.Stagger(reactForce*(damage/c.stats.MaxHealth), true, 1)
			}
		}
		if !once {
			once = true
			c.stats.AddStamina(-staminaDamage)
			force := pushForce
			if contactType >= hitbox.Block {
				force = reactForce
			}
			if (contactType >= hitbox.Block && c.anim.FlipX) || (contactType < hitbox.Block && !c.anim.FlipX) {
				force *= -1
			}
			c.body.Vx += force
		}
	})
}

func (c *Control) ShieldUp() {
	if c.PausingState() || c.BlockingState() {
		return
	}
	c.anim.SetState(vars.ParryBlockTag)
	c.anim.SetStateEffect(func() func() {
		prevMaxX, prevStaminaRecoverRate := c.body.MaxX, c.stats.StaminaRecoverRate
		c.body.MaxX /= 2
		c.stats.StaminaRecoverRate /= 3

		return func() { c.body.MaxX, c.stats.StaminaRecoverRate = prevMaxX, prevStaminaRecoverRate }
	}, vars.ParryBlockTag, vars.BlockTag)
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
	return c.paused || slices.Contains([]string{vars.StaggerTag, vars.ConsumeTag}, c.anim.State)
}

func (c *Control) BlockingState() bool {
	return c.anim.State == vars.BlockTag || c.anim.State == vars.ParryBlockTag
}

func (c *Control) CanJump() bool {
	return c.stats.Stamina > 0 && (c.anim.State == vars.ClimbTag || c.body.Ground) && !c.BlockingState() && c.anim.State != vars.ConsumeTag
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
	if !c.body.InsidePassThrough {
		return
	}
	c.ShieldDown()
	c.anim.SetState(vars.ClimbTag)
	c.anim.SetStateEffect(func() func() {
		prevWeight := c.body.Weight
		c.body.Weight = 0

		return func() { c.body.Weight = prevWeight }
	})
}

func (c *Control) Stagger(force float64, moveBack bool, timeMult float64) {
	c.ShieldDown()
	c.anim.SetState(vars.StaggerTag)
	if timeMult != 1 {
		c.anim.SetStateEffect(func() func() {
			prevPlaySpeed := c.anim.Data.PlaySpeed
			c.anim.Data.PlaySpeed = float32(1.0 / timeMult)

			return func() { c.anim.Data.PlaySpeed = prevPlaySpeed }
		})
	}
	if moveBack && c.anim.FlipX {
		force *= -1
	}
	c.body.Vx += force
}

func (c *Control) ClimbOff() {
	if c.anim.State != vars.ClimbTag {
		return
	}
	c.anim.SetState(vars.IdleTag)
}

func (c *Control) Heal() {
	if c.PausingState() || c.stats.Heal <= 0 || !c.body.Ground {
		return
	}
	c.ShieldDown()
	c.anim.SetState(vars.ConsumeTag)
	c.anim.SetStateEffect(func() func() {
		prevMaxX := c.body.MaxX
		c.body.MaxX /= 3

		return func() { c.body.MaxX = prevMaxX }
	})
	c.anim.OnSlicePresent(vars.HealSliceName, func(_ bump.Rect, segmented bool) {
		if segmented {
			c.stats.AddHeal(-1)
		}
	})
	c.anim.OnFrame(3, func() { c.stats.AddHeal(-1) }) // TODO: Add "healbox" slice to player anim file and delete this line
}
