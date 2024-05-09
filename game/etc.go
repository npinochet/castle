package game

import (
	"game/core"
	"game/vars"

	"github.com/hajimehoshi/ebiten/v2"
)

type Transition interface {
	Init()
	Update(dt float64) bool
	Draw(screen *ebiten.Image)
}

func toEntityContructor[T core.Entity](contructor func(float64, float64, float64, float64, *core.Properties) T) core.EntityContructor {
	return func(x, y, w, h float64, props *core.Properties) core.Entity { return contructor(x, y, w, h, props) }
}

type Manager struct{}

func (m Manager) Reset()      { Reset() }
func (m Manager) Save() error { return Save() }
func init()                   { vars.SetGameManager(Manager{}) }
