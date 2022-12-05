package render

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Comp struct {
	Image *ebiten.Image
	X, Y  float64
}

func (c *Comp) Draw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(c.X, c.Y)
	screen.DrawImage(c.Image, op)
}
