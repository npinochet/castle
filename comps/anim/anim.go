package anim

import (
	"fmt"
	"game/core"
	"game/libs/bump"
	"log"
	"math"

	"github.com/damienfamed75/aseprite"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	HurtboxSliceName = "hurtbox"
	HitboxSliceName  = "hitbox"
	BlockSliceName   = "blockbox"
)

type FrameCallback func(frame int)

type Fsm struct {
	Initial        string
	Transitions    map[string]string
	ExitCallbacks  map[string]func(*Comp)
	EnterCallbacks map[string]func(*Comp)
}

type Comp struct {
	FilesName    string
	X, Y         float64
	FlipX, FlipY bool
	w, h         float64
	State        string
	Image        *ebiten.Image
	Data         *aseprite.File
	Fsm          *Fsm
	callback     FrameCallback
	slices       [3]*aseprite.Slice
}

func (c *Comp) Init(entity *core.Entity) {
	var err error
	if c.Image, _, err = ebitenutil.NewImageFromFile(fmt.Sprintf("%s.png", c.FilesName)); err != nil {
		log.Fatal(err)
	}
	if c.Data, err = aseprite.Open(fmt.Sprintf("%s.json", c.FilesName)); err != nil {
		log.Fatal(err)
	}

	c.SetState(c.Data.Meta.Animations[0].Name)
	rect := c.Data.Frames.FrameAtIndex(c.Data.CurrentFrame).SpriteSourceSize
	c.w, c.h = float64(rect.Width), float64(rect.Height)

	for i, sliceName := range [3]string{HurtboxSliceName, HitboxSliceName, BlockSliceName} {
		c.slices[i] = c.Data.Slice(sliceName)
	}
}

func (c *Comp) SetState(state string) {
	if c.State == state {
		return
	}
	c.State = state
	_ = c.Data.Play(state)
	c.callback = nil
	if callback := c.Fsm.EnterCallbacks[c.State]; callback != nil {
		callback(c)
	}
}

func (c *Comp) Update(dt float64) {
	c.Data.Update(float32(dt))
	if c.Data.AnimationFinished() {
		if callback := c.Fsm.ExitCallbacks[c.State]; callback != nil {
			callback(c)
		}
		if nextState := c.Fsm.Transitions[c.State]; nextState != "" {
			c.SetState(nextState)
		}
	}
	if c.callback != nil { // TODO: review and refactor whole per frame callback frame mechanic.
		c.callback(c.Data.CurrentFrame - c.Data.CurrentAnimation.From)
	}
}

func (c *Comp) Draw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{}
	var x, y, sx, sy, dx, dy float64 = c.X, c.Y, 1, 1, 0, 0
	if c.FlipX {
		x, sx, dx = -x, -1, math.Floor(c.w/2)
	}
	if c.FlipY {
		y, sy, dy = -y, -1, math.Floor(c.h/2)
	}
	op.GeoM.Scale(sx, sy)
	op.GeoM.Translate(x, y)
	op.GeoM.Translate(dx, dy)
	op.GeoM.Concat(enitiyPos)
	sprite, _ := c.Image.SubImage(c.Data.FrameBoundaries().Rectangle()).(*ebiten.Image)
	screen.DrawImage(sprite, op)
}

func (c *Comp) OnFrames(callback FrameCallback) {
	c.callback = callback
}

func (c *Comp) GetFrameHitboxes() (hurtbox, hitbox, blockbox *bump.Rect) {
	hurtbox = c.findCurrenctSlice(c.slices[0])
	hitbox = c.findCurrenctSlice(c.slices[1])
	blockbox = c.findCurrenctSlice(c.slices[2])

	return hurtbox, hitbox, blockbox
}

func (c *Comp) findCurrenctSlice(slice *aseprite.Slice) *bump.Rect {
	if slice == nil {
		return nil
	}
	currentFrame := c.Data.CurrentFrame
	frame := c.Data.Frames.FrameAtIndex(currentFrame)
	ssx, ssy := float64(frame.SpriteSourceSize.X), float64(frame.SpriteSourceSize.Y)
	sw, sh := float64(frame.SourceSize.Width), float64(frame.SourceSize.Height)

	// TODO: Find a way to get the truly frame slices, aseprite goes nuts on these.
	for _, key := range slice.Keys {
		if key.FrameNum != currentFrame {
			continue
		}
		bound := key.Bounds
		rect := &bump.Rect{
			X: float64(bound.X) - ssx + c.X, Y: float64(bound.Y) - ssy + c.Y,
			W: float64(bound.Width), H: float64(bound.Height),
		}

		if c.FlipX {
			rect.X += sw - rect.W - float64(bound.X)*2
		}
		if c.FlipY {
			rect.Y += sh - rect.W - float64(bound.Y)*2
		}

		return rect
	}

	return nil
}
