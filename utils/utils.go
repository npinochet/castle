package utils

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	PlayerUID = iota
)

type ControlPack [6]ebiten.Key
type ControlKey int

const (
	KeyRight ControlKey = iota
	KeyLeft
	KeyUp
	KeyDown
	KeyAction
	KeyGuard
)

func NewControlPack() ControlPack {
	return ControlPack{
		KeyRight:  ebiten.KeyArrowRight,
		KeyLeft:   ebiten.KeyArrowLeft,
		KeyUp:     ebiten.KeyArrowUp,
		KeyDown:   ebiten.KeyArrowDown,
		KeyAction: ebiten.KeyE,
		KeyGuard:  ebiten.KeyR,
	}
}

func (cp ControlPack) KeyDown(key ControlKey) bool {
	return ebiten.IsKeyPressed(cp[key])
}

func (cp ControlPack) KeyPressed(key ControlKey) bool {
	return inpututil.IsKeyJustPressed(cp[key])
}

func (cp ControlPack) KeyReleased(key ControlKey) bool {
	return inpututil.IsKeyJustReleased(cp[key])
}
