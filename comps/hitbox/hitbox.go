package hitbox

import (
	"game/core"
	"game/libs/bump"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
)

const defaultITime = 1

type HitFunc func(*Comp, bump.Collision, float64)

type Hitbox struct {
	rect  bump.Rect
	comp  *Comp
	block bool
}

type Comp struct {
	Entity              *core.Entity
	space               *bump.Space
	boxes               []*Hitbox
	debugLastHitbox     *bump.Rect
	HurtFunc, BlockFunc HitFunc
	ITimer, ITime       float64
}

func (c *Comp) Init(entity *core.Entity) {
	c.Entity = entity
	c.ITime = defaultITime
	c.space = entity.World.Space
}

func (c *Comp) Update(dt float64) {
	c.ITimer -= dt
	for _, box := range c.boxes {
		p := bump.Vec2{X: c.Entity.X + box.rect.X, Y: c.Entity.Y + box.rect.Y}
		c.space.Move(box, p, bump.NilFilter)
	}
}

func (c *Comp) DebugDraw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {
	for _, box := range c.boxes {
		image := ebiten.NewImage(int(box.rect.W), int(box.rect.H))
		image.Fill(color.RGBA{0, 0, 255, 100})
		if box.block {
			image.Fill(color.RGBA{255, 0, 0, 100})
		}
		op := &ebiten.DrawImageOptions{GeoM: enitiyPos}
		op.GeoM.Translate(box.rect.X, box.rect.Y)
		screen.DrawImage(image, op)
	}

	if c.debugLastHitbox != nil {
		image := ebiten.NewImage(int(c.debugLastHitbox.W), int(c.debugLastHitbox.H))
		image.Fill(color.RGBA{255, 255, 0, 100})
		op := &ebiten.DrawImageOptions{GeoM: enitiyPos}
		op.GeoM.Translate(c.debugLastHitbox.X, c.debugLastHitbox.Y)
		screen.DrawImage(image, op)
		c.debugLastHitbox = nil
	}
}

func (c *Comp) PushHitbox(x, y, w, h float64, block bool) {
	rect := bump.Rect{X: x, Y: y, W: w, H: h}
	box := &Hitbox{rect, c, block}
	c.space.Set(box, rect)
	c.boxes = append(c.boxes, box)
}

func (c *Comp) PopHitbox() *Hitbox {
	size := len(c.boxes) - 1
	box := c.boxes[size]
	c.space.Remove(box)
	c.boxes = c.boxes[:size]

	return box
}

// TODO: Review how this function works, it's complicated.
// Remove I frames, and let one hit per hurtbox, for the duration of the attack somehow.
func (c *Comp) HitFromSpriteBox(rect *bump.Rect, damage float64) (blocked bool) {
	c.debugLastHitbox = rect
	cols := c.space.Query(bump.Rect{X: rect.X + c.Entity.X, Y: rect.Y + c.Entity.Y, W: rect.W, H: rect.H}, c.hitFilter())

	type hitInfo struct {
		hit bool
		col bump.Collision
	}

	doesHit := map[*Comp]hitInfo{}
	for _, col := range cols {
		if other, ok := col.Other.(*Hitbox); ok {
			if other.block {
				doesHit[other.comp] = hitInfo{false, col}
				blocked = true
			} else if _, set := doesHit[other.comp]; !set {
				doesHit[other.comp] = hitInfo{true, col}
			}
		} else if _, ok := col.Other.(*tiled.Object); ok {
			blocked = true
		}
	}

	for comp, info := range doesHit {
		if comp.ITimer > 0 {
			continue
		}
		if info.hit {
			comp.ITimer = comp.ITime
			if comp.HurtFunc != nil {
				comp.HurtFunc(c, info.col, damage)
			}
		} else if comp.BlockFunc != nil {
			comp.BlockFunc(c, info.col, damage)
		}
	}

	return blocked
}

func (c *Comp) QueryFront(dist, height float64, lookingRight bool) []*core.Entity { // TODO: This should be on body Comp.
	rect := bump.Rect{X: c.Entity.X - dist, Y: c.Entity.Y - height/2, W: dist, H: height}
	if lookingRight {
		rect.X += dist
	}
	cols := c.space.Query(rect, c.hitFilter())
	var ents []*core.Entity
	for _, c := range cols {
		if box, ok := c.Item.(*Hitbox); ok {
			ents = append(ents, box.comp.Entity)
		}
	}

	return ents
}

func (c *Comp) hitFilter() bump.SimpleFilter {
	return func(item bump.Item) bool {
		if box, ok := item.(*Hitbox); ok {
			return box.comp != c
		}

		return true
	}
}
