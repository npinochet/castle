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
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

var DebugDraw = false

type Comp struct {
	Solid, Unmovable, Friction bool
	Ground, InsidePassThrough  bool
	droppingThrough            bool
	Vx, Vy                     float64
	MaxX, MaxY                 float64
	Weight                     float64
	Tags, QueryTags            []bump.Tag
	FilterOut                  []core.Entity
	entity                     core.Entity
	space                      *bump.Space
	prevVx                     float64
	coyoteTime                 float64
}

func (c *Comp) Init(entity core.Entity) {
	c.entity = entity
	if c.MaxX == 0 {
		c.MaxX = vars.DefaultMaxX
	}
	if c.MaxY == 0 {
		c.MaxY = vars.DefaultMaxY
	}
	if c.Weight == 0 {
		c.Weight = 1
	}
	if c.Tags == nil {
		c.Tags = []bump.Tag{"body"}
	}
	if c.QueryTags == nil {
		c.QueryTags = []bump.Tag{"body", "map", "solid"}
	}
	c.Friction = true
	c.space = vars.World.Space
	c.space.Set(entity, bump.NewRect(entity.Rect()), c.Tags...)
}

func (c *Comp) Remove() { c.space.Remove(c.entity) }

func (c *Comp) Update(dt float64) {
	if c.Solid {
		return
	}
	noForceApplied := !c.Ground || c.prevVx == c.Vx
	c.updateMovement(dt, noForceApplied)
	c.prevVx = c.Vx
	if c.coyoteTime -= dt; c.coyoteTime > 0 {
		c.Ground = true
	}
}

func (c *Comp) Draw(pipeline *core.Pipeline, entityPos ebiten.GeoM) {
	if !DebugDraw {
		return
	}
	friction := (c.Friction && !c.Ground || c.prevVx == c.Vx) || math.Abs(c.Vx) > c.MaxX
	_, _, ew, eh := c.entity.Rect()
	image := ebiten.NewImage(int(ew), int(eh))
	image.Fill(color.NRGBA{255, 0, 0, 75})
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	pipeline.Add(vars.PipelineScreenTag, vars.PipelineUILayer, func(screen *ebiten.Image) {
		screen.DrawImage(image, op)
		op.GeoM.Translate(-5, -22)
		utils.DrawText(screen, fmt.Sprintf(`FRIC:%v`, friction), assets.NanoFont, op)
		op.GeoM.Translate(0, 6)
		utils.DrawText(screen, fmt.Sprintf(`MAX:%v`, c.MaxX), assets.NanoFont, op)
	})
}

func (c *Comp) QueryFloor(tags ...bump.Tag) bool {
	x, y, w, h := c.entity.Rect()

	return len(c.space.Query(bump.NewRect(x, y+h, w, 1), nil, tags...)) > 0
}

func (c *Comp) DropThrough() bool {
	onPassThrough := c.QueryFloor("passthrough")
	if onPassThrough {
		c.droppingThrough = true
	}

	return onPassThrough
}

func (c *Comp) updateMovement(dt float64, noForceApplied bool) {
	if (c.Friction && noForceApplied) || math.Abs(c.Vx) > c.MaxX {
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
	t := bump.Vec2{X: ex + c.Vx*dt, Y: ey + c.Vy*dt}
	goal, cols := c.space.Move(c.entity, t, c.bodyFilter(), c.QueryTags...)
	c.entity.SetPosition(goal.X, goal.Y)

	c.Ground = false
	c.InsidePassThrough = false
	for _, col := range cols {
		if col.Type == bump.Slide {
			if col.Normal.X != 0 {
				c.Vx = 0
			}
			if col.Normal.Y < 0 || (col.Normal.Y > 0 && c.Vy < 0) {
				c.Vy = 0
			}
			c.Ground = c.Ground || col.Normal.Y < 0
		}
		if c.Unmovable {
			continue
		}
		if _, ok := col.Other.(core.Entity); ok && col.Type == bump.Cross && col.Overlaps {
			c.applyOverlapForce(col)
		}
		c.InsidePassThrough = c.InsidePassThrough || (c.space.Has(col.Other, "passthrough") && col.Overlaps)
	}
	if c.Ground {
		c.coyoteTime = vars.CoyoteTimeSeconds
		if c.QueryFloor("slope") {
			c.Vy = c.MaxY / 4
		}
	}
	if !c.InsidePassThrough {
		c.droppingThrough = false
	}
}

func (c *Comp) applyOverlapForce(col *bump.Collision) {
	irect, orect := col.ItemRect, col.OtherRect
	overlap := (math.Min(irect.X+irect.W, orect.X+orect.W) - math.Max(irect.X, orect.X)) / math.Min(irect.W, orect.W)
	side := col.ItemRect.X + col.ItemRect.W/2 - (col.OtherRect.X + col.OtherRect.W/2)
	if side > 0 && c.Vx < 0 || side < 0 && c.Vx > 0 {
		c.Vx = math.Min(vars.GroundFriction, math.Max(-vars.GroundFriction, c.Vx))
	}
	c.Vx += math.Copysign(overlap*vars.CollisionStiffness, side) * vars.GroundFriction
}

func (c *Comp) bodyFilter() func(bump.Item, bump.Item) (bump.ColType, bool) {
	return func(item, other bump.Item) (bump.ColType, bool) {
		if entity, ok := other.(core.Entity); ok {
			if slices.Contains(c.FilterOut, entity) {
				return 0, false
			}
			if c.space.Has(entity, "solid") {
				return bump.Slide, true
			}

			return bump.Cross, true
		}
		if c.space.Has(other, "passthrough") {
			itemRect, otherRect := c.space.Rect(item), c.space.Rect(other)
			if !c.droppingThrough && itemRect.Y+itemRect.H <= otherRect.Y {
				return bump.Slide, true
			}

			return bump.Cross, true
		}

		return bump.Slide, true
	}
}
