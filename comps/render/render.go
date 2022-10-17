package render

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Comp struct {
	Image *ebiten.Image
	X, Y  float64
}

func (c *Comp) Draw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{GeoM: enitiyPos}
	op.GeoM.Translate(c.X, c.Y)
	screen.DrawImage(c.Image, op)
}
