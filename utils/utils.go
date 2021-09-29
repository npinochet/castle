package utils

import (
	"game/core"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	Player core.UID = iota
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
	cp := ControlPack{}
	cp[KeyRight] = ebiten.KeyArrowRight
	cp[KeyLeft] = ebiten.KeyArrowLeft
	cp[KeyUp] = ebiten.KeyArrowUp
	cp[KeyDown] = ebiten.KeyArrowDown
	cp[KeyAction] = ebiten.KeyE
	cp[KeyGuard] = ebiten.KeyR
	return cp
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
