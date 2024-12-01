package entity

import (
	"game/assets"
	"game/comps/render"
	"game/core"
	"game/vars"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

const (
	smokeSize         = 3
	smokeAnimDuration = 1.0
	smokeMaxDistance  = 40
)

var (
	smokeImage, _, _ = ebitenutil.NewImageFromFileSystem(assets.FS, "smoke.png")
)

type Smoke struct {
	*core.BaseEntity
	render           *render.Comp
	tween            *gween.Tween
	startX, startY   float64
	targetX, targetY float64
	timer            float64
	rotationSpeed    float64
}

func NewSmoke(from core.Entity) *Smoke {
	x, y, w, h := from.Rect()
	x += w * rand.Float64()
	y += h * rand.Float64()
	distance := rand.Float64() * smokeMaxDistance

	smoke := &Smoke{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: smokeSize, H: smokeSize},
		render:     &render.Comp{Image: smokeImage, R: rand.Float64() * 2 * math.Pi, Layer: 1},
		tween:      gween.New(0, 1, smokeAnimDuration, ease.OutCubic),
		timer:      smokeAnimDuration,
		startX:     x, startY: y,
		targetX: RandSignedFloat() * distance, targetY: RandSignedFloat() * distance,
		rotationSpeed: RandSignedFloat() * 4 * math.Pi,
	}
	smoke.Add(smoke.render)

	return smoke
}

func (s *Smoke) Init() {}

func (s *Smoke) Update(dt float64) {
	porg32, done := s.tween.Update(float32(dt))
	prog := float64(porg32)
	if done {
		defer vars.World.Remove(s)
	}

	s.X, s.Y = s.startX+prog*s.targetX, s.startY+prog*s.targetY
	s.render.R += s.rotationSpeed * dt

	alpha := uint8(100 + (math.MaxUint8-100)*(1-prog))
	s.render.ColorScale = color.RGBA{alpha, alpha, alpha, alpha}
}

func RandSignedFloat() float64 {
	return (rand.Float64() - 0.5) * 2 //nolint: mnd
}
