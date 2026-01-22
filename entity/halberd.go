package entity

import (
	"game/assets"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/render"
	"game/core"
	"game/entity/actor"
	"game/libs/bump"
	"game/vars"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	halberdSize                  = 5
	halberdDamage                = 5
	halberdWeight                = 0.6
	halberdMinVel, halberdMaxVel = 30.0, 100.0
	halberdRemoveTime            = 3
)

var (
	halberdImage, _, _ = ebitenutil.NewImageFromFileSystem(assets.FS, "halberd.png")
)

type Halberd struct {
	*core.BaseEntity
	render          *render.Comp
	body            *body.Comp
	hitbox          *hitbox.Comp
	owner           actor.Actor
	filterOutHitbox []*hitbox.Comp
	removeTimer     float64
}

func NewHalberd(x, y float64, owner actor.Actor) *Halberd {
	ownerAnim, _, ownerHitbox, _, _ := owner.Comps()
	vx, vy := -halberdMaxVel, 60.0
	if ownerAnim.FlipX {
		vx = -vx
	}

	halberd := &Halberd{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: halberdSize, H: halberdSize},
		render:     &render.Comp{Image: halberdImage, FlipX: ownerAnim.FlipX},
		body: &body.Comp{
			Weight: halberdWeight,
			Vx:     vx, Vy: -vy,
			MaxX:      halberdMaxVel,
			FilterOut: []core.Entity{owner},
		},
		hitbox:          &hitbox.Comp{},
		owner:           owner,
		filterOutHitbox: []*hitbox.Comp{ownerHitbox},
		removeTimer:     halberdRemoveTime,
	}
	halberd.Add(halberd.render, halberd.body, halberd.hitbox)

	return halberd
}

func (h *Halberd) Init() {
	h.body.Friction = false
}

func (h *Halberd) Update(dt float64) {
	_, h.filterOutHitbox = h.hitbox.HitFromHitBox(bump.Rect{H: halberdSize, W: halberdSize}, halberdDamage, h.filterOutHitbox)
	if h.body.Ground {
		h.body.Friction = true
		if h.removeTimer -= dt; h.removeTimer <= 0 {
			vars.World.Remove(h)
		}
	}
}

func (h *Halberd) Destroy() {}
