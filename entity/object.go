package entity

import (
	"game/assets"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/render"
	"game/core"
	"game/ext"
	"game/libs/bump"
	"game/vars"
	"log"
	"math"
	"math/rand/v2"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Object struct {
	*core.BaseEntity
	body                 *body.Comp
	hitbox               *hitbox.Comp
	render, renderNormal *render.Comp
	reward               int
}

func init() { core.RegisterEntityName("Object", NewObject) }

func NewObject(x, y, w, h float64, props *core.Properties) *Object {
	image, normalImage := constructTileImages(x, y, w, h)
	dx, dy := x-math.Floor(x/tileSize)*tileSize, y-math.Floor(y/tileSize)*tileSize
	reward, _ := strconv.Atoi(props.Custom["reward"])
	object := &Object{
		BaseEntity:   &core.BaseEntity{X: x, Y: y, W: w, H: h},
		body:         &body.Comp{Tags: []bump.Tag{"object"}},
		hitbox:       &hitbox.Comp{},
		render:       &render.Comp{Image: image, X: -dx, Y: -dy},
		renderNormal: &render.Comp{Image: normalImage, X: -dx, Y: -dy, Normal: true},
		reward:       reward,
	}
	object.Add(object.body, object.hitbox, object.render, object.renderNormal)

	return object
}

func (o *Object) Init() {
	o.hitbox.HitFunc = o.hurt
	const shrink = 1
	o.hitbox.PushHitbox(bump.Rect{X: shrink, Y: shrink, W: o.W - shrink*2, H: o.H - shrink*2}, hitbox.Hit, nil)
}

func (o *Object) Update(_ float64) {
	for _, e := range ext.QueryItems(o, bump.Rect{X: o.X, Y: o.Y - 1, W: o.W, H: 1}, "object") {
		if e.Y+e.H <= o.Y && math.Abs(o.body.Vx) > math.Abs(e.body.Vx) {
			e.body.Vx = o.body.Vx
		}
	}
}

func (o *Object) hurt(other core.Entity, _ *bump.Collision, _ float64, _ hitbox.ContactType) {
	for range 5 + rand.IntN(5) {
		vars.World.Add(NewSmoke(o))
		vars.World.Add(NewDebris(o))
	}
	for range o.reward {
		vars.World.Add(NewFlake(o))
	}
	vars.World.Remove(o)
}

func constructTileImages(x, y, w, h float64) (*ebiten.Image, *ebiten.Image) {
	x, y = math.Floor(x/tileSize)*tileSize, math.Floor(y/tileSize)*tileSize
	w, h = math.Ceil(w/tileSize)*tileSize, math.Ceil(h/tileSize)*tileSize
	image := ebiten.NewImage(int(w), int(h))
	normalImage := ebiten.NewImage(int(w), int(h))
	for ty := y; ty < y+h; ty += tileSize {
		for tx := x; tx < x+w; tx += tileSize {
			tiles, err := vars.World.Map.TilesFromPosition(tx, ty, true, vars.World.Space)
			if err != nil {
				log.Panic("object: Failed to get tiles from position: ", err)
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

const debrisDuration = 5.0

var debrisImage, _, _ = ebitenutil.NewImageFromFileSystem(assets.FS, "debris.png")

type Debris struct {
	*core.BaseEntity
	body                     *body.Comp
	render                   *render.Comp
	from                     core.Entity
	randTargetW, randTargetH float64
	timer                    float64
	imageIndex               int
	rotationSpeed            float64
}

func NewDebris(from core.Entity) *Debris {
	imgW, imgH := debrisImage.Bounds().Dx(), debrisImage.Bounds().Dy()
	x, y, w, h := from.Rect()
	debris := &Debris{
		BaseEntity:  &core.BaseEntity{X: x + w/2, Y: y + h/2, W: float64(imgW), H: float64(imgH)},
		body:        &body.Comp{Tags: []bump.Tag{}, QueryTags: []bump.Tag{"map"}},
		render:      &render.Comp{Image: debrisImage},
		from:        from,
		randTargetW: rand.Float64(), randTargetH: rand.Float64(),
		timer:         debrisDuration * (0.5 + rand.Float64()),
		rotationSpeed: RandSignedFloat() * 4 * math.Pi,
	}
	debris.Add(debris.body, debris.render)

	return debris
}

func (d *Debris) Init() {
	vx := flakeSpawnMinX + rand.Float64()*(flakeSpawnMaxX-flakeSpawnMinX)
	vy := flakeSpawnMinY + rand.Float64()*(flakeSpawnMaxY-flakeSpawnMinY)
	d.body.Vx, d.body.Vy = vx, vy
}

func (d *Debris) Update(dt float64) {
	if d.body.Ground {
		d.rotationSpeed *= 0.98 * dt
	}
	d.render.R += d.rotationSpeed * dt
	if d.timer -= dt; d.timer <= 0 {
		vars.World.Remove(d)
	}
}
