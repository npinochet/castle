package entity

import (
	"fmt"
	"game/comp"
	"game/core"
	"game/libs/bump"
	"game/utils"
	"log"

	"github.com/damienfamed75/aseprite"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func (p *Player) IsActive() bool        { return p.Active }
func (p *Player) SetActive(active bool) { p.Active = active }

var (
	playerSprite              *ebiten.Image
	playerMetadata            *aseprite.File
	playerWidth, playerHeight float64 = 10, 14 // Image 14x14
)

func init() {
	var err error
	playerSprite, _, err = ebitenutil.NewImageFromFile("assets/knight.png")
	if err != nil {
		log.Fatal(err)
	}
	playerMetadata, err = aseprite.Open("assets/knight.json")
	if err != nil {
		log.Fatal(err)
	}
}

type Player struct {
	*Actor
	pad       utils.ControlPack
	speed     float64
	jumpSpeed float64
}

func NewPlayer(x, y float64, props map[string]interface{}) *Player {
	body := &comp.BodyComponent{W: playerWidth, H: playerHeight}
	anim := &comp.AsepriteComponent{
		X: -2, Y: 0, W: playerWidth + 4, H: playerHeight, Image: playerSprite, MetaData: playerMetadata,
		Fsm: &comp.AnimFsm{
			Transitions: map[string]string{"Walk": "Idle", "Attack": "Idle", "Stagger": "Idle"},
			ExitCallbacks: map[string]func(*comp.AsepriteComponent){
				"Stagger": func(ac *comp.AsepriteComponent) { ac.MetaData.PlaySpeed = 1 },
			},
		},
	}

	player := &Player{
		Actor:     NewActor(x, y, body, anim, nil),
		pad:       utils.NewControlPack(),
		speed:     350,
		jumpSpeed: 110,
	}
	player.Actor.reactForce = player.speed * 5
	player.AddComponent(player)
	return player
}

func (p *Player) Init(entity *core.Entity) {
	p.hitbox.HurtFunc, p.hitbox.BlockFunc = p.Hurt, p.Block
	p.hitbox.PushHitbox(0, 0, playerWidth, playerHeight, false)
}

func (p *Player) Update(dt float64) {
	moving := p.control(dt)
	p.ManageAnim("Idle", "Walk", "Attack", "Stagger")
	if p.anim.State == "Walk" && !moving {
		p.anim.SetState("Idle")
	}
}

func (p *Player) DebugDraw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {
	hud := fmt.Sprintf("%0.2f/%0.2f/%0.2f", p.stats.Health, p.stats.Stamina, p.stats.Poise)
	ebitenutil.DebugPrintAt(screen, hud, 0, 10)
}

func (p *Player) control(dt float64) bool {
	if p.anim.State == "Attack" || p.anim.State == "Stagger" {
		return false
	}

	if p.pad.KeyDown(utils.KeyGuard) {
		p.ShieldUp()
	}
	if p.pad.KeyReleased(utils.KeyGuard) {
		p.ShieldDown()
	}

	moving, flip := false, p.anim.FlipX
	if p.pad.KeyDown(utils.KeyLeft) {
		p.body.Vx -= p.speed * dt
		moving, flip = true, true
	}
	if p.pad.KeyDown(utils.KeyRight) {
		p.body.Vx += p.speed * dt
		moving, flip = true, false
	}

	if p.anim.State == "Block" {
		return moving
	}

	p.anim.FlipX = flip
	if p.body.Ground && p.pad.KeyPressed(utils.KeyUp) {
		p.body.Vy = -p.jumpSpeed
	}
	if p.pad.KeyPressed(utils.KeyAction) {
		p.Attack()
	}
	return moving
}

func (p *Player) Attack() {
	hitbox := bump.Rect{X: playerWidth, Y: 0, W: 10, H: playerHeight}
	if p.anim.FlipX {
		hitbox.X = -hitbox.W
	}
	p.Actor.Attack("Attack", hitbox, 1, 3, 20, 20)
}

func (p *Player) Stagger(force float64) {
	p.Actor.Stagger("Stagger", force*50)
}

func (p *Player) ShieldUp() {
	if p.anim.State == "Block" {
		return
	}
	p.anim.SetState("Block")
	p.speed /= 1.2
	p.body.MaxX /= 2
	p.stats.StaminaRecoverRate /= 2
	x, w := playerWidth-2, 4.0
	if p.anim.FlipX {
		x = -w + 2
	}
	p.hitbox.PushHitbox(x, 0, w, playerHeight, true)
}

func (p *Player) ShieldDown() {
	if p.anim.State != "Block" {
		return
	}
	p.anim.SetState("Idle")
	p.speed *= 1.2
	p.body.MaxX *= 2
	p.stats.StaminaRecoverRate *= 2
	p.hitbox.PopHitbox()
}

func (p *Player) Hurt(otherHc *comp.HitboxComponent, col bump.Colision, damage float64) {
	if p.anim.State == "Block" {
		p.ShieldDown()
	}
	p.Actor.Hurt(*otherHc.EntX, damage, func(force float64) {
		p.Stagger(force)
	})
	p.World.Camera.Shake(0.2, 1)
}

func (p *Player) Block(otherHc *comp.HitboxComponent, col bump.Colision, damage float64) {
	p.Actor.Block(*otherHc.EntX, damage, func(force float64) {
		p.ShieldDown()
		p.Stagger(force)
		p.anim.MetaData.PlaySpeed = 0.5 // double time stagger
	})
}
