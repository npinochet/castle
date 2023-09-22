package actor

import (
	"game/core"
	"game/libs/bump"
	"game/utils"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
)

const hBlockPriority = 1

var HDebugDraw = false

type HHitFunc func(actor *Actor, col *bump.Collision, damage float64)

type HBox struct {
	rect  bump.Rect
	actor *Actor
	block bool
}

type Hitbox struct {
	HurtFunc, BlockFunc HHitFunc
	space               *bump.Space
	hurtBoxes           []*HBox
	debugLastHitbox     bump.Rect
}

func (c *Hitbox) Init(a *Actor) {
	c.space = a.World.Space
}

func (c *Hitbox) Update(a *Actor) {
	for _, box := range c.hurtBoxes {
		c.space.Move(box, bump.Vec2{X: a.X + box.rect.X, Y: a.Y + box.rect.Y}, bump.NilFilter)
	}
}

func (c *Hitbox) Draw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	if !HDebugDraw {
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

func (c *Hitbox) PushHitbox(a *Actor, relRect bump.Rect, block bool) {
	if block && relRect.Priority == 0 {
		relRect.Priority = hBlockPriority
	}
	box := &HBox{relRect, a, block}
	c.space.Set(box, relRect)
	c.hurtBoxes = append(c.hurtBoxes, box)
}

func (c *Hitbox) PopHitbox() *HBox {
	last := len(c.hurtBoxes) - 1
	if last < 0 {
		return nil
	}
	box := c.hurtBoxes[last]
	c.space.Remove(box)
	c.hurtBoxes = c.hurtBoxes[:last]

	return box
}

func (c *Hitbox) Hit(a *Actor, relRect bump.Rect, damage float64, contactedToIgnore []*Actor) (bool, []*Actor) {
	c.debugLastHitbox = relRect
	relRect.X += a.X
	relRect.Y += a.Y

	type hitInfo struct {
		hit bool
		col *bump.Collision
	}

	blocked := false
	doesHit := map[*Actor]hitInfo{}
	for _, col := range c.space.Query(relRect, c.hitFilter(a)) {
		if other, ok := col.Other.(*HBox); ok {
			if utils.Contains(contactedToIgnore, other.actor) {
				continue
			}
			contactedToIgnore = append(contactedToIgnore, other.actor)
			if other.block {
				blocked = true
				doesHit[other.actor] = hitInfo{false, col}
			} else if _, set := doesHit[other.actor]; !set {
				doesHit[other.actor] = hitInfo{true, col}
			}
		} else if _, ok := col.Other.(*tiled.Object); ok {
			blocked = true // TODO: Should not stagger when hitting slope
		}
	}

	for actor, info := range doesHit {
		if info.hit {
			if actor.HurtFunc != nil {
				actor.HurtFunc(actor, info.col, damage)
			}
		} else if actor.BlockFunc != nil {
			actor.BlockFunc(actor, info.col, damage)
		}
	}

	return blocked, contactedToIgnore
}

func (c *Hitbox) hitFilter(a *Actor) bump.SimpleFilter {
	return func(item bump.Item) bool {
		if box, ok := item.(*HBox); ok {
			return box.actor != a
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
