package comp

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func (rc *RenderComponent) IsActive() bool        { return rc.active }
func (rc *RenderComponent) SetActive(active bool) { rc.active = active }

type RenderComponent struct {
	active bool
	Image  *ebiten.Image
	X, Y   float64
}

func (rc *RenderComponent) Draw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{GeoM: enitiyPos}
	op.GeoM.Translate(rc.X, rc.Y)
	screen.DrawImage(rc.Image, op)
}
