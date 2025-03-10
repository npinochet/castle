package main_test

import (
	"errors"
	"game/game"
	"log"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

type testGame struct {
	game   *game.Game
	frames int
}

func (g *testGame) Update() error {
	if g.frames--; g.frames <= 0 {
		return ebiten.Termination
	}

	return g.game.Update()
}

func (g *testGame) Draw(screen *ebiten.Image)  { g.game.Draw(screen) }
func (g *testGame) Layout(w, h int) (int, int) { return g.game.Layout(w, h) }

func TestMain(_ *testing.M) {
	if err := ebiten.RunGame(&testGame{game: &game.Game{}, frames: 10}); err != nil {
		if errors.Is(err, ebiten.Termination) {
			os.Exit(0)
		}
		log.Fatal(err)
	}
}
