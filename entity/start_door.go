package entity

import (
	"game/comps/render"
	"game/core"
	"game/vars"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

var backgroundColor = color.RGBA{50, 60, 57, 255}

func init() { core.RegisterEntityName("StartDoor", NewStartDoor) }

type StartDoor struct {
	*core.BaseEntity
	render *render.Comp
}

func NewStartDoor(x, y, w, h float64, _ *core.Properties) *StartDoor {
	image := ebiten.NewImage(int(w), int(h))
	image.Fill(backgroundColor)

	door := &StartDoor{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: w, H: h},
		render:     &render.Comp{Image: image},
	}
	door.Add(door.render)
	time.AfterFunc(2*time.Second, func() {
		vars.World.Camera.Shake(0.1, 0.1)
		for range 10 {
			vars.World.Add(NewSmoke(door))
		}
		vars.World.Remove(door)
	})

	return door
}

func (sd *StartDoor) Init() {}

func (sd *StartDoor) Update(float64) {}
