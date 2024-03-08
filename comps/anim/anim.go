package anim

import (
	"fmt"
	"game/assets"
	"game/core"
	"game/libs/bump"
	"game/utils"
	"game/vars"
	"log"
	"math"

	"github.com/damienfamed75/aseprite"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var DebugDraw = false

type SliceCallback func(slice bump.Rect, segmented bool)

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
	Fsm            *Fsm

	State          string
	Image          *ebiten.Image
	Data           *aseprite.File
	w, h           float64
	sliceCallback  func()
	frameCallbacks map[int]func()
	exitFunc       func()
	slices         map[string]map[int]bump.Rect
}

func (c *Comp) Init(_ core.Entity) {
	var err error
	if c.Image, _, err = ebitenutil.NewImageFromFile(c.FilesName + ".png"); err != nil {
		log.Fatal(err)
	}
	if c.Data, err = aseprite.Open(c.FilesName + ".json"); err != nil {
		log.Fatal(err)
	}

	c.SetState(c.Data.Meta.Animations[0].Name, nil)
	c.frameCallbacks = map[int]func(){}
	rect := c.Data.Frames.FrameAtIndex(c.Data.CurrentFrame).SpriteSourceSize
	c.w, c.h = float64(rect.Width), float64(rect.Height)

	if err := c.allocateSlices(); err != nil {
		log.Println(err)
	}
	if c.Fsm == nil {
		c.Fsm = DefaultFsm()
	}
}

func (c *Comp) SetState(state string, exitFunc func()) {
	if c.exitFunc != nil {
		c.exitFunc()
	}
	c.exitFunc = exitFunc
	if c.State == state {
		return
	}
	if callback := c.Fsm.Exit[c.State]; callback != nil {
		callback(c)
	}
	c.State = state
	if err := c.Data.Play(state); err != nil {
		log.Panicf("anim: %s", err)
	}
	c.sliceCallback = nil
	c.frameCallbacks = map[int]func(){}
	// c.Data.AnimationInfo.frameCounter = 0 // TODO: This needs to happen. opened a PR: https://github.com/damienfamed75/aseprite/pull/4
	if callback := c.Fsm.Entry[c.State]; callback != nil {
		callback(c)
	}
}

func (c *Comp) Update(dt float64) {
	c.Data.Update(float32(dt))
	if c.Data.AnimationFinished() {
		nextState, ok := c.Fsm.Transitions[c.State]
		if !ok {
			nextState = vars.IdleTag
		}
		if nextState != "" {
			c.SetState(nextState, nil)
		}
	}
	currentAnimFrame := c.Data.CurrentFrame - c.Data.CurrentAnimation.From
	if frameCallback := c.frameCallbacks[currentAnimFrame]; frameCallback != nil {
		frameCallback()
		delete(c.frameCallbacks, currentAnimFrame)
	}
	if c.sliceCallback != nil {
		c.sliceCallback()
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

func (c *Comp) OnSlicePresent(sliceName string, callback SliceCallback) {
	newSlice := true
	c.sliceCallback = func() {
		slice, err := c.FrameSlice(sliceName)
		if err != nil {
			newSlice = true

			return
		}
		callback(slice, newSlice)
		newSlice = false
	}
}

func (c *Comp) OnFrame(frame int, callback func()) { c.frameCallbacks[frame] = callback }

func (c *Comp) FrameSlice(sliceName string) (bump.Rect, error) {
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

func (c *Comp) allocateSlices() error {
	c.slices = map[string]map[int]bump.Rect{}

	for _, sliceName := range []string{vars.HurtboxSliceName, vars.HitboxSliceName, vars.BlockSliceName} {
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
	utils.DrawText(screen, "ANIM:%s"+c.State, assets.TinyFont, op)
}
