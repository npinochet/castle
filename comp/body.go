package comp

import (
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

func (bc *BodyComponent) IsActive() bool        { return bc.active }
func (bc *BodyComponent) SetActive(active bool) { bc.active = active }

type BodyComponent struct {
	active         bool
	Static, Ground bool
	Friction       bool
	space          *bump.Space
	entX, entY     *float64
	X, Y, W, H     float64
	Vx, Vy         float64
	MaxX, MaxY     float64
	image          *ebiten.Image
}

func (bc *BodyComponent) Init(entity *core.Entity) {
	bc.entX, bc.entY = &entity.X, &entity.Y
	if bc.MaxX == 0 {
		bc.MaxX = defaultMaxX
	}
	if bc.MaxY == 0 {
		bc.MaxY = defaultMaxY
	}
	bc.Friction = true
	bc.space = entity.World.Space
	bc.space.Set(bc, bump.Rect{X: entity.X + bc.X, Y: entity.Y + bc.X, W: bc.W, H: bc.H})
	bc.image = ebiten.NewImage(int(bc.W), int(bc.H))
	bc.image.Fill(color.RGBA{255, 0, 0, 100})
}

func (bc *BodyComponent) Update(dt float64) {
	if bc.Static {
		return
	}
	bc.updateMovement(dt)
}

func (bc *BodyComponent) DebugDraw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{GeoM: enitiyPos}
	op.GeoM.Translate(bc.X, bc.Y)
	screen.DrawImage(bc.image, op)
}

func (bc *BodyComponent) Destroy() {
	bc.space.Remove(bc)
}

func (bc *BodyComponent) updateMovement(dt float64) {
	if bc.Friction || math.Abs(bc.Vx) > bc.MaxX {
		var fric float64 = groundFriction
		if !bc.Ground {
			fric = airFriction
		}
		bc.Vx -= bc.Vx * fric * dt
		if math.Abs(bc.Vx) < frictionEpsilon {
			bc.Vx = 0
		}
	}
	bc.Vy += gravity * dt
	bc.Vy = math.Min(bc.MaxY, math.Max(-bc.MaxY, bc.Vy))

	p := bump.Vec2{X: *bc.entX + bc.X + bc.Vx*dt, Y: *bc.entY + bc.Y + bc.Vy*dt}
	goal, cols := bc.space.Move(bc, p, bodyFilter)
	*bc.entX, *bc.entY = goal.X-bc.X, goal.Y-bc.Y

	bc.Ground = false
	for _, col := range cols {
		if col.Type == bump.Slide {
			if col.Normal.Y < 0 {
				bc.Ground = true
				bc.Vy = 0
			}
			if col.Normal.X != 0 {
				bc.Vx = 0
			}
		}
		if col.Type == bump.Cross && col.Overlaps {
			bc.applyOverlapForce(col)
		}
	}
}

func (bc *BodyComponent) applyOverlapForce(col bump.Collision) {
	irect, orect := col.ItemRect, col.OtherRect
	overlap := (math.Min(irect.X+irect.W, orect.X+orect.W) - math.Max(irect.X, orect.X)) / math.Min(irect.W, orect.W)
	side := col.ItemRect.X + col.ItemRect.W/2 - (col.OtherRect.X + col.OtherRect.W/2)
	bc.Vx = math.Min(groundFriction, math.Max(-groundFriction, bc.Vx))
	bc.Vx += math.Copysign(overlap*collisionStiffness, side) * groundFriction
}

func bodyFilter(item, other bump.Item) (bump.ColType, bool) {
	if obc, ok := other.(*BodyComponent); ok && !obc.Static {
		return bump.Cross, true
	}
	if _, ok := other.(*Hitbox); ok {
		return bump.Cross, false
	}

	return bump.Slide, true
}
