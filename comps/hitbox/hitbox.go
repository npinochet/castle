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
	comp  *Comp
	rect  bump.Rect
	block bool
	image *ebiten.Image
}

type Comp struct {
	EntX, EntY          *float64
	space               *bump.Space
	boxes               []*Hitbox
	debugLastHitbox     bump.Rect
	HurtFunc, BlockFunc HitFunc
	ITimer, ITime       float64
}

func (c *Comp) Init(entity *core.Entity) {
	c.EntX, c.EntY = &entity.X, &entity.Y
	c.ITime = defaultITime
	c.space = entity.World.Space
}

func (c *Comp) Update(dt float64) {
	c.ITimer -= dt
	for _, box := range c.boxes {
		p := bump.Vec2{X: *c.EntX + box.rect.X, Y: *c.EntY + box.rect.Y}
		c.space.Move(box, p, func(i, o bump.Item) (bump.ColType, bool) { return 0, false })
	}
}

func (c *Comp) DebugDraw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {
	for _, box := range c.boxes {
		op := &ebiten.DrawImageOptions{GeoM: enitiyPos}
		op.GeoM.Translate(box.rect.X, box.rect.Y)
		screen.DrawImage(box.image, op)
	}

	if c.debugLastHitbox.W > 0 && c.debugLastHitbox.H > 0 {
		image := ebiten.NewImage(int(c.debugLastHitbox.W), int(c.debugLastHitbox.H))
		image.Fill(color.RGBA{255, 255, 0, 100})
		op := &ebiten.DrawImageOptions{GeoM: enitiyPos}
		op.GeoM.Translate(c.debugLastHitbox.X, c.debugLastHitbox.Y)
		screen.DrawImage(image, op)
		c.debugLastHitbox.W = 0
	}
}

func (c *Comp) Destroy() {
	for len(c.boxes) > 0 {
		c.PopHitbox()
	}
}

func (c *Comp) PushHitbox(x, y, w, h float64, block bool) {
	image := ebiten.NewImage(int(w), int(h))
	rect := bump.Rect{X: x, Y: y, W: w, H: h}
	image.Fill(color.RGBA{0, 0, 255, 100})
	if block {
		image.Fill(color.RGBA{255, 0, 0, 100})
	}
	box := &Hitbox{c, rect, block, image}
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

func (c *Comp) Hit(x, y, w, h, damage float64) (blocked bool) {
	c.debugLastHitbox = bump.Rect{X: x, Y: y, W: w, H: h}
	cols := c.space.Query(bump.Rect{X: x + *c.EntX, Y: y + *c.EntY, W: w, H: h}, c.hitFilter())

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
		} else {
			if comp.BlockFunc != nil {
				comp.BlockFunc(c, info.col, damage)
			}
		}
	}

	return blocked
}

func (c *Comp) hitFilter() bump.SimpleFilter {
	return func(item bump.Item) bool {
		if box, ok := item.(*Hitbox); ok {
			return box.comp != c
		}

		return true
	}
}
