package render

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func (c *Comp) IsActive() bool        { return c.active }
func (c *Comp) SetActive(active bool) { c.active = active }

type Comp struct {
	active bool
	Image  *ebiten.Image
	X, Y   float64
}

func (c *Comp) Draw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{GeoM: enitiyPos}
	op.GeoM.Translate(c.X, c.Y)
	screen.DrawImage(c.Image, op)
}
