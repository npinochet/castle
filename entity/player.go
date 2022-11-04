package entity

import (
	"fmt"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/core"
	"game/libs/bump"
	"game/utils"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	playerAnimFile            = "assets/knight"
	playerWidth, playerHeight = 10, 11
)

func (p *Player) IsActive() bool        { return p.Active }
func (p *Player) SetActive(active bool) { p.Active = active }

type Player struct {
	*Actor
	pad       utils.ControlPack
	speed     float64
	jumpSpeed float64
}

func NewPlayer(x, y float64, props map[string]interface{}) *Player {
	body := &body.Comp{W: playerWidth, H: playerHeight}
	anim := &anim.Comp{FilesName: playerAnimFile, X: -2, Y: -3}

	player := &Player{
		Actor: NewActor(x, y, body, anim, nil, 20, 20),
		pad:   utils.NewControlPack(),
		speed: 350, jumpSpeed: 110,
	}
	player.AddComponent(player)
	player.Stats.Hud = true

	return player
}

func (p *Player) Init(entity *core.Entity) {
	p.Hitbox.HurtFunc = p.OnHurt
	hurtbox, _, _ := p.Anim.GetFrameHitboxes()
	p.Hitbox.PushHitbox(hurtbox.X, hurtbox.Y, hurtbox.W, hurtbox.H, false)
}

func (p *Player) Update(dt float64) {
	moving := p.control(dt)
	p.ManageAnim()
	if p.Anim.State == anim.WalkTag && !moving {
		p.Anim.SetState(anim.IdleTag)
	}
}

func (p *Player) DebugDraw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {
	hud := fmt.Sprintf("%0.2f/%0.2f/%0.2f", p.Stats.Health, p.Stats.Stamina, p.Stats.Poise)
	ebitenutil.DebugPrintAt(screen, hud, 35, 0)
}

func (p *Player) control(dt float64) bool {
	if p.Stats.Health <= 0 {
		return false
	}
	if p.Anim.State == anim.AttackTag || p.Anim.State == anim.StaggerTag {
		return false
	}

	if p.pad.KeyDown(utils.KeyGuard) {
		p.ShieldUp()
	}
	if p.pad.KeyReleased(utils.KeyGuard) {
		p.ShieldDown()
	}

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

	if p.Anim.State == anim.BlockTag {
		return moving
	}

	p.Anim.FlipX = flip
	if p.Body.Ground && p.pad.KeyPressed(utils.KeyUp) {
		p.Body.Vy = -p.jumpSpeed
	}
	if p.pad.KeyPressed(utils.KeyAction) {
		p.Attack()
	}

	return moving
}

func (p *Player) OnHurt(otherHc *hitbox.Comp, col *bump.Collision, damage float64) {
	p.Actor.Hurt(otherHc.Entity.X, damage, nil)
	p.World.Camera.Shake(0.5, 1)
}
