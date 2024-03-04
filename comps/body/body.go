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
	FilterOut                  []*Comp
	entity                     *core.Entity
	space                      *bump.Space
	prevVx                     float64
	debugQueryRect             bump.Rect
}

func (c *Comp) Init(entity *core.Entity) {
	c.entity = entity
	if c.MaxX == 0 {
		c.MaxX = vars.DefaultMaxX
	}
	if c.MaxY == 0 {
		c.MaxY = vars.DefaultMaxY
	}
	c.Friction = true
	c.space = entity.World.Space

	c.space.Set(c, bump.NewRect(c.entity.Rect()))
}

func (c *Comp) Update(dt float64) {
	if c.Solid {
		return
	}
	c.Friction = !c.Ground || math.Abs(c.prevVx-c.Vx) >= 0
	c.updateMovement(dt)
	c.prevVx = c.Vx
}

func (c *Comp) Query(rect bump.Rect, filter func(item bump.Item) bool) []*bump.Collision {
	c.debugQueryRect = rect

	return c.space.Query(rect, filter)
}

func (c *Comp) QueryEntites(rect bump.Rect) []*core.Entity {
	entityFilter := func(item bump.Item) bool {
		if comp, ok := item.(*Comp); ok {
			return comp != c
		}

		return false
	}

	cols := c.Query(rect, entityFilter)
	var ents []*core.Entity
	for _, c := range cols {
		if comp, ok := c.Other.(*Comp); ok {
			ents = append(ents, comp.entity)
		}
	}

	return ents
}

func (c *Comp) QueryFront(dist, height float64, lookingRight bool) []*core.Entity {
	rect := bump.Rect{X: -dist, Y: -height / 2, W: dist, H: height}
	if lookingRight {
		rect.X += dist
	}
	rect.X += c.entity.X
	rect.Y += c.entity.Y

	return c.QueryEntites(rect)
}

func (c *Comp) Draw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	if !DebugDraw {
		return
	}
	image := ebiten.NewImage(int(c.entity.W), int(c.entity.H))
	image.Fill(color.RGBA{255, 0, 0, 100})
	screen.DrawImage(image, &ebiten.DrawImageOptions{GeoM: entityPos})

	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-5, -22)
	utils.DrawText(screen, fmt.Sprintf(`FRIC:%v`, c.Friction), assets.TinyFont, op)

	if c.debugQueryRect.W != 0 || c.debugQueryRect.H != 0 {
		image := ebiten.NewImage(int(c.debugQueryRect.W), int(c.debugQueryRect.H))
		image.Fill(color.RGBA{255, 255, 0, 100})
		op := &ebiten.DrawImageOptions{GeoM: entityPos}
		op.GeoM.Translate(-c.entity.X, -c.entity.Y)
		op.GeoM.Translate(c.debugQueryRect.X, c.debugQueryRect.Y)
		screen.DrawImage(image, op)
		c.debugQueryRect = bump.Rect{}
	}
}

func (c *Comp) updateMovement(dt float64) {
	if c.Friction || math.Abs(c.Vx) > c.MaxX {
		var fric float64 = vars.GroundFriction
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

	p := bump.Vec2{X: c.entity.X + c.Vx*dt, Y: c.entity.Y + c.Vy*dt}
	goal, cols := c.space.Move(c, p, c.bodyFilter())
	c.entity.X, c.entity.Y = goal.X, goal.Y

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
		if _, ok := col.Other.(*Comp); ok && col.Type == bump.Cross && col.Overlaps {
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
	return func(item, other bump.Item) (bump.ColType, bool) {
		if obc, ok := other.(*Comp); ok && !obc.Solid {
			if utils.Contains(c.FilterOut, obc) {
				return 0, false
			}

			return bump.Cross, true
		}

		return bump.Slide, true
	}
}
