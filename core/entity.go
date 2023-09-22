package core

import "github.com/hajimehoshi/ebiten/v2"

type CoreEntity struct {
	ID         uint64
	World      *World
	Active     bool
	X, Y, W, H float64
}

func (e *CoreEntity) Position() (float64, float64)               { return e.X, e.Y }
func (e *CoreEntity) Rect() (float64, float64, float64, float64) { return e.X, e.Y, e.W, e.H }

func (e *CoreEntity) Draw(screen *ebiten.Image) {
	if !e.Active {
		return
	}

	entityPos := ebiten.GeoM{}
	entityPos.Translate(e.X, e.Y)
	x, y := e.World.Camera.Position()
	entityPos.Translate(-x, -y)
}
