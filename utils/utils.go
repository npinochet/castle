package utils

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

type ControlPack [9]ebiten.Key
type ControlKey int

const (
	KeyRight ControlKey = iota
	KeyLeft
	KeyUp
	KeyDown
	KeyJump
	KeyAction
	KeyGuard
	KeyHeal
	KeyDash
)

var buffer = map[ControlKey]bool{}
var bufferTimers = map[ControlKey]*time.Timer{}

func NewControlPack() ControlPack {
	return ControlPack{
		KeyRight:  ebiten.KeyArrowRight,
		KeyLeft:   ebiten.KeyArrowLeft,
		KeyUp:     ebiten.KeyArrowUp,
		KeyDown:   ebiten.KeyArrowDown,
		KeyJump:   ebiten.KeyZ,
		KeyAction: ebiten.KeyX,
		KeyGuard:  ebiten.KeyC,
		KeyHeal:   ebiten.KeyV,
		KeyDash:   ebiten.KeySpace, // TODO: Reconsider dash mechanic
	}
}

func (cp ControlPack) KeyDown(key ControlKey) bool {
	return ebiten.IsKeyPressed(cp[key])
}

func (cp ControlPack) KeyPressed(key ControlKey) bool {
	return inpututil.IsKeyJustPressed(cp[key])
}

func (cp ControlPack) KeyPressedBuffered(key ControlKey, timeBuffer time.Duration) func() bool {
	pressed := inpututil.IsKeyJustPressed(cp[key])
	if pressed {
		buffer[key] = true
		if bufferTimers[key] != nil {
			bufferTimers[key].Stop()
		}
		bufferTimers[key] = time.AfterFunc(timeBuffer, func() { buffer[key] = false })
	}

	return func() bool {
		pressed := buffer[key]
		buffer[key] = false
		if bufferTimers[key] != nil {
			bufferTimers[key].Stop()
		}

		return pressed
	}
}

func (cp ControlPack) KeyReleased(key ControlKey) bool {
	return inpututil.IsKeyJustReleased(cp[key])
}

func Distante(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x1-x2, 2) + math.Pow(y1-y2, 2))
}

func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}

	return false
}

func DrawText(dst *ebiten.Image, txt string, face font.Face, op *ebiten.DrawImageOptions) (int, int) {
	if op == nil {
		op = &ebiten.DrawImageOptions{}
	}
	size, _ := font.BoundString(face, txt)
	w, h := size.Max.X.Ceil(), -size.Min.Y.Floor()
	op.GeoM.Translate(0, float64(h))
	text.DrawWithOptions(dst, txt, face, op)

	return w, h
}
