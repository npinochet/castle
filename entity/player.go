package entity

import (
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
	anim   *anim.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	stats  *stats.Comp

	actionTags                  []string
	speed, jumpSpeed            float64
	reactForce, attackPushForce float64

	pad utils.ControlPack
}

func (p *Player) ActionTags() []string { return p.actionTags }
func (p *Player) Anim() *anim.Comp     { return p.anim }
func (p *Player) Body() *body.Comp     { return p.body }
func (p *Player) Hitbox() *hitbox.Comp { return p.hitbox }
func (p *Player) Stats() *stats.Comp   { return p.stats }

func NewPlayer(x, y float64, actionTags []string) *Player {
	p := &Player{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: playerWidth, H: playerHeight},
		anim:       &anim.Comp{FilesName: playerAnimFile, OX: playerOffsetX, OY: playerOffsetY, OXFlip: playerOffsetFlip},
		body:       &body.Comp{MaxX: playerMaxX},
		hitbox:     &hitbox.Comp{},
		stats:      &stats.Comp{Hud: true, NoDebug: true, Stamina: 65},

		attackPushForce: defaultAttackPushForce,
		reactForce:      defaultReactForce,
		actionTags:      actionTags,
		speed:           playerSpeed, jumpSpeed: playerJumpSpeed,

		pad: utils.NewControlPack(),
	}
	p.Add(p.anim, p.body, p.hitbox, p.stats)

	return p
}

func (p *Player) Init() {
	p.hitbox.HitFunc = func(other core.Entity, col *bump.Collision, damage float64, contactType hitbox.ContactType) {
		switch contactType {
		case hitbox.Hit:
			actor.Hurt(p, other, damage, p.reactForce)
			vars.World.Camera.Shake(0.5, 1)
			vars.World.Freeze(0.1)
		case hitbox.Block, hitbox.ParryBlock:
			actor.Block(p, other, damage, p.reactForce, contactType)
		}
	}

	hurtbox, err := p.anim.FrameSlice(vars.HurtboxSliceName)
	if err != nil {
		log.Panicf("player: %s", err)
	}
	p.hitbox.PushHitbox(hurtbox, hitbox.Hit, nil)
}

func (p *Player) SimpleUpdate(dt float64, target core.Entity) {
	if p.stats.Health <= 0 {
		actor.Remove(p)

		return
	}
	p.stats.Pause = actor.PausingState(p)
	if state := p.anim.State; state == vars.IdleTag || state == vars.WalkTag {
		nextState := vars.IdleTag
		if p.body.Vx != 0 {
			nextState = vars.WalkTag
		}
		p.anim.SetState(nextState, nil)
	}

	if target != nil {
		tx, _ := target.Position()
		if p.anim.State == vars.WalkTag || p.anim.State == vars.IdleTag {
			p.anim.FlipX = tx > p.X
		}
	}

	if !actor.PausingState(p) {
		if p.anim.FlipX {
			p.body.Vx += p.speed * dt
		} else {
			p.body.Vx -= p.speed * dt
		}
	}
}

func (p *Player) Update(dt float64) {
	p.input(dt)
	if moving := p.pad.KeyDown(utils.KeyLeft) || p.pad.KeyDown(utils.KeyRight); p.anim.State == vars.WalkTag && !moving {
		p.anim.SetState(vars.IdleTag, nil)
	}
}

func (p *Player) input(dt float64) { // TODO: refactor this
	actionPressed := p.pad.KeyPressedBuffered(utils.KeyAction, keyBufferDuration)
	healPressed := p.pad.KeyPressedBuffered(utils.KeyHeal, keyBufferDuration)
	if actor.PausingState(p) && p.anim.State != vars.ConsumeTag {
		return
	}
	if actionPressed {
		actor.Act(p, vars.AttackTag, playerDamage, playerDamage, p.reactForce, p.attackPushForce)
	}
	if healPressed {
		actor.Heal(p, playerHealFrame, playerHeal)
	}
	if p.pad.KeyDown(utils.KeyGuard) {
		actor.ShieldUp(p)
	}
	if p.pad.KeyReleased(utils.KeyGuard) {
		actor.ShieldDown(p)
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

	if !actor.BlockingState(p) {
		p.anim.FlipX = flip
	}
	if p.pad.KeyPressed(utils.KeyJump) && actor.CanJump(p) {
		actor.ClimbOff(p)
		p.body.Vy = -p.jumpSpeed
	}

	// TODO: Debug, remove later.
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		p.stats.Heal = p.stats.MaxHeal
	}
}

func (p *Player) inputClimbing(dt float64) {
	if p.pad.KeyDown(utils.KeyUp) || p.pad.KeyDown(utils.KeyDown) {
		actor.ClimbOn(p, p.pad.KeyDown(utils.KeyDown))
	}
	if p.anim.State != vars.ClimbTag {
		return
	}
	/*if !p.body.OnLadder {
		p.Control.ClimbOff()
	}*/
	p.body.Vy = 0
	speed := p.speed * 5 * dt
	if p.pad.KeyDown(utils.KeyUp) {
		p.body.Vy = -speed
	}
	if p.pad.KeyDown(utils.KeyDown) {
		p.body.Vy = speed
	}
}
