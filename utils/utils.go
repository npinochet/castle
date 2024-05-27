package utils

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text" // TODO: Fix this import
	"golang.org/x/image/font"
)

type ControlPack [9][]ebiten.Key
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
		KeyRight:  {ebiten.KeyArrowRight, ebiten.KeyD},
		KeyLeft:   {ebiten.KeyArrowLeft, ebiten.KeyA},
		KeyUp:     {ebiten.KeyArrowUp, ebiten.KeyW},
		KeyDown:   {ebiten.KeyArrowDown, ebiten.KeyS},
		KeyJump:   {ebiten.KeyZ, ebiten.KeyN},
		KeyAction: {ebiten.KeyX, ebiten.KeyM},
		KeyGuard:  {ebiten.KeyC, ebiten.KeyB},
		KeyHeal:   {ebiten.KeyV, ebiten.KeyShiftLeft, ebiten.KeyShiftRight},
		//KeyDash:   {ebiten.KeySpace}, // TODO: Reconsider dash mechanic
	}
}

func (cp ControlPack) KeyDown(key ControlKey) bool {
	for _, key := range cp[key] {
		if ebiten.IsKeyPressed(key) {
			return true
		}
	}

	return false
}

func (cp ControlPack) KeyPressed(key ControlKey) bool {
	for _, key := range cp[key] {
		if inpututil.IsKeyJustPressed(key) {
			return true
		}
	}

	return false
}

func (cp ControlPack) KeyReleased(key ControlKey) bool {
	for _, key := range cp[key] {
		if inpututil.IsKeyJustReleased(key) {
			return true
		}
	}

	return false
}

func (cp ControlPack) KeyPressedBuffered(key ControlKey, timeBuffer time.Duration) func() bool {
	pressed := cp.KeyPressed(key)
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

func TextSize(text string, face font.Face) (int, int) {
	size, _ := font.BoundString(face, text)

	return size.Max.X.Ceil(), -size.Min.Y.Floor()
}

func DrawText(img *ebiten.Image, txt string, face font.Face, op *ebiten.DrawImageOptions) (int, int) {
	if op == nil {
		op = &ebiten.DrawImageOptions{}
	}
	w, h := TextSize(txt, face)
	op.GeoM.Translate(0, float64(h))
	text.DrawWithOptions(img, txt, face, op)

	return w, h
}
