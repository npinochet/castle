package entity

import (
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/core"
	"game/libs/bump"
	"game/utils"
	"math"
	"time"
)

const (
	playerAnimFile                                 = "assets/knight"
	playerWidth, playerHeight                      = 8, 11
	playerOffsetX, playerOffsetY, playerOffsetFlip = -4, -3, 5
	playerSpeed, playerJumpSpeed                   = 350, 110
	playerDamage                                   = 20
)

var player *Player

type Player struct {
	*Actor
	pad       utils.ControlPack
	speed     float64
	jumpSpeed float64
}

func NewPlayer(x, y float64, props map[string]interface{}) *Player {
	body := &body.Comp{}
	anim := &anim.Comp{FilesName: playerAnimFile, OX: playerOffsetX, OY: playerOffsetY, OXFlip: playerOffsetFlip}

	playerActor := &Player{
		Actor: NewActor(x, y, playerWidth, playerHeight, body, anim, nil, playerDamage, playerDamage),
		pad:   utils.NewControlPack(),
		speed: playerSpeed, jumpSpeed: playerJumpSpeed,
	}
	playerActor.AddComponent(playerActor)
	playerActor.Stats.Hud = true
	playerActor.Stats.NoDebug = true

	player = playerActor

	return player
}

func (p *Player) Init(entity *core.Entity) {
	p.Hitbox.HurtFunc = p.OnHurt
	hurtbox, err := p.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic(err)
	}
	p.Hitbox.PushHitbox(hurtbox, false)
}

func (p *Player) Update(dt float64) {
	moving := p.control(dt)
	p.ManageAnim([]string{anim.AttackTag})
	if p.Anim.State == anim.WalkTag && !moving {
		p.Anim.SetState(anim.IdleTag)
	}
}

func (p *Player) control(dt float64) bool { // TODO: refactor this
	actionPressed := p.pad.KeyPressedBuffered(utils.KeyAction, 500*time.Millisecond)
	if p.Stats.Health <= 0 || p.Anim.State == anim.AttackTag || p.Anim.State == anim.StaggerTag {
		return false
	}

	p.controlClimbing(dt)
	p.controlBlocking()

	moving, flip := false, p.Anim.FlipX
	if p.pad.KeyDown(utils.KeyLeft) {
		if math.Abs(p.Body.Vx) <= p.Body.MaxX {
			p.Body.Vx -= p.speed * dt
		}
		moving, flip = true, true
	}
	if p.pad.KeyDown(utils.KeyRight) {
		if math.Abs(p.Body.Vx) <= p.Body.MaxX {
			p.Body.Vx += p.speed * dt
		}
		moving, flip = true, false
	}

	if p.Anim.State != anim.BlockTag {
		if actionPressed {
			p.Attack(anim.AttackTag)
		}
		p.Anim.FlipX = flip
		if (p.Body.Ground || p.Anim.State == anim.ClimbTag) && p.pad.KeyPressed(utils.KeyJump) {
			p.Body.Vy = -p.jumpSpeed
			p.ResetState()
		}
	}

	return moving
}

func (p *Player) controlClimbing(dt float64) {
	if (p.pad.KeyDown(utils.KeyUp) || p.pad.KeyDown(utils.KeyDown)) && p.Body.TouchLadder {
		p.ClimbOn(p.pad.KeyDown(utils.KeyDown))
	}
	if p.Anim.State != anim.ClimbTag {
		return
	}
	if !p.Body.OnLadder {
		p.ResetState()
	}
	p.Body.Vy = 0
	speed := p.speed * 5
	if p.pad.KeyDown(utils.KeyUp) {
		p.Body.Vy -= speed * dt
	}
	if p.pad.KeyDown(utils.KeyDown) {
		p.Body.Vy += speed * dt
	}
}

func (p *Player) controlBlocking() {
	if p.pad.KeyDown(utils.KeyGuard) {
		p.ShieldUp()
	}
	if p.pad.KeyReleased(utils.KeyGuard) {
		p.ShieldDown()
	}
}

func (p *Player) OnHurt(otherHc *hitbox.Comp, col *bump.Collision, damage float64) {
	p.Actor.Hurt(otherHc.Entity.X, damage, nil)
	p.World.Camera.Shake(0.5, 1)
}
