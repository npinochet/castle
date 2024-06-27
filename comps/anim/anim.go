package anim

import (
	"fmt"
	"game/assets"
	"game/core"
	"game/libs/bump"
	"game/utils"
	"game/vars"
	"image/color"
	"log"
	"math"
	"slices"

	"github.com/damienfamed75/aseprite"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var DebugDraw = false

type Fsm struct {
	Initial     string
	Transitions map[string]string
}

type SliceCallback func(slice bump.Rect, segmented bool)

type stateEffect struct {
	restore func()
	states  []string
}

type Comp struct {
	FilesName      string
	OX, OY         float64
	OXFlip, OYFlip float64
	FlipX, FlipY   bool
	Layer          int
	ColorScale     color.Color
	Fsm            *Fsm

	State          string
	Image          *ebiten.Image
	Data           *aseprite.File
	w, h           float64
	slices         map[string]map[int]bump.Rect
	stateEffect    *stateEffect
	sliceCallback  func()
	frameCallbacks map[int]func()
}

func (c *Comp) Init(_ core.Entity) {
	var err error
	if c.Image, _, err = ebitenutil.NewImageFromFileSystem(assets.FS, c.FilesName+".png"); err != nil {
		log.Panic(err)
	}
	animData, err := assets.FS.ReadFile(c.FilesName + ".json")
	if err != nil {
		log.Panic(err)
	}
	if c.Data, err = aseprite.NewFile(animData); err != nil {
		log.Panic(err)
	}
	if c.Fsm == nil {
		c.Fsm = DefaultFsm()
	}

	if c.ColorScale == nil {
		c.ColorScale = color.White
	}
	c.SetState(c.Data.Meta.Animations[0].Name)
	c.frameCallbacks = map[int]func(){}
	rect := c.Data.Frames.FrameAtIndex(c.Data.CurrentFrame).SpriteSourceSize
	c.w, c.h = float64(rect.Width), float64(rect.Height)

	c.allocateSlices()
}

func (c *Comp) Remove() {}

func (c *Comp) SetState(state string) {
	if c.State == state {
		return
	}
	c.State = state
	if err := c.Data.Play(state); err != nil {
		log.Printf("anim: %s", err)

		return
	}
	if c.stateEffect != nil && !slices.Contains(c.stateEffect.states, c.State) {
		c.stateEffect.restore()
		c.stateEffect = nil
	}
	c.sliceCallback = nil
	c.frameCallbacks = map[int]func(){}
}

func (c *Comp) Update(dt float64) {
	c.Data.Update(float32(dt))
	if c.Data.AnimationFinished() {
		nextState, ok := c.Fsm.Transitions[c.State]
		if !ok {
			nextState = c.Fsm.Initial
		}
		if nextState != "" {
			c.SetState(nextState)
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

func (c *Comp) Draw(pipeline *core.Pipeline, entityPos ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{}
	var x, y, sx, sy, dx, dy float64 = c.OX, c.OY, 1, 1, 0, 0
	if c.FlipX {
		sx, dx = -1, math.Floor(c.w/2)+c.OXFlip
	}
	if c.FlipY {
		sy, dy = -1, math.Floor(c.h/2)+c.OYFlip
	}
	op.GeoM.Scale(sx, sy)
	op.GeoM.Translate(x+dx, y+dy)
	op.GeoM.Concat(entityPos)
	op.ColorScale.ScaleWithColor(c.ColorScale)
	sprite, _ := c.Image.SubImage(c.Data.FrameBoundaries().Rectangle()).(*ebiten.Image)
	pipeline.AddDraw(vars.PipelineScreenTag, c.Layer, func(screen *ebiten.Image) { screen.DrawImage(sprite, op) })
	normalOp := &colorm.DrawImageOptions{GeoM: op.GeoM}
	pipeline.AddDraw(vars.PipelineNormalMapTag, c.Layer, func(normalMap *ebiten.Image) {
		colorm.DrawImage(normalMap, sprite, FillNormalMaskColorM, normalOp)
	})
	if DebugDraw {
		pipeline.AddDraw(vars.PipelineScreenTag, vars.PipelineUILayer, func(screen *ebiten.Image) { c.debugDraw(screen, entityPos) })
	}
}

func (c *Comp) OnSlicePresent(sliceName string, callback SliceCallback) {
	if callback == nil {
		c.sliceCallback = nil

		return
	}
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

func (c *Comp) SetStateEffect(applyAndGetRestore func() func(), forStates ...string) {
	if c.stateEffect != nil {
		c.stateEffect.restore()
	}
	c.stateEffect = &stateEffect{applyAndGetRestore(), forStates}
}

func (c *Comp) allocateSlices() {
	c.slices = map[string]map[int]bump.Rect{}

	for _, slice := range c.Data.Meta.Slices {
		c.slices[slice.Name] = map[int]bump.Rect{}
		for _, key := range slice.Keys {
			sss := c.Data.Frames.FrameAtIndex(key.FrameNum).SpriteSourceSize

			bound := key.Bounds
			c.slices[slice.Name][key.FrameNum] = bump.Rect{
				X: float64(bound.X) - float64(sss.X), Y: float64(bound.Y) - float64(sss.Y),
				W: float64(bound.Width), H: float64(bound.Height),
			}
		}
	}
}

func (c *Comp) debugDraw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-5, -22)
	utils.DrawText(screen, "ANIM:"+c.State, assets.TinyFont, op)
}
