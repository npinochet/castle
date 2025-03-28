package entity

import (
	"game/assets"
	"game/comps/body"
	"game/comps/render"
	"game/comps/stats"
	"game/core"
	"game/libs/bump"
	"game/vars"
	"image"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

const (
	flakeSize                            = 3
	flakeAnimDuration                    = 0.2
	flakeSpawnMinTime, flakeSpawnMaxTime = 500 * time.Millisecond, 1000 * time.Millisecond
	flakeSpawnMinX, flakeSpawnMaxX       = -100, 100
	flakeSpawnMinY, flakeSpawnMaxY       = -50, -100
)

var (
	flakeImage, _, _ = ebitenutil.NewImageFromFileSystem(assets.FS, "flake.png")
	flakeImages      = []*ebiten.Image{
		flakeImage.SubImage(image.Rect(0, 0, flakeSize, flakeSize)).(*ebiten.Image),
		flakeImage.SubImage(image.Rect(flakeSize, 0, flakeSize*2, flakeSize*2)).(*ebiten.Image),
	}
)

type Flake struct {
	*core.BaseEntity
	body                     *body.Comp
	render                   *render.Comp
	captureTween             *gween.Tween
	from, target             core.Entity
	startX, startY           float64
	randTargetW, randTargetH float64
	timer                    float64
	imageIndex               int
}

func NewFlake(from core.Entity) *Flake {
	x, y, w, h := from.Rect()
	flake := &Flake{
		BaseEntity:  &core.BaseEntity{X: x + w/2, Y: y + h/2, W: flakeSize, H: flakeSize},
		body:        &body.Comp{Tags: []bump.Tag{}, QueryTags: []bump.Tag{"map"}},
		render:      &render.Comp{Image: flakeImages[0]},
		from:        from,
		target:      vars.Player,
		randTargetW: rand.Float64(), randTargetH: rand.Float64(),
		timer: rand.Float64() * flakeAnimDuration,
	}
	flake.Add(flake.body, flake.render)

	return flake
}

func (f *Flake) Init() {
	vx := flakeSpawnMinX + rand.Float64()*(flakeSpawnMaxX-flakeSpawnMinX)
	vy := flakeSpawnMinY + rand.Float64()*(flakeSpawnMaxY-flakeSpawnMinY)
	f.body.Vx, f.body.Vy = vx, vy
	captureTime := float64(flakeSpawnMinTime) + rand.Float64()*float64(flakeSpawnMaxTime-flakeSpawnMinTime)
	if f.from != f.target {
		time.AfterFunc(time.Duration(captureTime), func() {
			f.captureTween = gween.New(0, 1, 0.8, ease.InQuad)
			f.startX, f.startY = f.X, f.Y
		})
	}
}

func (f *Flake) Update(dt float64) {
	if f.timer += dt; f.timer >= flakeAnimDuration {
		f.timer = 0
		f.imageIndex = (f.imageIndex + 1) % len(flakeImages)
		f.render.Image = flakeImages[f.imageIndex]
	}
	if f.captureTween == nil {
		return
	}

	tx, ty, tw, th := f.target.Rect()
	distX, distY := tx+tw*f.randTargetW-f.startX, ty+th*f.randTargetH-f.startY
	path, done := f.captureTween.Update(float32(dt))
	f.X = f.startX + float64(path)*distX
	f.Y = f.startY + float64(path)*distY
	if done {
		vars.World.Remove(f)
		f.captureTween = nil
		core.Get[*stats.Comp](f.target).AddExp(1)
	}
}
