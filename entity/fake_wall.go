package entity

import (
	"game/comps/body"
	"game/comps/render"
	"game/core"
	"math"
	"math/rand"
)

type FakeWall struct {
	*core.BaseEntity
	body   *body.Comp
	render *render.Comp
}

// TODO: finish this
func NewFakeWall(x, y, _, _ float64, props *core.Properties) *FakeWall {
	wall := &FakeWall{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: tileSize, H: tileSize},
		body:       &body.Comp{Solid: true},
		render:     &render.Comp{Image: smokeImage, R: rand.Float64() * 2 * math.Pi, Layer: 1},
	}
	wall.Add(wall.render)

	return wall
}

func (fw *FakeWall) Init() {}

func (fw *FakeWall) Update(dt float64) {

}
