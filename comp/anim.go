package comp

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

func (ac *AsepriteComponent) IsActive() bool        { return ac.active }
func (ac *AsepriteComponent) SetActive(active bool) { ac.active = active }

const (
	HurtboxSliceName = "hurtbox"
	HitboxSliceName  = "hitbox"
	BlockSliceName   = "blockbox"
)

type FrameCallback func(frame int)

type AnimFsm struct {
	Initial        string
	Transitions    map[string]string
	ExitCallbacks  map[string]func(*AsepriteComponent)
	EnterCallbacks map[string]func(*AsepriteComponent)
}

type AsepriteComponent struct {
	active       bool
	FlipX, FlipY bool
	FilesName    string
	X, Y         float64
	w, h         float64
	State        string
	Image        *ebiten.Image
	Data         *aseprite.File
	Fsm          *AnimFsm
	callback     FrameCallback
	slices       [3]*aseprite.Slice
}

func (ac *AsepriteComponent) Init(entity *core.Entity) {
	var err error
	if ac.Image, _, err = ebitenutil.NewImageFromFile(fmt.Sprintf("%s.png", ac.FilesName)); err != nil {
		log.Fatal(err)
	}
	if ac.Data, err = aseprite.Open(fmt.Sprintf("%s.json", ac.FilesName)); err != nil {
		log.Fatal(err)
	}

	ac.SetState(ac.Data.Meta.Animations[0].Name)
	rect := ac.Data.Frames.FrameAtIndex(ac.Data.CurrentFrame).SpriteSourceSize
	ac.w, ac.h = float64(rect.Width), float64(rect.Height)

	for i, sliceName := range [3]string{HurtboxSliceName, HitboxSliceName, BlockSliceName} {
		ac.slices[i] = ac.Data.Slice(sliceName)
	}
}

func (ac *AsepriteComponent) SetState(state string) {
	if ac.State == state {
		return
	}
	ac.State = state
	_ = ac.Data.Play(state)
	ac.callback = nil
	if callback := ac.Fsm.EnterCallbacks[ac.State]; callback != nil {
		callback(ac)
	}
}

func (ac *AsepriteComponent) Update(dt float64) {
	ac.Data.Update(float32(dt))
	if ac.Data.AnimationFinished() {
		if callback := ac.Fsm.ExitCallbacks[ac.State]; callback != nil {
			callback(ac)
		}
		if nextState := ac.Fsm.Transitions[ac.State]; nextState != "" {
			ac.SetState(nextState)
		}
	}
	if ac.callback != nil {
		ac.callback(ac.Data.CurrentFrame - ac.Data.CurrentAnimation.From)
	}
}

func (ac *AsepriteComponent) Draw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{}
	var x, y, sx, sy, dx, dy float64 = ac.X, ac.Y, 1, 1, 0, 0
	if ac.FlipX {
		x, sx, dx = -x, -1, math.Floor(ac.w/2)
	}
	if ac.FlipY {
		y, sy, dy = -y, -1, math.Floor(ac.h/2)
	}
	op.GeoM.Scale(sx, sy)
	op.GeoM.Translate(x, y)
	op.GeoM.Translate(dx, dy)
	op.GeoM.Concat(enitiyPos)
	sprite, _ := ac.Image.SubImage(ac.Data.FrameBoundaries().Rectangle()).(*ebiten.Image)
	screen.DrawImage(sprite, op)
}

func (ac *AsepriteComponent) OnFrames(callback FrameCallback) {
	ac.callback = callback
}

func (ac *AsepriteComponent) GetFrameHitboxes() (hurtbox, hitbox, blockbox *bump.Rect) {
	hurtbox = ac.findCurrenctSlice(ac.slices[0])
	hitbox = ac.findCurrenctSlice(ac.slices[1])
	blockbox = ac.findCurrenctSlice(ac.slices[2])

	return
}

func (ac *AsepriteComponent) findCurrenctSlice(slice *aseprite.Slice) *bump.Rect {
	if slice == nil {
		return nil
	}
	currentFrame := ac.Data.CurrentFrame
	frame := ac.Data.Frames.FrameAtIndex(currentFrame)
	ssx, ssy := float64(frame.SpriteSourceSize.X), float64(frame.SpriteSourceSize.Y)
	sw, sh := float64(frame.SourceSize.Width), float64(frame.SourceSize.Height)

	for _, key := range slice.Keys {
		if key.FrameNum == currentFrame {
			bound := key.Bounds
			rect := &bump.Rect{
				X: float64(bound.X) - ssx + ac.X, Y: float64(bound.Y) - ssy + ac.Y,
				W: float64(bound.Width), H: float64(bound.Height),
			}

			if ac.FlipX {
				rect.X += sw - rect.W - float64(bound.X)*2
			}
			if ac.FlipY {
				rect.Y += sh - rect.W - float64(bound.Y)*2
			}

			return rect
		}
	}

	return nil
}
