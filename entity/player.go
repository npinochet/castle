package entity

import (
	"game/comp"
	"game/core"
	"game/libs/bump"
	"game/utils"
	"log"

	"github.com/damienfamed75/aseprite"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func (p *Player) IsActive() bool        { return p.Entity.Active }
func (p *Player) SetActive(active bool) { p.Entity.Active = active }

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
	core.Entity
	body      *comp.BodyComponent
	hitbox    *comp.HitboxComponent
	anim      *comp.AsepriteComponent
	pad       utils.ControlPack
	speed     float64
	jumpSpeed float64
}

func NewPlayer(x, y float64, props map[string]interface{}) *Player {
	player := &Player{
		Entity: core.Entity{X: x, Y: y},
		body:   &comp.BodyComponent{W: playerWidth, H: playerHeight},
		hitbox: &comp.HitboxComponent{},
		anim: &comp.AsepriteComponent{X: -2, W: playerWidth + 4, H: playerHeight, Image: playerSprite, MetaData: playerMetadata, Fsm: &comp.AnimFsm{
			Transitions: map[string]string{"Walk": "Idle", "Attack": "Idle", "Block": "Block", "Stagger": "Idle"},
		}},
		pad:       utils.NewControlPack(),
		speed:     350,
		jumpSpeed: 110,
	}
	player.AddComponent(player.body, player.hitbox, player.anim)
	player.AddComponent(player)
	return player
}

func (p *Player) Init(entity *core.Entity) {
	p.hitbox.HurtFunc, p.hitbox.BlockFunc = p.PlayerHurt, p.PlayerBlock
	p.hitbox.PushHitbox(0, 0, playerWidth, playerHeight, false)
}

func (p *Player) Update(dt float64) {
	p.control(dt)
}

func (p *Player) control(dt float64) {
	if p.anim.State == "Attack" || p.anim.State == "Stagger" {
		return
	}

	if p.anim.State != "Block" && p.pad.KeyDown(utils.KeyGuard) {
		p.ShieldUp(dt)
	}
	if p.anim.State == "Block" && p.pad.KeyReleased(utils.KeyGuard) {
		p.ShieldDown(dt)
	}

	moving := false
	if p.anim.State == "Block" {
		if p.pad.KeyDown(utils.KeyLeft) {
			p.body.Vx -= p.speed * dt
			moving = true
		}
		if p.pad.KeyDown(utils.KeyRight) {
			p.body.Vx += p.speed * dt
			moving = true
		}
		p.body.Friction = !moving
		return
	}

	if p.pad.KeyDown(utils.KeyLeft) {
		p.body.Vx -= p.speed * dt
		moving = true
		p.anim.FlipX = true
	}
	if p.pad.KeyDown(utils.KeyRight) {
		p.body.Vx += p.speed * dt
		moving = true
		p.anim.FlipX = false
	}

	if p.body.Ground && p.pad.KeyPressed(utils.KeyUp) {
		p.body.Vy = -p.jumpSpeed
	}
	if p.pad.KeyPressed(utils.KeyAction) {
		moving = false
		p.Attack(dt)
	}

	p.body.Friction = !moving
	if moving {
		p.anim.SetState("Walk")
	} else if p.anim.State == "Walk" {
		p.anim.SetState("Idle")
	}
}

func (p *Player) Attack(dt float64) {
	p.anim.SetState("Attack")
	force := p.speed * 5
	if p.anim.FlipX {
		force *= -1
	}
	pushed := false
	p.anim.OnFrames(1, 3, func(frame int) {
		if frame == 1 && !pushed {
			p.body.Vx += force * dt
		} else {
			x, w := playerWidth, 10.0
			if p.anim.FlipX {
				x = -w
			}
			if p.hitbox.Hit(x, 0, w, playerHeight) {
				p.Stagger(dt, force)
			}
		}
	})
}

func (p *Player) Stagger(dt float64, force float64) {
	p.anim.SetState("Stagger")
	p.body.Vx = -force * 10 * dt
}

func (p *Player) ShieldUp(dt float64) {
	p.anim.SetState("Block")
	p.speed /= 1.2
	p.body.MaxX /= 2
	x, w := playerWidth, 2.0
	if p.anim.FlipX {
		x = -w
	}
	p.hitbox.PushHitbox(x, 0, w, playerHeight, true)
}

func (p *Player) ShieldDown(dt float64) {
	p.anim.SetState("Idle")
	p.speed *= 1.2
	p.body.MaxX *= 2
	p.hitbox.PopHitbox()
}

func (p *Player) PlayerHurt(otherHc *comp.HitboxComponent, col bump.Colision) {
	dt := 1.0 / 60
	force := 5 * p.speed
	p.Stagger(dt, force)
}

func (p *Player) PlayerBlock(otherHc *comp.HitboxComponent, col bump.Colision) {

}
