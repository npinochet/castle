package hitbox

import (
	"game/core"
	"game/libs/bump"
	"game/utils"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
)

const blockPriority = 1

var DebugDraw = false

type BlockType int

const (
	NoBlock BlockType = iota
	Block
	ParryBlock
)

type HitFunc func(*core.Entity, *bump.Collision, float64)

type Hitbox struct {
	rect  bump.Rect
	comp  *Comp
	block bool
}

type Comp struct {
	*core.Entity
	HurtFunc, BlockFunc HitFunc
	ParryBlocking       func() bool
	space               *bump.Space
	hurtBoxes           []*Hitbox
	debugLastHitbox     bump.Rect
}

func (c *Comp) Init(entity *core.Entity) {
	c.Entity = entity
	c.space = entity.World.Space
}

func (c *Comp) Update(_ float64) {
	for _, box := range c.hurtBoxes {
		p := bump.Vec2{X: c.Entity.X + box.rect.X, Y: c.Entity.Y + box.rect.Y}
		c.space.Move(box, p, bump.NilFilter)
	}
}

func (c *Comp) Draw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	if !DebugDraw {
		return
	}
	for _, box := range c.hurtBoxes {
		image := ebiten.NewImage(int(box.rect.W), int(box.rect.H))
		image.Fill(color.RGBA{0, 0, 255, 100})
		if box.block {
			image.Fill(color.RGBA{255, 0, 0, 100})
		}
		op := &ebiten.DrawImageOptions{GeoM: entityPos}
		op.GeoM.Translate(box.rect.X, box.rect.Y)
		screen.DrawImage(image, op)
	}

	if c.debugLastHitbox.W != 0 || c.debugLastHitbox.H != 0 {
		image := ebiten.NewImage(int(c.debugLastHitbox.W), int(c.debugLastHitbox.H))
		image.Fill(color.RGBA{255, 255, 0, 100})
		op := &ebiten.DrawImageOptions{GeoM: entityPos}
		op.GeoM.Translate(c.debugLastHitbox.X, c.debugLastHitbox.Y)
		screen.DrawImage(image, op)
		c.debugLastHitbox = bump.Rect{}
	}
}

func (c *Comp) PushHitbox(rect bump.Rect, block bool) {
	if block {
		rect.Priority = blockPriority
	}
	box := &Hitbox{rect, c, block}
	c.space.Set(box, rect)
	c.hurtBoxes = append(c.hurtBoxes, box)
}

func (c *Comp) PopHitbox() *Hitbox {
	size := len(c.hurtBoxes) - 1
	if size < 0 {
		return nil
	}
	box := c.hurtBoxes[size]
	c.space.Remove(box)
	c.hurtBoxes = c.hurtBoxes[:size]

	return box
}

func (c *Comp) HitFromHitBox(rect bump.Rect, damage float64, filterOut []*Comp) (BlockType, []*Comp) {
	c.debugLastHitbox = rect
	rect.X += c.Entity.X
	rect.Y += c.Entity.Y
	cols := c.space.Query(rect, c.hitFilter())

	type hitInfo struct {
		hit bool
		col *bump.Collision
	}

	var contacted []*Comp
	blocked := NoBlock
	doesHit := map[*Comp]hitInfo{}
	for _, col := range cols {
		if other, ok := col.Other.(*Hitbox); ok {
			if utils.Contains(filterOut, other.comp) {
				continue
			}
			contacted = append(contacted, other.comp)
			if other.block {
				doesHit[other.comp] = hitInfo{false, col}
				if blocked == NoBlock {
					blocked = Block
				}
				if other.comp.ParryBlocking() {
					blocked = ParryBlock
				}
			} else if _, set := doesHit[other.comp]; !set {
				doesHit[other.comp] = hitInfo{true, col}
			}
		} else if _, ok := col.Other.(*tiled.Object); ok && blocked == NoBlock {
			blocked = Block // TODO: Should not stagger when hitting slope
		}
	}

	for comp, info := range doesHit {
		if info.hit {
			if comp.HurtFunc != nil {
				comp.HurtFunc(c.Entity, info.col, damage)
			}
		} else if comp.BlockFunc != nil {
			comp.BlockFunc(c.Entity, info.col, damage)
		}
	}

	return blocked, append(filterOut, contacted...)
}

func (c *Comp) hitFilter() bump.SimpleFilter {
	return func(item bump.Item) bool {
		if box, ok := item.(*Hitbox); ok {
			return box.comp != c
		}

		if obj, ok := item.(*tiled.Object); ok {
			itemRect := c.space.Rects[item]
			if obj.Class == core.LadderClass || itemRect.IsSlope() {
				return false
			}
		}
		// TODO: If a slope is hit, maybe it shouldn't return true
		return true
	}
}