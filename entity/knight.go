package entity

import (
	"game/comp"
	"game/core"
	"game/libs/bump"
	"game/utils"
	"log"
	"time"

	"github.com/damienfamed75/aseprite"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func (p *Knight) IsActive() bool        { return p.Entity.Active }
func (p *Knight) SetActive(active bool) { p.Entity.Active = active }

var (
	knightSprite              *ebiten.Image
	knightMetadata            *aseprite.File
	knightWidth, knightHeight float64 = 10, 14 // Image 14x14
)

func init() {
	var err error
	knightSprite, _, err = ebitenutil.NewImageFromFile("assets/knight.png")
	if err != nil {
		log.Fatal(err)
	}
	knightMetadata, err = aseprite.Open("assets/knight.json")
	if err != nil {
		log.Fatal(err)
	}
}

type Knight struct {
	core.Entity
	body   *comp.BodyComponent
	hitbox *comp.HitboxComponent
	anim   *comp.AsepriteComponent
	speed  float64
}

func NewKnight(x, y float64, props map[string]interface{}) *core.Entity {
	knight := &Knight{
		Entity: core.Entity{X: x, Y: y},
		body:   &comp.BodyComponent{W: knightWidth, H: knightHeight},
		hitbox: &comp.HitboxComponent{},
		anim: &comp.AsepriteComponent{X: -2, W: knightWidth + 4, H: knightHeight, Image: knightSprite, MetaData: knightMetadata, Fsm: &comp.AnimFsm{
			Transitions: map[string]string{"Walk": "Idle", "Attack": "Idle", "Stagger": "Idle"},
		}},
		speed: 50,
	}
	knight.AddComponent(knight.body, knight.hitbox, knight.anim)
	knight.AddComponent(knight)
	return &knight.Entity
}

func (p *Knight) Init(entity *core.Entity) {
	p.hitbox.HurtFunc, p.hitbox.BlockFunc = p.KnightHurt, p.KnightBlock
	p.hitbox.PushHitbox(0, 0, knightWidth, knightHeight, false)
	p.Attack(1.0 / 60)
}

func (p *Knight) Update(dt float64) {
	if p.anim.State == "Idle" {
		p.Walk(dt)
	}

	p.processAccion(dt)
}

func (p *Knight) processAccion(dt float64) {
	if p.anim.State == "Attack" || p.anim.State == "Stagger" {
		p.body.Vx = 0
		return
	}

	moving := p.anim.State == "Walk"
	p.body.Friction = !moving
}

func (p *Knight) Walk(dt float64) {
	p.anim.SetState("Walk")

	player := p.World.GetEntityById(utils.Player)
	right := player.X < p.X
	p.anim.FlipX = right
	p.body.Vx = p.speed
	if right {
		p.body.Vx *= -1
	}
}

func (p *Knight) Attack(dt float64) {
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
			x, w := knightWidth, 10.0
			if p.anim.FlipX {
				x = -w
			}
			if p.hitbox.Hit(x, 0, w, knightHeight) {
				p.Stagger(dt, force)
			}
		}
	})
	time.AfterFunc(2*time.Second, func() { p.Attack(dt) })
}

func (p *Knight) Stagger(dt float64, force float64) {
	p.anim.SetState("Stagger")
	p.body.Vx = -force * 10 * dt
}

func (p *Knight) ShieldUp(dt float64) {
	p.anim.SetState("Block")
	p.speed /= 1.2
	p.body.MaxX /= 2
	x, w := knightWidth, 2.0
	if p.anim.FlipX {
		x = -w
	}
	p.hitbox.PushHitbox(x, 0, w, knightHeight, true)
}

func (p *Knight) ShieldDown(dt float64) {
	p.anim.SetState("Idle")
	p.speed *= 1.2
	p.body.MaxX *= 2
	p.hitbox.PopHitbox()
}

func (p *Knight) KnightHurt(otherHc *comp.HitboxComponent, col bump.Colision) {
	dt := 1.0 / 60
	force := 5 * p.speed
	if *p.hitbox.EntX > *otherHc.EntX {
		force *= -1
	}
	p.Stagger(dt, force)
}

func (p *Knight) KnightBlock(otherHc *comp.HitboxComponent, col bump.Colision) {
	p.Stagger(1.0/60, 0)
}
