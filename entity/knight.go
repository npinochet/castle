package entity

import (
	"fmt"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/stats"
	"game/core"
	"game/utils"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const knightAnimFile = "assets/knight"

func (k *Knight) IsActive() bool        { return k.Active }
func (k *Knight) SetActive(active bool) { k.Active = active }

type Knight struct {
	*Actor
	player *core.Entity
}

func NewKnight(x, y float64, props map[string]interface{}) *core.Entity {
	speed := 180.0
	anim := &anim.Comp{FilesName: knightAnimFile, X: -2, Y: -3}
	body := &body.Comp{W: 10, H: 11, MaxX: 35}

	knight := &Knight{
		Actor: NewActor(x, y, body, anim, &stats.Comp{MaxPoise: 25}, 20, 20),
	}
	knight.speed = speed
	knight.AI = NewDefaultAI(knight.Actor, nil)
	knight.AddComponent(knight.AI)
	knight.AddComponent(knight)

	return &knight.Entity
}

func (k *Knight) Init(entity *core.Entity) {
	hurtbox, _, _ := k.Anim.GetFrameHitboxes()
	k.Hitbox.PushHitbox(hurtbox.X, hurtbox.Y, hurtbox.W, hurtbox.H, false)
	k.player = k.World.GetEntityByID(utils.PlayerUID)
	k.AI.Target = k.player
}

func (k *Knight) Update(dt float64) {
	k.ManageAnim()
	if k.Anim.State == anim.WalkTag && k.speed == 0 {
		k.Anim.SetState(anim.IdleTag)
	}
	if k.Anim.State == anim.WalkTag || k.Anim.State == anim.IdleTag {
		k.Anim.FlipX = k.player.X < k.X
	}
	if k.Anim.State != anim.AttackTag && k.Anim.State != anim.StaggerTag {
		if k.Anim.FlipX {
			k.Body.Vx -= k.speed * dt
		} else {
			k.Body.Vx += k.speed * dt
		}
	}

	if k.Stats.Health <= 0 {
		k.World.RemoveEntity(k.ID) // creats recursion loop
	}
}

func (k *Knight) DebugDraw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(`State: %s`, k.AI.Fsm.State), 0, 10)
}
