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
	*Actor
	pad       utils.ControlPack
	speed     float64
	jumpSpeed float64
}

func NewPlayer(x, y float64, props map[string]interface{}) *Player {
	animc := &anim.Comp{FilesName: playerAnimFile, OX: playerOffsetX, OY: playerOffsetY, OXFlip: playerOffsetFlip}
	bodyc := &body.Comp{MaxX: playerMaxX, Team: body.PlayerTeam}
	player := &Player{
		Actor: NewActor(x, y, playerWidth, playerHeight, []string{anim.AttackTag}, animc, bodyc, nil),
		pad:   utils.NewControlPack(),
		speed: playerSpeed, jumpSpeed: playerJumpSpeed,
	}
	player.AddComponent(player)
	player.Stats.Hud = true
	player.Stats.NoDebug = true
	player.Stats.MaxStamina, player.Stats.Stamina = 65, 65
	PlayerRef = player

	return PlayerRef
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
	p.control(dt)
	p.ManageAnim()
	if moving := p.pad.KeyDown(utils.KeyLeft) || p.pad.KeyDown(utils.KeyRight); p.Anim.State == anim.WalkTag && !moving {
		p.Anim.SetState(anim.IdleTag)
	}
}

func (p *Player) control(dt float64) { // TODO: refactor this
	actionPressed := p.pad.KeyPressedBuffered(utils.KeyAction, keyBufferDuration)
	healPressed := p.pad.KeyPressedBuffered(utils.KeyHeal, keyBufferDuration)
	if p.pausedState() && p.Anim.State != anim.ConsumeTag {
		return
	}
	if actionPressed {
		p.Attack(anim.AttackTag, playerDamage, playerDamage)
	}
	if healPressed {
		p.Heal(playerHealFrame, playerHeal)
	}
	p.controlBlocking()
	p.controlClimbing(dt)

	flip := p.Anim.FlipX
	if p.pad.KeyDown(utils.KeyLeft) {
		if math.Abs(p.Body.Vx) <= p.Body.MaxX {
			p.Body.Vx -= p.speed * dt
		}
		flip = false
	}
	if p.pad.KeyDown(utils.KeyRight) {
		if math.Abs(p.Body.Vx) <= p.Body.MaxX {
			p.Body.Vx += p.speed * dt
		}
		flip = true
	}

	if p.Anim.State != anim.BlockTag && p.Anim.State != anim.ParryBlockTag {
		p.Anim.FlipX = flip
	}
	if p.pad.KeyPressed(utils.KeyJump) && p.canJump() {
		p.ClimbOff()
		p.Body.Vy = -p.jumpSpeed
	}

	// TODO: Debug, remove later
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		p.Stats.Heal = p.Stats.MaxHeal
	}
}

func (p *Player) canJump() bool {
	return (p.Anim.State == anim.ClimbTag || p.Body.Ground) &&
		p.Anim.State != anim.BlockTag &&
		p.Anim.State != anim.ParryBlockTag &&
		p.Anim.State != anim.ConsumeTag
}

func (p *Player) controlClimbing(dt float64) {
	if p.pad.KeyDown(utils.KeyUp) || p.pad.KeyDown(utils.KeyDown) {
		p.ClimbOn(p.pad.KeyDown(utils.KeyDown))
	}
	if p.Anim.State != anim.ClimbTag {
		return
	}
	if !p.Body.OnLadder {
		p.ClimbOff()
	}
	p.Body.Vy = 0
	speed := p.speed * 5 * dt
	if p.pad.KeyDown(utils.KeyUp) {
		p.Body.Vy = -speed
	}
	if p.pad.KeyDown(utils.KeyDown) {
		p.Body.Vy = speed
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
	p.Actor.Hurt(otherHc.Entity, damage, nil)
	p.World.Camera.Shake(0.5, 1)
	p.World.Freeze(0.1)
}
