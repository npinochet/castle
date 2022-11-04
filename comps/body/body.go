package body

import (
	"game/comps/hitbox"
	"game/core"
	"game/libs/bump"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	gravity                     = 300
	defaultMaxX, defaultMaxY    = 60, 200
	groundFriction, airFriction = 12, 4
	collisionStiffness          = 1
	frictionEpsilon             = 0.05
)

type Comp struct {
	Static, Ground bool
	Friction       bool
	space          *bump.Space
	entX, entY     *float64
	X, Y, W, H     float64
	Vx, Vy         float64
	MaxX, MaxY     float64
	image          *ebiten.Image
}

func (c *Comp) Init(entity *core.Entity) {
	c.entX, c.entY = &entity.X, &entity.Y
	if c.MaxX == 0 {
		c.MaxX = defaultMaxX
	}
	if c.MaxY == 0 {
		c.MaxY = defaultMaxY
	}
	c.Friction = true
	c.space = entity.World.Space
	c.space.Set(c, bump.Rect{X: entity.X + c.X, Y: entity.Y + c.X, W: c.W, H: c.H})
	c.image = ebiten.NewImage(int(c.W), int(c.H))
	c.image.Fill(color.RGBA{255, 0, 0, 100})
}

func (c *Comp) Update(dt float64) {
	if c.Static {
		return
	}
	c.updateMovement(dt)
}

func (c *Comp) DebugDraw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{GeoM: enitiyPos}
	op.GeoM.Translate(c.X, c.Y)
	screen.DrawImage(c.image, op)
}

func (c *Comp) updateMovement(dt float64) {
	if c.Friction || math.Abs(c.Vx) > c.MaxX {
		var fric float64 = groundFriction
		if !c.Ground {
			fric = airFriction
		}
		c.Vx -= c.Vx * fric * dt
		if math.Abs(c.Vx) < frictionEpsilon {
			c.Vx = 0
		}
	}
	c.Vy += gravity * dt
	c.Vy = math.Min(c.MaxY, math.Max(-c.MaxY, c.Vy))

	p := bump.Vec2{X: *c.entX + c.X + c.Vx*dt, Y: *c.entY + c.Y + c.Vy*dt}
	goal, cols := c.space.Move(c, p, bodyFilter)
	*c.entX, *c.entY = goal.X-c.X, goal.Y-c.Y

	c.Ground = false
	for _, col := range cols {
		if col.Type == bump.Slide {
			if col.Normal.Y < 0 {
				c.Ground = true
				c.Vy = 0
			}
			if col.Normal.X != 0 {
				c.Vx = 0
			}
		}
		if col.Type == bump.Cross && col.Overlaps {
			c.applyOverlapForce(col)
		}
	}
}

func (c *Comp) applyOverlapForce(col *bump.Collision) {
	irect, orect := col.ItemRect, col.OtherRect
	overlap := (math.Min(irect.X+irect.W, orect.X+orect.W) - math.Max(irect.X, orect.X)) / math.Min(irect.W, orect.W)
	side := col.ItemRect.X + col.ItemRect.W/2 - (col.OtherRect.X + col.OtherRect.W/2)
	c.Vx = math.Min(groundFriction, math.Max(-groundFriction, c.Vx))
	c.Vx += math.Copysign(overlap*collisionStiffness, side) * groundFriction
}

func bodyFilter(item, other bump.Item) (bump.ColType, bool) {
	if obc, ok := other.(*Comp); ok && !obc.Static {
		return bump.Cross, true
	}
	if _, ok := other.(*hitbox.Hitbox); ok {
		return bump.Cross, false
	}

	return bump.Slide, true
}
