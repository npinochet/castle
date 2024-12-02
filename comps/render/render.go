package render

import (
	"game/comps/anim"
	"game/core"
	"game/vars"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

type Comp struct {
	Image        *ebiten.Image
	X, Y         float64
	R            float64
	FlipX, FlipY bool
	Layer        int
	ColorScale   color.Color
	RollingTime  time.Duration
	Normal       bool
	rollingTimer *time.Timer
	w, h         float64
}

func (c *Comp) Init(_ core.Entity) {
	is := c.Image.Bounds().Size()
	w, h := is.X, is.Y
	c.w, c.h = float64(w), float64(h)
	if c.ColorScale == nil {
		c.ColorScale = color.White
	}
	if c.RollingTime != 0 {
		c.rollingTimer = time.AfterFunc(c.RollingTime, func() {
			c.R += math.Pi / 2
			if c.rollingTimer != nil {
				c.rollingTimer.Reset(c.RollingTime)
			}
		})
	}
}

func (c *Comp) Remove() {
	if c.rollingTimer != nil {
		c.rollingTimer.Stop()
	}
}

func (c *Comp) Update(_ float64) {}

func (c *Comp) Draw(pipeline *core.Pipeline, entityPos ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{}
	var sx, sy, dx, dy float64 = 1, 1, 0, 0
	if c.FlipX {
		sx, dx = -1, math.Floor(c.w)
	}
	if c.FlipY {
		sy, dy = -1, math.Floor(c.h)
	}
	op.GeoM.Scale(sx, sy)
	op.GeoM.Translate(-c.w/2, -c.h/2)
	op.GeoM.Rotate(c.R)
	op.GeoM.Translate(c.w/2, c.h/2)
	op.GeoM.Translate(c.X, c.Y)
	op.GeoM.Translate(dx, dy)
	op.GeoM.Concat(entityPos)
	op.ColorScale.ScaleWithColor(c.ColorScale)
	imageTag := vars.PipelineScreenTag
	if !c.Normal {
		normalOp := &colorm.DrawImageOptions{GeoM: op.GeoM}
		pipeline.Add(vars.PipelineNormalMapTag, c.Layer, func(normalMap *ebiten.Image) {
			colorm.DrawImage(normalMap, c.Image, anim.FillNormalMaskColorM, normalOp)
		})
	} else {
		imageTag = vars.PipelineNormalMapTag
	}
	pipeline.Add(imageTag, c.Layer, func(screen *ebiten.Image) { screen.DrawImage(c.Image, op) })
}
