package entity

import (
	"game/comp"
	"game/core"
	"game/libs/bump"
	"game/utils"
	"time"
)

const knightAnimFile = "assets/knight"

func (k *Knight) IsActive() bool        { return k.Active }
func (k *Knight) SetActive(active bool) { k.Active = active }

type Knight struct {
	*Actor
	speed  float64
	player *core.Entity
}

func NewKnight(x, y float64, props map[string]interface{}) *core.Entity {
	anim := &comp.AsepriteComponent{FilesName: knightAnimFile,
		Fsm: &comp.AnimFsm{
			Transitions: map[string]string{"Walk": "Idle", "Attack": "Idle", "Stagger": "Idle"},
			ExitCallbacks: map[string]func(*comp.AsepriteComponent){
				"Stagger": func(ac *comp.AsepriteComponent) { ac.Data.PlaySpeed = 1 },
			},
		},
	}
	// hurtbox, _, _ := anim.GetFrameHitboxes().
	body := &comp.BodyComponent{W: 10, H: 14}

	knight := &Knight{
		Actor: NewActor(x, y, body, anim, &comp.StatsComponent{MaxPoise: 10}),
		speed: 50,
	}
	knight.AddComponent(knight)

	return &knight.Entity
}

func (k *Knight) Init(entity *core.Entity) {
	k.hitbox.HurtFunc, k.hitbox.BlockFunc = k.Hurt, k.Block
	hurtbox, _, _ := k.anim.GetFrameHitboxes()
	k.hitbox.PushHitbox(hurtbox.X, hurtbox.Y, hurtbox.W, hurtbox.H, false)
	k.player = k.World.GetEntityByID(utils.PlayerUID)
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
	k.Actor.Attack("Attack", 20, 20)
	time.AfterFunc(2*time.Second, func() { k.Attack() })
}

func (k *Knight) Stagger(force float64) {
	k.Actor.Stagger("Stagger", force)
}

func (k *Knight) ShieldUp() {
	if k.anim.State == "Block" {
		return
	}
	k.anim.SetState("Block")
	_, _, blockbox := k.anim.GetFrameHitboxes()
	k.speed /= 1.2
	k.body.MaxX /= 2
	k.stats.StaminaRecoverRate /= 2
	k.hitbox.PushHitbox(blockbox.X, blockbox.Y, blockbox.W, blockbox.H, true)
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

func (k *Knight) Hurt(otherHc *comp.HitboxComponent, col bump.Collision, damage float64) {
	if k.anim.State == "Block" {
		k.ShieldDown()
	}
	k.Actor.Hurt(*otherHc.EntX, damage, func(force float64) {
		k.Stagger(force)
	})
}

func (k *Knight) Block(otherHc *comp.HitboxComponent, col bump.Collision, damage float64) {
	k.Actor.Block(*otherHc.EntX, damage, func(force float64) {
		k.ShieldDown()
		k.Stagger(force)
		k.anim.Data.PlaySpeed = 0.5 // double time stagger.
	})
}
