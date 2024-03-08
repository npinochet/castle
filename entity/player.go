package entity

import (
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/libs/bump"
	"game/utils"
	"game/vars"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	playerAnimFile                                 = "assets/knight"
	playerWidth, playerHeight                      = 8, 11
	playerOffsetX, playerOffsetY, playerOffsetFlip = -10, -3, 17
	playerMaxX, playerSpeed, playerJumpSpeed       = 60, 350, 110
	playerDamage, playerHeal                       = 20, 20
	playerHealFrame                                = 3

	keyBufferDuration = 500 * time.Millisecond
)

const (
	defaultAttackPushForce = 100
	defaultReactForce      = 50
)

var PlayerRef *Player

type Player struct {
	*core.BaseEntity
	Anim   *anim.Comp
	Body   *body.Comp
	Hitbox *hitbox.Comp
	Stats  *stats.Comp

	ActionTags                  []string
	Speed, JumpSpeed            float64
	ReactForce, AttackPushForce float64

	pad utils.ControlPack
}

func NewPlayer(x, y float64, actionTags []string) *Player {
	p := &Player{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: playerWidth, H: playerHeight},
		Anim:       &anim.Comp{FilesName: playerAnimFile, OX: playerOffsetX, OY: playerOffsetY, OXFlip: playerOffsetFlip},
		Body:       &body.Comp{MaxX: playerMaxX},
		Hitbox:     &hitbox.Comp{},
		Stats:      &stats.Comp{Hud: true, NoDebug: true, Stamina: 65},

		AttackPushForce: defaultAttackPushForce,
		ReactForce:      defaultReactForce,
		ActionTags:      actionTags,
		Speed:           playerSpeed, JumpSpeed: playerJumpSpeed,

		pad: utils.NewControlPack(),
	}
	p.Add(p.Anim, p.Body, p.Hitbox, p.Stats)

	return p
}

func (p *Player) Init() {
	p.Hitbox.HitFunc = func(other core.Entity, col *bump.Collision, damage float64, contactType hitbox.ContactType) {
		switch contactType {
		case hitbox.Hit:
			p.Hurt(other, damage)
			vars.World.Camera.Shake(0.5, 1)
			vars.World.Freeze(0.1)
		case hitbox.Block, hitbox.ParryBlock:
			p.Block(other, damage, contactType)
		}
	}

	hurtbox, err := p.Anim.FrameSlice(vars.HurtboxSliceName)
	if err != nil {
		log.Panicf("player: %s", err)
	}
	p.Hitbox.PushHitbox(hurtbox, hitbox.Hit, nil)
}

func (p *Player) Hurt(other core.Entity, damage float64) {
	// TODO: Figure out force stuff here. For Block() too.
	p.ShieldDown()
	p.Stats.AddPoise(-damage)
	p.Stats.AddHealth(-damage)

	force := p.ReactForce
	if ox, _ := other.Position(); p.X < ox {
		force *= -1
	}

	p.Body.Vx += force
	if p.Stats.Poise <= 0 || p.Anim.State == vars.ConsumeTag {
		force *= 2 * (damage / p.Stats.MaxHealth)
		p.Anim.SetState(vars.StaggerTag, nil)
		p.Body.Vx += force
	}

	/*if aic := core.GetComponent[*ai.Comp](a); aic != nil {
		if aic != nil && aic.Target == nil {
			aic.Target = other
		}
	}*/
}

func (p *Player) Block(other core.Entity, damage float64, contactType hitbox.ContactType) {
	p.Stats.AddStamina(-damage)
	if contactType == hitbox.ParryBlock {
		return
	}

	force := p.ReactForce / 2
	if ox, _ := other.Position(); p.X < ox {
		force *= -1
	}

	p.Body.Vx += force
	if p.Stats.Stamina < 0 {
		p.ShieldDown()
		prevPlaySpeed := p.Anim.Data.PlaySpeed
		p.Anim.Data.PlaySpeed = 0.5 // double time stagger.
		force *= 2 * (damage / p.Stats.MaxHealth)
		p.Anim.SetState(vars.StaggerTag, func() { p.Anim.Data.PlaySpeed = prevPlaySpeed })
		p.Body.Vx += force
	}
}

func (p *Player) SimpleUpdate(dt float64, target core.Entity) {
	if p.Stats.Health <= 0 {
		p.Remove()

		return
	}
	p.Stats.Pause = p.PausingState()
	if state := p.Anim.State; state == vars.IdleTag || state == vars.WalkTag {
		nextState := vars.IdleTag
		if p.Body.Vx != 0 {
			nextState = vars.WalkTag
		}
		p.Anim.SetState(nextState, nil)
	}

	if target != nil {
		tx, _ := target.Position()
		if p.Anim.State == vars.WalkTag || p.Anim.State == vars.IdleTag {
			p.Anim.FlipX = tx > p.X
		}
	}

	if !p.PausingState() {
		if p.Anim.FlipX {
			p.Body.Vx += p.Speed * dt
		} else {
			p.Body.Vx -= p.Speed * dt
		}
	}
}

func (p *Player) PausingState() bool {
	return utils.Contains(append(p.ActionTags, vars.StaggerTag, vars.ConsumeTag), p.Anim.State)
}

func (p *Player) BlockingState() bool {
	return p.Anim.State == vars.BlockTag || p.Anim.State == vars.ParryBlockTag
}

func (p *Player) CanJump() bool {
	return (p.Anim.State == vars.ClimbTag || p.Body.Ground) &&
		!p.BlockingState() && p.Anim.State != vars.ConsumeTag
}

// TODO: This is a mess...
func (p *Player) Act(actionTag string, damage, staminaDamage float64) {
	if p.PausingState() || p.Stats.Stamina <= 0 {
		return
	}
	p.Anim.SetState(actionTag, nil)

	var contactType hitbox.ContactType
	var contacted []*hitbox.Comp
	var once bool
	p.Anim.OnSlicePresent(vars.HitboxSliceName, func(slice bump.Rect, segmented bool) {
		if segmented {
			contacted = nil
		}
		contactType, contacted = p.Hitbox.HitFromHitBox(slice, damage, contacted)
		/*if len(contacted) > 0 && c.Entity == entity.PlayerRef.Entity && !freezeOnce {
			freezeOnce = true
			c.World.Freeze(0.1)
			c.World.Camerp.Shake(0.1, 1)
		}*/
		if contactType != hitbox.Hit {
			blockForce := p.ReactForce / 2
			if !once {
				blockForce *= 2
			}
			if p.Anim.FlipX {
				blockForce *= -1
			}
			p.Body.Vx += blockForce
			if contactType == hitbox.ParryBlock {
				if p.Stats.AddPoise(-damage); p.Stats.Poise <= 0 {
					p.Anim.SetState(vars.StaggerTag, nil)
					force := p.ReactForce
					if p.Anim.FlipX {
						force *= -1
					}
					p.Body.Vx += force * (damage / p.Stats.MaxHealth)
				}
			}
		}
		if !once {
			once = true
			p.Stats.AddStamina(-staminaDamage)
			force := p.AttackPushForce
			if p.Anim.FlipX {
				force *= -1
			}
			p.Body.Vx -= force
		}
	})
}

func (p *Player) ClimbOn(goingDown bool) {
	if p.PausingState() {
		return
	}
	/*p.Body.ClipLadder = p.Body.ClipLadder || goingDown
	if !p.Body.OnLadder || p.Anim.State == anim.ClimbTag {
		return
	}
	*/
	prevWeight := p.Body.Weight
	p.Body.Weight = -1
	p.Anim.SetState(vars.ClimbTag, func() { p.Body.Weight = prevWeight })
}

func (p *Player) ClimbOff() {
	if p.Anim.State != vars.ClimbTag {
		return
	}
	//p.Body.ClipLadder = false
	p.Anim.SetState(vars.IdleTag, nil)
}

func (p *Player) Heal(effectFrame int, amount float64) {
	if p.PausingState() || p.Stats.Heal <= 0 || !p.Body.Ground {
		return
	}
	prevMaxX := p.Body.MaxX
	p.Body.MaxX /= 2
	p.Anim.SetState(vars.ConsumeTag, func() { p.Body.MaxX = prevMaxX })
	p.Anim.OnFrame(effectFrame, func() { // TODO: Can be replaced with OnSlicePresent.
		p.Stats.AddHeal(-1)
		p.Stats.AddHealth(amount)
	})
}

func (p *Player) ShieldUp() {
	if p.PausingState() || p.BlockingState() || p.Stats.Stamina <= 0 {
		return
	}
	prevMaxX, prevStaminaRecoverRate := p.Body.MaxX, p.Stats.StaminaRecoverRate
	p.Body.MaxX /= 2
	p.Stats.StaminaRecoverRate /= 3
	p.Anim.SetState(vars.ParryBlockTag, func() { p.Body.MaxX, p.Stats.StaminaRecoverRate = prevMaxX, prevStaminaRecoverRate })
	blockSlice, err := p.Anim.FrameSlice(vars.BlockSliceName)
	if err != nil {
		panic(err)
	}
	p.Hitbox.PushHitbox(blockSlice, hitbox.ParryBlock, func() hitbox.ContactType {
		if p.Anim.State == vars.ParryBlockTag {
			return hitbox.ParryBlock
		}

		return hitbox.Block
	})
}

func (p *Player) ShieldDown() {
	if !p.BlockingState() {
		return
	}
	p.Anim.SetState(vars.IdleTag, nil)
	p.Hitbox.PopHitbox()
}

func (p *Player) Remove() {
	if p.Body != nil {
		vars.World.Space.Remove(p)
	}
	if p.Hitbox != nil {
		for p.Hitbox.PopHitbox() != nil { //nolint: revive
		}
	}
	vars.World.Remove(p)
}

func (p *Player) Update(dt float64) {
	p.input(dt)
	if moving := p.pad.KeyDown(utils.KeyLeft) || p.pad.KeyDown(utils.KeyRight); p.Anim.State == vars.WalkTag && !moving {
		p.Anim.SetState(vars.IdleTag, nil)
	}
}

func (p *Player) input(dt float64) { // TODO: refactor this
	actionPressed := p.pad.KeyPressedBuffered(utils.KeyAction, keyBufferDuration)
	healPressed := p.pad.KeyPressedBuffered(utils.KeyHeal, keyBufferDuration)
	if p.PausingState() && p.Anim.State != vars.ConsumeTag {
		return
	}
	if actionPressed {
		p.Act(vars.AttackTag, playerDamage, playerDamage)
	}
	if healPressed {
		p.Heal(playerHealFrame, playerHeal)
	}
	if p.pad.KeyDown(utils.KeyGuard) {
		p.ShieldUp()
	}
	if p.pad.KeyReleased(utils.KeyGuard) {
		p.ShieldDown()
	}
	p.inputClimbing(dt)

	flip := p.Anim.FlipX
	if p.pad.KeyDown(utils.KeyLeft) {
		if math.Abs(p.Body.Vx) <= p.Body.MaxX {
			p.Body.Vx -= p.Speed * dt
		}
		flip = false
	}
	if p.pad.KeyDown(utils.KeyRight) {
		if math.Abs(p.Body.Vx) <= p.Body.MaxX {
			p.Body.Vx += p.Speed * dt
		}
		flip = true
	}

	if !p.BlockingState() {
		p.Anim.FlipX = flip
	}
	if p.pad.KeyPressed(utils.KeyJump) && p.CanJump() {
		p.ClimbOff()
		p.Body.Vy = -p.JumpSpeed
	}

	// TODO: Debug, remove later.
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		p.Stats.Heal = p.Stats.MaxHeal
	}
}

func (p *Player) inputClimbing(dt float64) {
	if p.pad.KeyDown(utils.KeyUp) || p.pad.KeyDown(utils.KeyDown) {
		p.ClimbOn(p.pad.KeyDown(utils.KeyDown))
	}
	if p.Anim.State != vars.ClimbTag {
		return
	}
	/*if !p.Body.OnLadder {
		p.Control.ClimbOff()
	}*/
	p.Body.Vy = 0
	speed := p.Speed * 5 * dt
	if p.pad.KeyDown(utils.KeyUp) {
		p.Body.Vy = -speed
	}
	if p.pad.KeyDown(utils.KeyDown) {
		p.Body.Vy = speed
	}
}
