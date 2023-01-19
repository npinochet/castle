package anim

import (
	"fmt"
	"game/assets"
	"game/core"
	"game/libs/bump"
	"game/utils"
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

var DebugDraw = false

type FrameCallback func(frame int)

type Fsm struct {
	Initial     string
	Transitions map[string]string
	Entry       map[string]func(*Comp)
	Exit        map[string]func(*Comp)
}

type Comp struct {
	FilesName      string
	OX, OY         float64
	OXFlip, OYFlip float64
	FlipX, FlipY   bool
	w, h           float64
	State          string
	Image          *ebiten.Image
	Data           *aseprite.File
	Fsm            *Fsm
	callback       FrameCallback
	slices         map[string]map[int]bump.Rect
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

	if err := c.allocateHitboxSlices(); err != nil {
		log.Println(err)
	}
}

func (c *Comp) SetState(state string) {
	if c.State == state {
		return
	}
	c.State = state
	if err := c.Data.Play(state); err != nil {
		panic(err)
	}
	c.callback = nil
	if callback := c.Fsm.Entry[c.State]; callback != nil {
		callback(c)
	}
}

func (c *Comp) Update(dt float64) {
	c.Data.Update(float32(dt))
	if c.Data.AnimationFinished() {
		if callback := c.Fsm.Exit[c.State]; callback != nil {
			callback(c)
		}
		nextState, ok := c.Fsm.Transitions[c.State]
		if !ok {
			nextState = IdleTag
		}
		if nextState != "" {
			c.SetState(nextState)
		}
	}
	if c.callback != nil { // TODO: review and refactor whole per frame callback frame mechanic.
		c.callback(c.Data.CurrentFrame - c.Data.CurrentAnimation.From)
	}
}

func (c *Comp) Draw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{}
	var x, y, sx, sy, dx, dy float64 = c.OX, c.OY, 1, 1, 0, 0
	if c.FlipX {
		sx, dx = -1, math.Floor(c.w/2)+c.OXFlip
	}
	if c.FlipY {
		sy, dy = -1, math.Floor(c.h/2)+c.OYFlip
	}
	op.GeoM.Scale(sx, sy)
	op.GeoM.Translate(x, y)
	op.GeoM.Translate(dx, dy)
	op.GeoM.Concat(entityPos)
	sprite, _ := c.Image.SubImage(c.Data.FrameBoundaries().Rectangle()).(*ebiten.Image)
	screen.DrawImage(sprite, op)
	if DebugDraw {
		c.debugDraw(screen, entityPos)
	}
}

func (c *Comp) OnFrames(callback FrameCallback) {
	c.callback = callback
}

func (c *Comp) GetFrameHitbox(sliceName string) (bump.Rect, error) {
	keys := c.slices[sliceName]
	if keys == nil {
		return bump.Rect{}, fmt.Errorf("slice name %s not found", sliceName)
	}
	rect, ok := keys[c.Data.CurrentFrame]
	if !ok {
		return bump.Rect{}, fmt.Errorf("no slice in current frame %d", c.Data.CurrentFrame)
	}

	frame := c.Data.Frames.FrameAtIndex(c.Data.CurrentFrame)
	ssx, ssy := float64(frame.SpriteSourceSize.X), float64(frame.SpriteSourceSize.Y)
	sw, sh := float64(frame.SourceSize.Width), float64(frame.SourceSize.Height)

	if c.FlipX {
		rect.X += sw - rect.W - (rect.X+ssx)*2
	}
	if c.FlipY {
		rect.Y += sh - rect.W - (rect.Y+ssy)*2
	}
	rect.X += c.OX
	rect.Y += c.OY

	return rect, nil
}

func (c *Comp) allocateHitboxSlices() error {
	c.slices = map[string]map[int]bump.Rect{}

	for _, sliceName := range []string{HurtboxSliceName, HitboxSliceName, BlockSliceName} {
		slices := c.Data.Slice(sliceName)
		if slices == nil {
			return fmt.Errorf("slice name %s not found", sliceName)
		}

		c.slices[sliceName] = map[int]bump.Rect{}
		for _, key := range slices.Keys {
			sss := c.Data.Frames.FrameAtIndex(key.FrameNum).SpriteSourceSize

			bound := key.Bounds
			c.slices[sliceName][key.FrameNum] = bump.Rect{
				X: float64(bound.X) - float64(sss.X), Y: float64(bound.Y) - float64(sss.Y),
				W: float64(bound.Width), H: float64(bound.Height),
			}
		}
	}

	return nil
}

func (c *Comp) debugDraw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-5, -22)
	utils.DrawText(screen, fmt.Sprintf(`ANIM:%s`, c.State), assets.TinyFont, op)
}
