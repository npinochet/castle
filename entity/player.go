package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/entity/actor"
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
	playerMaxX, playerSpeed, playerJumpSpeed, playerClimbSpeed = 60, 350, 110, 5
	playerDamage, playerHeal                                   = 20, 20
	playerHealFrame                                            = 3

	keyBufferDuration = 500 * time.Millisecond
)

type Player struct {
	*core.BaseEntity
	*actor.Control
	anim   *anim.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	stats  *stats.Comp

	speed, jumpSpeed            float64
	reactForce, attackPushForce float64

	pad utils.ControlPack
}

func NewPlayer(x, y float64) *Player {
	p := &Player{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: knightWidth, H: knightHeight},
		anim:       &anim.Comp{FilesName: knightAnimFile, OX: knightOffsetX, OY: knightOffsetY, OXFlip: knightOffsetFlip},
		body:       &body.Comp{MaxX: playerMaxX},
		hitbox:     &hitbox.Comp{},
		stats:      &stats.Comp{Hud: true, NoDebug: true, Stamina: 65},

		attackPushForce: vars.DefaultAttackPushForce,
		reactForce:      vars.DefaultReactForce,
		speed:           playerSpeed, jumpSpeed: playerJumpSpeed,

		pad: utils.NewControlPack(),
	}
	p.Add(p.anim, p.body, p.hitbox, p.stats)
	p.Control = actor.NewControl(p)
	core.SetFlag(p, vars.PlayerTeamFlag, true)

	return p
}

func (p *Player) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return p.anim, p.body, p.hitbox, p.stats, nil
}

func (p *Player) Init() {
	hurtbox, err := p.anim.FrameSlice(vars.HurtboxSliceName)
	if err != nil {
		log.Panicf("player: %s", err)
	}
	p.hitbox.PushHitbox(hurtbox, hitbox.Hit, nil)
	p.hitbox.HitFunc = func(other core.Entity, _ *bump.Collision, damage float64, contactType hitbox.ContactType) {
		switch contactType {
		case hitbox.Hit:
			p.Hurt(other, damage, p.reactForce)
			vars.World.Camera.Shake(0.5, 1)
			vars.World.Freeze(0.1)
		case hitbox.Block, hitbox.ParryBlock:
			p.Block(other, damage, p.reactForce, contactType)
		}
	}
}

func (p *Player) Update(dt float64) {
	p.input(dt)
	p.SimpleUpdate()
	if moving := p.pad.KeyDown(utils.KeyLeft) || p.pad.KeyDown(utils.KeyRight); !moving && p.anim.State == vars.WalkTag {
		p.anim.SetState(vars.IdleTag)
	}
}

func (p *Player) input(dt float64) {
	actionPressed := p.pad.KeyPressedBuffered(utils.KeyAction, keyBufferDuration)
	healPressed := p.pad.KeyPressedBuffered(utils.KeyHeal, keyBufferDuration)
	dashPressed := p.pad.KeyPressedBuffered(utils.KeyDash, keyBufferDuration)
	if p.PausingState() && p.anim.State != vars.ConsumeTag {
		return
	}
	if actionPressed() {
		p.Attack(vars.AttackTag, playerDamage, playerDamage, p.reactForce, p.attackPushForce)
	}
	if healPressed() {
		p.Heal(playerHealFrame, playerHeal)
	}
	if dashPressed() {
		speed := p.body.MaxX * 4
		if (!p.anim.FlipX && !p.pad.KeyDown(utils.KeyRight)) || p.pad.KeyDown(utils.KeyLeft) {
			speed *= -1
		}
		p.body.Vx = speed
	}
	if p.pad.KeyDown(utils.KeyGuard) {
		p.ShieldUp()
	}
	if p.pad.KeyReleased(utils.KeyGuard) {
		p.ShieldDown()
	}
	p.inputClimbing(dt)

	flip := p.anim.FlipX
	if p.pad.KeyDown(utils.KeyLeft) {
		if math.Abs(p.body.Vx) <= p.body.MaxX {
			p.body.Vx -= p.speed * dt
		}
		flip = false
	}
	if p.pad.KeyDown(utils.KeyRight) {
		if math.Abs(p.body.Vx) <= p.body.MaxX {
			p.body.Vx += p.speed * dt
		}
		flip = true
	}

	if !p.BlockingState() {
		p.anim.FlipX = flip
	}
	if p.pad.KeyPressed(utils.KeyJump) && p.CanJump() {
		p.ClimbOff()
		p.body.Vy = -p.jumpSpeed
	}

	if vars.Debug {
		if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
			p.stats.Heal = p.stats.MaxHeal
		}
	}
}

func (p *Player) inputClimbing(dt float64) {
	if p.pad.KeyDown(utils.KeyUp) || p.pad.KeyDown(utils.KeyDown) {
		p.ClimbOn(p.pad.KeyDown(utils.KeyDown))
	}
	if p.anim.State != vars.ClimbTag {
		return
	}
	if !p.body.InsidePassThrough {
		p.Control.ClimbOff()

		return
	}
	p.body.Vy = 0
	speed := p.speed * playerClimbSpeed * dt
	if p.pad.KeyDown(utils.KeyUp) {
		p.body.Vy = -speed
	}
	if p.pad.KeyDown(utils.KeyDown) {
		p.body.Vy = speed
	}
}
