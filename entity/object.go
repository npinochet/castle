package entity

import (
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/render"
	"game/core"
	"game/libs/bump"
	"game/vars"
	"log"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

type Object struct {
	*core.BaseEntity
	body                 *body.Comp
	hitbox               *hitbox.Comp
	render, renderNormal *render.Comp
	open                 bool
}

func init() { core.RegisterEntityName("Object", NewObject) }

func NewObject(x, y, w, h float64, _ *core.Properties) *Object {
	image, normalImage := constructTileImages(x, y, w, h)
	wall := &Object{
		BaseEntity:   &core.BaseEntity{X: x, Y: y, W: w, H: h},
		body:         &body.Comp{Tags: []bump.Tag{"object"}, FilterOut: nil},
		hitbox:       &hitbox.Comp{},
		render:       &render.Comp{Image: image},
		renderNormal: &render.Comp{Image: normalImage, Normal: true},
	}
	wall.Add(wall.body, wall.hitbox, wall.render, wall.renderNormal)

	return wall
}

func (o *Object) Init() {
	o.hitbox.HitFunc = o.hurt
	const shrink = 1
	o.hitbox.PushHitbox(bump.Rect{X: shrink, Y: shrink, W: o.W - shrink*2, H: o.H - shrink*2}, hitbox.Hit, nil)
}

func (o *Object) Update(_ float64) {}

func (o *Object) Opened() bool { return o.open }

func (o *Object) Open() {
	if o.open {
		return
	}
	o.open = true
	vars.World.Remove(o)
}

func (o *Object) hurt(other core.Entity, _ *bump.Collision, _ float64, _ hitbox.ContactType) {
	// TODO: Spawn debris
	for range 5 + rand.IntN(5) {
		vars.World.Add(NewSmoke(o))
	}
	vars.World.Remove(o)
}

func constructTileImages(x, y, w, h float64) (*ebiten.Image, *ebiten.Image) {
	image := ebiten.NewImage(int(w), int(h))
	normalImage := ebiten.NewImage(int(w), int(h))
	for ty := y; ty < y+h; ty += tileSize {
		for tx := x; tx < x+w; tx += tileSize {
			tiles, err := vars.World.Map.TilesFromPosition(tx, ty, true, vars.World.Space)
			if err != nil {
				log.Panic("fake wall: Failed to get tiles from position: ", err)
			}
			op := &ebiten.DrawImageOptions{}
			tile := tiles[vars.PipelineScreenTag]
			var sx, sy, dx, dy float64 = 1, 1, 0, 0
			if tile.FlipR {
				op.GeoM.Rotate(math.Pi / 2)
				sx = -1
			}
			if tile.FlipX {
				sx, dx = -1, tileSize
				if tile.FlipR {
					sx = 1
				}
			}
			if tile.FlipY {
				sy, dy = -1, tileSize
			}
			op.GeoM.Scale(sx, sy)
			op.GeoM.Translate(tx-x+dx, ty-y+dy)
			image.DrawImage(tile.Image, op)
			normalImage.DrawImage(tiles[vars.PipelineNormalMapTag].Image, op)
		}
	}

	return image, normalImage
}
