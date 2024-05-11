package game

import (
	"game/vars"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

var fadeImg *ebiten.Image

func init() {
	fadeImg = ebiten.NewImage(vars.ScreenWidth, vars.ScreenHeight)
	fadeImg.Fill(color.Black)
}

type RestartTransition struct {
	fadeTween *gween.Tween
}

func (t *RestartTransition) Init() {
	t.fadeTween = gween.New(0, 1, 3, ease.OutQuad)
}

func (t *RestartTransition) Update(dt float64) bool {
	if t == nil {
		return false
	}
	if _, done := t.fadeTween.Update(float32(dt)); done {
		Reset()

		return true
	}

	return false
}

func (t *RestartTransition) Draw(screen *ebiten.Image) {
	if t == nil {
		return
	}
	alpha, _ := t.fadeTween.Update(0)
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(fadeImg, op)
}
