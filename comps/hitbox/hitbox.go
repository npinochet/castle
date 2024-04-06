package hitbox

import (
	"game/core"
	"game/libs/bump"
	"game/utils"
	"game/vars"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
)

const blockPriority = 1

var DebugDraw = false

type ContactType int

const (
	Hit ContactType = iota
	Block
	ParryBlock
)

type HitFunc func(core.Entity, *bump.Collision, float64, ContactType)

type Hitbox struct {
	rect              bump.Rect
	comp              *Comp
	contactType       ContactType
	updateContactType func() ContactType
}

type Comp struct {
	HitFunc         HitFunc
	entity          core.Entity
	space           *bump.Space
	hurtBoxes       []*Hitbox
	debugLastHitbox bump.Rect
}

func (c *Comp) Init(entity core.Entity) {
	c.entity = entity
	c.space = vars.World.Space
}

func (c *Comp) Update(_ float64) {
	ex, ey := c.entity.Position()
	for _, box := range c.hurtBoxes {
		if box.updateContactType != nil {
			if newContact := box.updateContactType(); newContact > 0 {
				box.contactType = newContact
			}
		}
		p := bump.Vec2{X: ex + box.rect.X, Y: ey + box.rect.Y}
		c.space.Move(box, p, bump.NilFilter)
	}
}

func (c *Comp) Draw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	if !DebugDraw {
		return
	}
	for _, box := range c.hurtBoxes {
		image := ebiten.NewImage(int(box.rect.W), int(box.rect.H))
		image.Fill(color.NRGBA{0, 0, 255, 75})
		if box.contactType != Hit {
			image.Fill(color.NRGBA{255, 0, 0, 75})
		}
		op := &ebiten.DrawImageOptions{GeoM: entityPos}
		op.GeoM.Translate(box.rect.X, box.rect.Y)
		screen.DrawImage(image, op)
	}

	if c.debugLastHitbox.W != 0 || c.debugLastHitbox.H != 0 {
		image := ebiten.NewImage(int(c.debugLastHitbox.W), int(c.debugLastHitbox.H))
		image.Fill(color.NRGBA{255, 255, 0, 75})
		op := &ebiten.DrawImageOptions{GeoM: entityPos}
		op.GeoM.Translate(c.debugLastHitbox.X, c.debugLastHitbox.Y)
		screen.DrawImage(image, op)
		c.debugLastHitbox = bump.Rect{}
	}
}

func (c *Comp) PushHitbox(rect bump.Rect, block ContactType, updateContactType func() ContactType) {
	if block != Hit {
		rect.Priority = blockPriority
	}
	box := &Hitbox{rect, c, block, updateContactType}
	c.space.Set(box, rect, "hitbox")
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

func (c *Comp) HitFromHitBox(rect bump.Rect, damage float64, filterOut []*Comp) (ContactType, []*Comp) {
	c.debugLastHitbox = rect
	ex, ey := c.entity.Position()
	rect.X += ex
	rect.Y += ey
	cols := c.space.Query(rect, c.hitFilter(), "hitbox", "map")

	type contactInfo struct {
		contactType ContactType
		col         *bump.Collision
	}

	var contacted []*Comp
	contact := Hit
	doesHit := map[*Comp]contactInfo{}
	for _, col := range cols {
		if other, ok := col.Other.(*Hitbox); ok { //nolint: nestif
			if utils.Contains(filterOut, other.comp) {
				continue
			}
			contacted = append(contacted, other.comp)
			if other.contactType > contact {
				contact = other.contactType
			}
			if doesHit[other.comp].col == nil {
				doesHit[other.comp] = contactInfo{Hit, col}
			}
			if other.contactType > doesHit[other.comp].contactType {
				doesHit[other.comp] = contactInfo{other.contactType, col}
			}
		} else if _, ok := col.Other.(*tiled.Object); ok && contact < Block && !c.space.Has(col.Other, "slope") {
			contact = Block
		}
	}

	for comp, info := range doesHit {
		if comp.HitFunc != nil {
			comp.HitFunc(c.entity, info.col, damage, info.contactType)
		}
	}

	return contact, append(filterOut, contacted...)
}

func (c *Comp) hitFilter() bump.SimpleFilter {
	return func(item bump.Item) bool {
		if box, ok := item.(*Hitbox); ok {
			return box.comp != c
		}

		// TODO: Review
		/*if obj, ok := item.(*tiled.Object); ok {
			itemRect := c.space.Rects[item]
			if obj.Class == core.LadderClass || itemRect.IsSlope() {
				return false
			}
		}*/
		// TODO: If a slope is hit, maybe it shouldn't return true
		return true
	}
}
