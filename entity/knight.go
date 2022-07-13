package entity

import (
	"game/comp"
	"game/core"
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
	anim := &comp.AsepriteComponent{FilesName: knightAnimFile}
	body := &comp.BodyComponent{W: 10, H: 14}

	knight := &Knight{
		Actor: NewActor(x, y, body, anim, &comp.StatsComponent{MaxPoise: 10}, 20, 20),
		speed: 50,
	}
	knight.AddComponent(knight)

	return &knight.Entity
}

func (k *Knight) Init(entity *core.Entity) {
	hurtbox, _, _ := k.Anim.GetFrameHitboxes()
	k.Hitbox.PushHitbox(hurtbox.X, hurtbox.Y, hurtbox.W, hurtbox.H, false)
	k.player = k.World.GetEntityByID(utils.PlayerUID)
	k.Attack()
}

func (k *Knight) Update(dt float64) {
	k.ManageAnim()
	if k.Anim.State == animIdleTag {
		k.Anim.FlipX = k.player.X < k.X
		k.Body.Vx = k.speed
		if k.Anim.FlipX {
			k.Body.Vx *= -1
		}
	}
}

func (k *Knight) Attack() {
	k.Actor.Attack()
	time.AfterFunc(2*time.Second, func() { k.Attack() })
}
