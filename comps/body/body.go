package body

import (
	"fmt"
	"game/assets"
	"game/core"
	"game/libs/bump"
	"game/utils"
	"game/vars"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

var DebugDraw = false

type Comp struct {
	Solid, Unmovable, Friction bool
	Ground                     bool
	Vx, Vy                     float64
	MaxX, MaxY                 float64
	Weight                     float64
	FilterOut                  []core.Entity
	entity                     core.Entity
	space                      *bump.Space
}

func (c *Comp) Init(entity core.Entity) {
	c.entity = entity
	if c.MaxX == 0 {
		c.MaxX = vars.DefaultMaxX
	}
	if c.MaxY == 0 {
		c.MaxY = vars.DefaultMaxY
	}
	c.Friction = true
	c.space = vars.World.Space

	c.space.Set(entity, bump.NewRect(c.entity.Rect()))
}

func (c *Comp) Update(dt float64) {
	if c.Solid {
		return
	}
	prevVx := c.Vx
	c.updateMovement(dt)
	c.Friction = !c.Ground || math.Abs(prevVx-c.Vx) >= 0
}

func (c *Comp) Draw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	if !DebugDraw {
		return
	}
	_, _, ew, eh := c.entity.Rect()
	image := ebiten.NewImage(int(ew), int(eh))
	image.Fill(color.RGBA{255, 0, 0, 100})
	screen.DrawImage(image, &ebiten.DrawImageOptions{GeoM: entityPos})

	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-5, -22)
	utils.DrawText(screen, fmt.Sprintf(`FRIC:%v`, c.Friction), assets.TinyFont, op)
}

func (c *Comp) updateMovement(dt float64) {
	if c.Friction || math.Abs(c.Vx) > c.MaxX {
		fric := vars.GroundFriction
		if !c.Ground {
			fric = vars.AirFriction
		}
		c.Vx -= c.Vx * fric * dt
		if math.Abs(c.Vx) < vars.FrictionEpsilon {
			c.Vx = 0
		}
	}
	c.Vy += vars.Gravity * c.Weight * dt
	c.Vy = math.Min(c.MaxY, math.Max(-c.MaxY, c.Vy))

	ex, ey := c.entity.Position()
	p := bump.Vec2{X: ex + c.Vx*dt, Y: ey + c.Vy*dt}
	goal, cols := c.space.Move(c, p, c.bodyFilter())
	c.entity.SetPosition(goal.X, goal.Y)

	if c.Unmovable {
		return
	}
	c.Ground = false
	for _, col := range cols {
		if col.Type == bump.Slide {
			if col.Normal.X != 0 {
				c.Vx = 0
			}
			if col.Normal.Y != 0 {
				c.Vy = 0
			}
			c.Ground = c.Ground || col.Normal.Y < 0
		}
		if _, ok := col.Other.(core.Entity); ok && col.Type == bump.Cross && col.Overlaps {
			c.applyOverlapForce(col)
		}
	}
}

func (c *Comp) applyOverlapForce(col *bump.Collision) {
	irect, orect := col.ItemRect, col.OtherRect
	overlap := (math.Min(irect.X+irect.W, orect.X+orect.W) - math.Max(irect.X, orect.X)) / math.Min(irect.W, orect.W)
	side := col.ItemRect.X + col.ItemRect.W/2 - (col.OtherRect.X + col.OtherRect.W/2)
	c.Vx = math.Min(vars.GroundFriction, math.Max(-vars.GroundFriction, c.Vx))
	c.Vx += math.Copysign(overlap*vars.CollisionStiffness, side) * vars.GroundFriction
}

func (c *Comp) bodyFilter() func(bump.Item, bump.Item) (bump.ColType, bool) {
	return func(_, other bump.Item) (bump.ColType, bool) {
		if entity, ok := other.(core.Entity); ok {
			if utils.Contains(c.FilterOut, entity) {
				return 0, false
			}

			return bump.Cross, true
		}

		return bump.Slide, true
	}
}
