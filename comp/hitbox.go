package comp

import (
	"game/core"
	"game/libs/bump"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
)

var defaultITime float64 = 0.2

func (hc *HitboxComponent) IsActive() bool        { return hc.active }
func (hc *HitboxComponent) SetActive(active bool) { hc.active = active }

type HitFunc func(*HitboxComponent, bump.Colision, float64)

type Hitbox struct {
	comp  *HitboxComponent
	rect  bump.Rect
	block bool
	image *ebiten.Image
}

type HitboxComponent struct {
	active        bool
	EntX, EntY    *float64
	space         *bump.Space
	boxes         []*Hitbox
	lastHitbox    bump.Rect
	HurtFunc      HitFunc
	BlockFunc     HitFunc
	ITimer, ITime float64
}

func (hc *HitboxComponent) Init(entity *core.Entity) {
	hc.EntX, hc.EntY = &entity.X, &entity.Y
	hc.ITime = defaultITime
	hc.space = entity.World.Space
}

func (hc *HitboxComponent) Update(dt float64) {
	hc.ITimer -= dt
	for _, box := range hc.boxes {
		p := bump.Vec2{X: *hc.EntX + box.rect.X, Y: *hc.EntY + box.rect.Y}
		hc.space.Move(box, p, func(i, o bump.Item) (bump.ColType, bool) { return 0, false })
	}
}

func (hc *HitboxComponent) DebugDraw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {
	for _, box := range hc.boxes {
		op := &ebiten.DrawImageOptions{GeoM: enitiyPos}
		op.GeoM.Translate(box.rect.X, box.rect.Y)
		screen.DrawImage(box.image, op)
	}

	if hc.lastHitbox.W > 0 && hc.lastHitbox.H > 0 {
		image := ebiten.NewImage(int(hc.lastHitbox.W), int(hc.lastHitbox.H))
		image.Fill(color.RGBA{255, 255, 0, 100})
		op := &ebiten.DrawImageOptions{GeoM: enitiyPos}
		op.GeoM.Translate(hc.lastHitbox.X, hc.lastHitbox.Y)
		screen.DrawImage(image, op)
		hc.lastHitbox.W = 0
	}
}

func (hc *HitboxComponent) Destroy() {
	for len(hc.boxes) > 0 {
		hc.PopHitbox()
	}
}

func (hc *HitboxComponent) PushHitbox(x, y, w, h float64, block bool) {
	image := ebiten.NewImage(int(w), int(h))
	rect := bump.Rect{X: x, Y: y, W: w, H: h}
	image.Fill(color.RGBA{0, 0, 255, 100})
	if block {
		image.Fill(color.RGBA{255, 0, 0, 100})
	}
	box := &Hitbox{hc, rect, block, image}
	hc.space.Set(box, rect)
	hc.boxes = append(hc.boxes, box)
}

func (hc *HitboxComponent) PopHitbox() *Hitbox {
	size := len(hc.boxes) - 1
	box := hc.boxes[size]
	hc.space.Remove(box)
	hc.boxes = hc.boxes[:size]
	return box
}

func (hc *HitboxComponent) Hit(x, y, w, h, damage float64) (blocked bool) {
	hc.lastHitbox = bump.Rect{X: x, Y: y, W: w, H: h}
	cols := hc.space.Query(bump.Rect{X: x + *hc.EntX, Y: y + *hc.EntY, W: w, H: h}, hc.hitFilter())

	type hitInfo struct {
		hit bool
		col bump.Colision
	}

	doesHit := map[*HitboxComponent]hitInfo{}
	for _, col := range cols {
		other, ok := col.Other.(*Hitbox)
		if ok {
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
		comp.ITimer = comp.ITime
		if info.hit && comp.HurtFunc != nil {
			comp.HurtFunc(hc, info.col, damage)
		} else if !info.hit && comp.BlockFunc != nil {
			comp.BlockFunc(hc, info.col, damage)
		}
	}
	return
}

func (hc *HitboxComponent) hitFilter() bump.SimpleFilter {
	return func(item bump.Item) bool {
		if box, ok := item.(*Hitbox); ok {
			return box.comp != hc
		}
		return true
	}
}
