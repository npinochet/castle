package particle

import (
	"game/comps/body"
	"game/comps/render"
	"game/comps/stats"
	"game/core"
	"game/libs/bump"
	"game/vars"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

const (
	flakeImageFile                       = "assets/flake.png"
	flakeSize                            = 3
	flakeSpawnMinTime, flakeSpawnMaxTime = 500 * time.Millisecond, 1000 * time.Millisecond
	flakeSpawnMinX, flakeSpawnMaxX       = -100, 100
	flakeSpawnMinY, flakeSpawnMaxY       = -50, -100
)

var flakeImage *ebiten.Image

func init() {
	var err error
	flakeImage, _, err = ebitenutil.NewImageFromFile(flakeImageFile)
	if err != nil {
		panic(err)
	}
}

type Flake struct {
	*core.BaseEntity
	body                     *body.Comp
	render                   *render.Comp
	captureTween             *gween.Tween
	target                   core.Entity
	startX, startY           float64
	randTargetW, randTargetH float64
}

func NewFlake(x, y float64) *Flake {
	flake := &Flake{
		BaseEntity:  &core.BaseEntity{X: x, Y: y, W: flakeSize, H: flakeSize},
		body:        &body.Comp{Tags: []bump.Tag{}, QueryTags: []bump.Tag{"map"}},
		render:      &render.Comp{Image: flakeImage},
		target:      vars.Player,
		randTargetW: rand.Float64(), randTargetH: rand.Float64(),
	}
	flake.Add(flake.body, flake.render)

	return flake
}

func (f *Flake) Init() {
	vx := flakeSpawnMinX + rand.Float64()*(flakeSpawnMaxX-flakeSpawnMinX)
	vy := flakeSpawnMinY + rand.Float64()*(flakeSpawnMaxY-flakeSpawnMinY)
	f.body.Vx, f.body.Vy = vx, vy
	captureTime := float64(flakeSpawnMinTime) + rand.Float64()*float64(flakeSpawnMaxTime-flakeSpawnMinTime)
	time.AfterFunc(time.Duration(captureTime), func() {
		f.captureTween = gween.New(0, 1, 0.8, ease.InQuad)
		f.startX, f.startY = f.X, f.Y
	})
}

func (f *Flake) Update(dt float64) {
	if f.captureTween == nil {
		return
	}

	tx, ty, tw, th := f.target.Rect()
	distX, distY := tx+tw*f.randTargetW-f.startX, ty+th*f.randTargetH-f.startY
	path, done := f.captureTween.Update(float32(dt))
	f.X = f.startX + float64(path)*distX
	f.Y = f.startY + float64(path)*distY
	if done {
		f.captureTween = nil
		if f.body != nil {
			vars.World.Space.Remove(f)
		}
		vars.World.Remove(f)
		core.Get[*stats.Comp](f.target).AddExp(1)
	}
}
