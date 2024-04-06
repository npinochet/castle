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
	playerAnimFile                                 = "assets/knight"
	playerWidth, playerHeight                      = 8, 11
	playerOffsetX, playerOffsetY, playerOffsetFlip = -10, -3, 17
	playerMaxX, playerSpeed, playerJumpSpeed       = 60, 350, 110
	playerDamage, playerHeal                       = 20, 20
	playerHealFrame                                = 3

	keyBufferDuration = 500 * time.Millisecond
)

var PlayerRef *Player

type Player struct {
	*core.BaseEntity
	*actor.Control
	anim   *anim.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	stats  *stats.Comp
	ai     *ai.Comp

	actionTags                  []string
	speed, jumpSpeed            float64
	reactForce, attackPushForce float64

	pad utils.ControlPack
}

func NewPlayer(x, y float64, actionTags []string) *Player {
	p := &Player{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: playerWidth, H: playerHeight},
		anim:       &anim.Comp{FilesName: playerAnimFile, OX: playerOffsetX, OY: playerOffsetY, OXFlip: playerOffsetFlip},
		body:       &body.Comp{MaxX: playerMaxX},
		hitbox:     &hitbox.Comp{},
		stats:      &stats.Comp{Hud: true, NoDebug: true, Stamina: 65},
		ai:         &ai.Comp{},

		attackPushForce: vars.DefaultAttackPushForce,
		reactForce:      vars.DefaultReactForce,
		actionTags:      actionTags,
		speed:           playerSpeed, jumpSpeed: playerJumpSpeed,

		pad: utils.NewControlPack(),
	}
	p.Add(p.anim, p.body, p.hitbox, p.stats)
	p.Control = actor.NewControl(p)

	return p
}

func (p *Player) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return p.anim, p.body, p.hitbox, p.stats, p.ai
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
	if p.stats.Health <= 0 {
		p.Remove()

		return
	}
	p.input(dt)
	p.stats.Pause = p.PausingState()
	if state := p.anim.State; state == vars.IdleTag || state == vars.WalkTag {
		nextState := vars.IdleTag
		if p.body.Vx != 0 {
			nextState = vars.WalkTag
		}
		p.anim.SetState(nextState)
	}
	if moving := p.pad.KeyDown(utils.KeyLeft) || p.pad.KeyDown(utils.KeyRight); !moving && p.anim.State == vars.WalkTag {
		p.anim.SetState(vars.IdleTag)
	}
}

func (p *Player) input(dt float64) { // TODO: refactor this
	actionPressed := p.pad.KeyPressedBuffered(utils.KeyAction, keyBufferDuration)
	healPressed := p.pad.KeyPressedBuffered(utils.KeyHeal, keyBufferDuration)
	if p.PausingState() && p.anim.State != vars.ConsumeTag {
		return
	}
	if actionPressed {
		p.Attack(vars.AttackTag, playerDamage, playerDamage, p.reactForce, p.attackPushForce)
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
