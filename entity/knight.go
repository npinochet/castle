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

func (k *Knight) IsActive() bool        { return k.Active }
func (k *Knight) SetActive(active bool) { k.Active = active }

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
	*Actor
	speed  float64
	player *core.Entity
}

func NewKnight(x, y float64, props map[string]interface{}) *core.Entity {
	body := &comp.BodyComponent{W: knightWidth, H: knightHeight}
	anim := &comp.AsepriteComponent{X: -2, W: knightWidth + 4, H: knightHeight, Image: knightSprite, MetaData: knightMetadata,
		Fsm: &comp.AnimFsm{
			Transitions: map[string]string{"Walk": "Idle", "Attack": "Idle", "Stagger": "Idle"},
			ExitCallbacks: map[string]func(*comp.AsepriteComponent){
				"Stagger": func(ac *comp.AsepriteComponent) { ac.MetaData.PlaySpeed = 1 },
			},
		},
	}

	knight := &Knight{
		Actor: NewActor(x, y, body, anim, &comp.StatsComponent{MaxPoise: 10}),
		speed: 50,
	}
	knight.AddComponent(knight)
	return &knight.Entity
}

func (k *Knight) Init(entity *core.Entity) {
	k.hitbox.HurtFunc, k.hitbox.BlockFunc = k.Hurt, k.Block
	k.hitbox.PushHitbox(0, 0, knightWidth, knightHeight, false)
	k.player = k.World.GetEntityById(utils.Player)
	k.Attack()
}

func (k *Knight) Update(dt float64) {
	k.ManageAnim("Idle", "Walk", "Attack", "Stagger")
	if k.anim.State == "Idle" {
		k.anim.FlipX = k.player.X < k.X
		k.body.Vx = k.speed
		if k.anim.FlipX {
			k.body.Vx *= -1
		}
	}
}

func (k *Knight) Attack() {
	hitbox := bump.Rect{X: knightWidth, Y: 0, W: 10, H: knightHeight}
	if k.anim.FlipX {
		hitbox.X = -hitbox.W
	}
	k.Actor.Attack("Attack", hitbox, 1, 3, 20, 20)
	time.AfterFunc(2*time.Second, func() { k.Attack() })
}

func (k *Knight) Stagger(force float64) {
	k.Actor.Stagger("Stagger", force*10)
}

func (k *Knight) ShieldUp() {
	if k.anim.State == "Block" {
		return
	}
	k.anim.SetState("Block")
	k.speed /= 1.2
	k.body.MaxX /= 2
	k.stats.StaminaRecoverRate /= 2
	x, w := playerWidth-2, 4.0
	if k.anim.FlipX {
		x = -w + 2
	}
	k.hitbox.PushHitbox(x, 0, w, playerHeight, true)
}

func (k *Knight) ShieldDown() {
	if k.anim.State != "Block" {
		return
	}
	k.anim.SetState("Idle")
	k.speed *= 1.2
	k.body.MaxX *= 2
	k.stats.StaminaRecoverRate *= 2
	k.hitbox.PopHitbox()
}

func (k *Knight) Hurt(otherHc *comp.HitboxComponent, col bump.Colision, damage float64) {
	if k.anim.State == "Block" {
		k.ShieldDown()
	}
	k.Actor.Hurt(*otherHc.EntX, damage, func(force float64) {
		k.Stagger(force)
	})
}

func (k *Knight) Block(otherHc *comp.HitboxComponent, col bump.Colision, damage float64) {
	k.Actor.Block(*otherHc.EntX, damage, func(force float64) {
		k.ShieldDown()
		k.Stagger(force)
		k.anim.MetaData.PlaySpeed = 0.5 // double time stagger
	})
}
