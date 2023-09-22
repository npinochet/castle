package actor

import (
	"fmt"
	"game/assets"
	"game/core"
	"game/libs/bump"
	"game/utils"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
)

const (
	BGravity                      = 300
	bDefaultMaxX, bDefaultMaxY    = 20, 200
	bGroundFriction, bAirFriction = 12.0, 2.0 // TODO: Tune this variables. They might be too high.
	bCollisionStiffness           = 1
	bFrictionEpsilon              = 0.05
)

type Team int // TODO: Review this thing about teams

const (
	NoneTeam Team = iota
	PlayerTeam
	EnemyTeam
)

var BDebugDraw = false

type Body struct {
	Solid, Unmovable     bool
	Ground, Friction     bool
	OnLadder, ClipLadder bool
	Team                 Team
	MaxXMultiplier       float64
	Vx, Vy               float64
	MaxX, MaxY           float64
	Weight               float64
	FilterOut            []*Actor
	space                *bump.Space
	prevVx               float64
	debugQueryRect       bump.Rect
}

func (c *Body) Init(a *Actor) {
	if c.MaxX == 0 {
		c.MaxX = bDefaultMaxX
	}
	if c.MaxY == 0 {
		c.MaxY = bDefaultMaxY
	}
	if c.Team == NoneTeam {
		c.Team = EnemyTeam
	}
	c.Friction = true
	c.space = a.World.Space

	c.space.Set(a, bump.NewRect(a.Rect()))
}

func (c *Body) Update(a *Actor, dt float64) {
	if c.Solid {
		return
	}
	c.Friction = !c.Ground || math.Abs(c.prevVx-c.Vx) >= 0 // TODO: Review this
	c.updateMovement(a, dt)
	c.prevVx = c.Vx
}

func (c *Body) Query(rect bump.Rect, filter func(item bump.Item) bool) []*bump.Collision {
	c.debugQueryRect = rect

	return c.space.Query(rect, filter)
}

func (c *Body) QueryActors(a *Actor, rect bump.Rect, ignoreOwnTeam bool) []*Actor {
	actorFilter := func(item bump.Item) bool {
		if oa, ok := item.(*Actor); ok {
			return oa != a && (!ignoreOwnTeam || oa.Body.Team != c.Team)
		}

		return false
	}

	cols := c.Query(rect, actorFilter)
	var actors []*Actor
	for _, c := range cols {
		if actor, ok := c.Other.(*Actor); ok {
			actors = append(actors, actor)
		}
	}

	return actors
}

func (c *Body) QueryFront(a *Actor, dist, height float64, lookingRight, ignoreOwnTeam bool) []*Actor {
	rect := bump.Rect{X: -dist, Y: -height / 2, W: dist, H: height}
	if lookingRight {
		rect.X += dist
	}
	rect.X += a.X
	rect.Y += a.Y

	return c.QueryActors(a, rect, ignoreOwnTeam)
}

func (c *Body) Draw(a *Actor, screen *ebiten.Image, entityPos ebiten.GeoM) {
	if !BDebugDraw {
		return
	}
	image := ebiten.NewImage(int(a.W), int(a.H))
	image.Fill(color.RGBA{255, 0, 0, 100})
	screen.DrawImage(image, &ebiten.DrawImageOptions{GeoM: entityPos})

	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-5, -22)
	utils.DrawText(screen, fmt.Sprintf(`FRIC:%v`, c.Friction), assets.TinyFont, op)

	if c.debugQueryRect.W != 0 || c.debugQueryRect.H != 0 {
		image := ebiten.NewImage(int(c.debugQueryRect.W), int(c.debugQueryRect.H))
		image.Fill(color.RGBA{255, 255, 0, 100})
		op := &ebiten.DrawImageOptions{GeoM: entityPos}
		op.GeoM.Translate(-a.X, -a.Y)
		op.GeoM.Translate(c.debugQueryRect.X, c.debugQueryRect.Y)
		screen.DrawImage(image, op)
		c.debugQueryRect = bump.Rect{}
	}
}

func (c *Body) updateMovement(a *Actor, dt float64) {
	if c.Friction || math.Abs(c.Vx) > c.MaxX*(c.MaxXMultiplier+1) {
		fric := bGroundFriction
		if !c.Ground {
			fric = bAirFriction
		}
		c.Vx -= c.Vx * fric * dt
		if math.Abs(c.Vx) < bFrictionEpsilon {
			c.Vx = 0
		}
	}
	c.Vy += BGravity * (c.Weight + 1) * dt
	c.Vy = math.Min(c.MaxY, math.Max(-c.MaxY, c.Vy))

	p := bump.Vec2{X: a.X + c.Vx*dt, Y: a.Y + c.Vy*dt}
	goal, cols := c.space.Move(c, p, c.bodyFilter())
	a.X, a.Y = goal.X, goal.Y
	if c.Unmovable {
		return
	}

	c.Ground, c.OnLadder = false, false
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
		if _, ok := col.Other.(*Actor); ok && col.Type == bump.Cross && col.Overlaps {
			c.applyOverlapForce(col)
		}
		if obj, ok := col.Other.(*tiled.Object); ok && obj.Class == core.LadderClass {
			c.OnLadder = c.OnLadder || col.Overlaps
		}
	}
}

func (c *Body) applyOverlapForce(col *bump.Collision) {
	irect, orect := col.ItemRect, col.OtherRect
	overlap := (math.Min(irect.X+irect.W, orect.X+orect.W) - math.Max(irect.X, orect.X)) / math.Min(irect.W, orect.W)
	side := col.ItemRect.X + col.ItemRect.W/2 - (col.OtherRect.X + col.OtherRect.W/2)
	c.Vx = math.Min(bGroundFriction, math.Max(-bGroundFriction, c.Vx))
	c.Vx += math.Copysign(overlap*bCollisionStiffness, side) * bGroundFriction
}

func (c *Body) bodyFilter() func(bump.Item, bump.Item) (bump.ColType, bool) {
	return func(item, other bump.Item) (bump.ColType, bool) {
		if obc, ok := other.(*Actor); ok && !obc.Solid {
			if utils.Contains(c.FilterOut, obc) {
				return 0, false
			}

			return bump.Cross, true
		}
		if _, ok := other.(*HBox); ok {
			return 0, false
		}
		if obj, ok := other.(*tiled.Object); ok && obj.Class == core.LadderClass {
			itemRect, otherRect := c.space.Rects[item], c.space.Rects[other]
			if !c.ClipLadder && itemRect.Y+itemRect.H <= otherRect.Y {
				return bump.Slide, true
			}

			return bump.Cross, true
		}

		return bump.Slide, true
	}
}
