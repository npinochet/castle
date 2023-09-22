package entity

import (
	"game/actor"
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
	actor.Actor
	pad       utils.ControlPack
	jumpSpeed float64
}

func NewPlayer(x, y float64, _ map[string]any) *Player {
	player := &Player{
		Actor:     actor.NewActor(x, y, playerWidth, playerHeight, []string{actor.AttackTag}),
		pad:       utils.NewControlPack(),
		jumpSpeed: playerJumpSpeed,
	}
	PlayerRef = player
	player.Speed = playerSpeed

	player.Anim.FilesName = playerAnimFile
	player.Anim.OX, player.Anim.OY = playerOffsetX, playerOffsetY
	player.Anim.OXFlip = playerOffsetFlip

	player.Body.MaxX = playerMaxX
	player.Body.Team = actor.PlayerTeam

	player.Stats.Hud = true
	player.Stats.NoDebug = true
	player.Stats.MaxStamina, player.Stats.Stamina = 65, 65

	return PlayerRef
}

func (p *Player) Init() {
	p.Actor.Init()
	p.Hitbox.HurtFunc = p.OnHurt
}

func (p *Player) Update(dt float64) {
	p.Actor.Update(dt)
	p.control(dt)
	p.ManageAnim()
	if moving := p.pad.KeyDown(utils.KeyLeft) || p.pad.KeyDown(utils.KeyRight); p.Anim.State == actor.WalkTag && !moving {
		p.Anim.SetState(&p.Actor, actor.IdleTag)
	}
}

func (p *Player) control(dt float64) { // TODO: refactor this.
	actionPressed := p.pad.KeyPressedBuffered(utils.KeyAction, keyBufferDuration)
	healPressed := p.pad.KeyPressedBuffered(utils.KeyHeal, keyBufferDuration)
	if p.PausedState() && p.Anim.State != actor.ConsumeTag {
		return
	}
	if actionPressed {
		p.Attack(actor.AttackTag, playerDamage, playerDamage)
	}
	if healPressed {
		p.Heal(playerHealFrame, playerHeal)
	}
	p.controlBlocking()
	p.controlClimbing(dt)

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

	if p.Anim.State != actor.BlockTag && p.Anim.State != actor.ParryBlockTag {
		p.Anim.FlipX = flip
	}
	if p.pad.KeyPressed(utils.KeyJump) && p.canJump() {
		p.ClimbOff()
		p.Body.Vy = -p.jumpSpeed
	}

	// TODO: Debug, remove later.
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		p.Stats.Heal = p.Stats.MaxHeal
	}
}

func (p *Player) canJump() bool {
	return (p.Anim.State == actor.ClimbTag || p.Body.Ground) &&
		p.Anim.State != actor.BlockTag &&
		p.Anim.State != actor.ParryBlockTag &&
		p.Anim.State != actor.ConsumeTag
}

func (p *Player) controlClimbing(dt float64) {
	if p.pad.KeyDown(utils.KeyUp) || p.pad.KeyDown(utils.KeyDown) {
		p.ClimbOn(p.pad.KeyDown(utils.KeyDown))
	}
	if p.Anim.State != actor.ClimbTag {
		return
	}
	if !p.Body.OnLadder {
		p.ClimbOff()
	}
	p.Body.Vy = 0
	speed := p.Speed * 5 * dt
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

func (p *Player) OnHurt(other *actor.Actor, _ *bump.Collision, damage float64) {
	p.Actor.Hurt(other, damage, nil)
	p.World.Camera.Shake(0.5, 1)
}
