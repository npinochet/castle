package actor

import (
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/libs/bump"
	"game/utils"
	"game/vars"
)

type Actor interface {
	core.Entity
	ActionTags() []string
	Anim() *anim.Comp
	Body() *body.Comp
	Hitbox() *hitbox.Comp
	Stats() *stats.Comp
}

func Hurt(a Actor, other core.Entity, damage, reactForce float64) {
	// TODO: Figure out force stuff here. For Block() too.
	ShieldDown(a)
	a.Stats().AddPoise(-damage)
	a.Stats().AddHealth(-damage)

	force := reactForce
	ax, _ := a.Position()
	if ox, _ := other.Position(); ax < ox {
		force *= -1
	}

	a.Body().Vx += force
	if a.Stats().Poise <= 0 || a.Anim().State == vars.ConsumeTag {
		force *= 2 * (damage / a.Stats().MaxHealth)
		a.Anim().SetState(vars.StaggerTag, nil)
		a.Body().Vx += force
	}

	/*if aic := core.GetComponent[*ai.Comp](a); aic != nil {
		if aic != nil && aic.Target == nil {
			aic.Target = other
		}
	}*/
}

func Block(a Actor, other core.Entity, damage, reactForce float64, contactType hitbox.ContactType) {
	a.Stats().AddStamina(-damage)
	if contactType == hitbox.ParryBlock {
		return
	}

	force := reactForce / 2
	ax, _ := a.Position()
	if ox, _ := other.Position(); ax < ox {
		force *= -1
	}

	a.Body().Vx += force
	if a.Stats().Stamina < 0 {
		ShieldDown(a)
		prevPlaySpeed := a.Anim().Data.PlaySpeed
		a.Anim().Data.PlaySpeed = 0.5 // double time stagger.
		force *= 2 * (damage / a.Stats().MaxHealth)
		a.Anim().SetState(vars.StaggerTag, func() { a.Anim().Data.PlaySpeed = prevPlaySpeed })
		a.Body().Vx += force
	}
}

// TODO: This is a mess...
func Act(a Actor, actionTag string, damage, staminaDamage, reactForce, pushForce float64) {
	if PausingState(a) || a.Stats().Stamina <= 0 {
		return
	}
	a.Anim().SetState(actionTag, nil)

	var contactType hitbox.ContactType
	var contacted []*hitbox.Comp
	var once bool
	a.Anim().OnSlicePresent(vars.HitboxSliceName, func(slice bump.Rect, segmented bool) {
		if segmented {
			contacted = nil
		}
		contactType, contacted = a.Hitbox().HitFromHitBox(slice, damage, contacted)
		/*if len(contacted) > 0 && c.Entity == entity.PlayerRef.Entity && !freezeOnce {
			freezeOnce = true
			c.World.Freeze(0.1)
			c.World.Camera.Shake(0.1, 1)
		}*/
		if contactType != hitbox.Hit {
			blockForce := reactForce / 2
			if !once {
				blockForce *= 2
			}
			if a.Anim().FlipX {
				blockForce *= -1
			}
			a.Body().Vx += blockForce
			if contactType == hitbox.ParryBlock {
				if a.Stats().AddPoise(-damage); a.Stats().Poise <= 0 {
					a.Anim().SetState(vars.StaggerTag, nil)
					force := reactForce
					if a.Anim().FlipX {
						force *= -1
					}
					a.Body().Vx += force * (damage / a.Stats().MaxHealth)
				}
			}
		}
		if !once {
			once = true
			a.Stats().AddStamina(-staminaDamage)
			force := pushForce
			if a.Anim().FlipX {
				force *= -1
			}
			a.Body().Vx -= force
		}
	})
}

func ShieldUp(a Actor) {
	if PausingState(a) || BlockingState(a) || a.Stats().Stamina <= 0 {
		return
	}
	prevMaxX, prevStaminaRecoverRate := a.Body().MaxX, a.Stats().StaminaRecoverRate
	a.Body().MaxX /= 2
	a.Stats().StaminaRecoverRate /= 3
	a.Anim().SetState(vars.ParryBlockTag, func() { a.Body().MaxX, a.Stats().StaminaRecoverRate = prevMaxX, prevStaminaRecoverRate })
	blockSlice, err := a.Anim().FrameSlice(vars.BlockSliceName)
	if err != nil {
		panic(err)
	}
	a.Hitbox().PushHitbox(blockSlice, hitbox.ParryBlock, func() hitbox.ContactType {
		if a.Anim().State == vars.ParryBlockTag {
			return hitbox.ParryBlock
		}

		return hitbox.Block
	})
}

func ShieldDown(a Actor) {
	if !BlockingState(a) {
		return
	}
	a.Anim().SetState(vars.IdleTag, nil)
	a.Hitbox().PopHitbox()
}

func PausingState(a Actor) bool {
	return utils.Contains(append(a.ActionTags(), vars.StaggerTag, vars.ConsumeTag), a.Anim().State)
}

func BlockingState(a Actor) bool {
	return a.Anim().State == vars.BlockTag || a.Anim().State == vars.ParryBlockTag
}

func CanJump(a Actor) bool {
	return (a.Anim().State == vars.ClimbTag || a.Body().Ground) &&
		!BlockingState(a) && a.Anim().State != vars.ConsumeTag
}

func ClimbOn(a Actor, goingDown bool) {
	if PausingState(a) {
		return
	}
	/*a.Body().ClipLadder = a.Body().ClipLadder || goingDown
	if !a.Body().OnLadder || a.Anim().State == anim.ClimbTag {
		return
	}
	*/
	prevWeight := a.Body().Weight
	a.Body().Weight = -1
	a.Anim().SetState(vars.ClimbTag, func() { a.Body().Weight = prevWeight })
}

func ClimbOff(a Actor) {
	if a.Anim().State != vars.ClimbTag {
		return
	}
	//a.Body().ClipLadder = false
	a.Anim().SetState(vars.IdleTag, nil)
}

func Remove(a Actor) {
	if a.Body() != nil {
		vars.World.Space.Remove(a)
	}
	if a.Hitbox() != nil {
		for a.Hitbox().PopHitbox() != nil { //nolint: revive
		}
	}
	vars.World.Remove(a)
}

func Heal(a Actor, effectFrame int, amount float64) {
	if PausingState(a) || a.Stats().Heal <= 0 || !a.Body().Ground {
		return
	}
	prevMaxX := a.Body().MaxX
	a.Body().MaxX /= 2
	a.Anim().SetState(vars.ConsumeTag, func() { a.Body().MaxX = prevMaxX })
	a.Anim().OnFrame(effectFrame, func() { // TODO: Can be replaced with OnSlicePresent.
		a.Stats().AddHeal(-1)
		a.Stats().AddHealth(amount)
	})
}
