package comp

import (
	"game/core"

	"github.com/damienfamed75/aseprite"
	"github.com/hajimehoshi/ebiten/v2"
)

func (ac *AsepriteComponent) IsActive() bool        { return ac.active }
func (ac *AsepriteComponent) SetActive(active bool) { ac.active = active }

type FrameCallback func(frame int)

type AnimFsm struct {
	Initial     string
	Transitions map[string]string
	Callbacks   map[string]func(*AsepriteComponent)
}

type AsepriteComponent struct {
	active         bool
	X, Y           float64
	W, H           float64
	FlipX, FlipY   bool
	Image          *ebiten.Image
	MetaData       *aseprite.File
	State          string
	Fsm            *AnimFsm
	callback       FrameCallback
	callbackFrames [2]int
}

func (ac *AsepriteComponent) Init(entity *core.Entity) {
	ac.SetState(ac.MetaData.Meta.Animations[0].Name)
}

func (ac *AsepriteComponent) SetState(state string) {
	ac.State = state
	ac.MetaData.Play(ac.State)
	ac.callback = nil
}

func (ac *AsepriteComponent) Update(dt float64) {
	ac.MetaData.Update(float32(dt))
	if ac.Fsm != nil && ac.MetaData.AnimationFinished() {
		callback := ac.Fsm.Callbacks[ac.State]
		nextState := ac.Fsm.Transitions[ac.State]
		if callback != nil {
			callback(ac)
		}
		if nextState != "" {
			ac.SetState(nextState)
		}
	}
	if ac.callback != nil {
		from := ac.MetaData.CurrentAnimation.From
		if ac.MetaData.CurrentFrame >= from+ac.callbackFrames[0] {
			ac.callback(ac.MetaData.CurrentFrame - from)
			if ac.MetaData.CurrentFrame >= from+ac.callbackFrames[1] {
				ac.callback = nil
			}
		}
	}
}

func (ac *AsepriteComponent) Draw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{}
	var dx, dy, sx, sy float64 = 0, 0, 1, 1
	if ac.FlipX {
		sx, dx = -1, ac.W
	}
	if ac.FlipY {
		sy, dy = -1, ac.H
	}
	op.GeoM.Scale(sx, sy)
	op.GeoM.Translate(ac.X, ac.Y)
	op.GeoM.Translate(dx, dy)
	op.GeoM.Concat(enitiyPos)
	sprite := ac.Image.SubImage(ac.MetaData.FrameBoundaries().Rectangle()).(*ebiten.Image)
	screen.DrawImage(sprite, op)
}

func (ac *AsepriteComponent) OnFrames(startFrame, endFrame int, callback FrameCallback) {
	ac.callbackFrames, ac.callback = [2]int{startFrame, endFrame}, callback
}
