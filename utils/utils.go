package utils

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

const PlayerUID = 100

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

func DrawText(dst *ebiten.Image, txt string, face font.Face, op *ebiten.DrawImageOptions) {
	size := text.BoundString(face, txt)
	op.GeoM.Translate(0, float64(-size.Min.Y))
	text.DrawWithOptions(dst, txt, face, op)
}
