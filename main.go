package main

import (
	"errors"
	"fmt"
	"game/core"
	"game/entity"
	"game/utils"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

/* lista
- Add animation tiles on update for Tiled map
- Maybe stop time while camera transition is playing, and move follower entity to border?
- IA component for enemies.
*/

const (
	scale                     = 3
	screenWidth, screenHeight = 160, 90 // 320, 240
)

var (
	game   = &Game{}
	player *entity.Player
)

type Game struct {
	world *core.World
}

func (g *Game) init() {
	g.world = core.NewWorld(screenWidth, screenHeight)
	g.world.SetMap(core.NewMap("maps/test/first/test.tmx", "foreground", "background"), "rooms")

	playerX, playerY, err := g.world.Map.FindObjectPosition("entities", 114)
	if err != nil {
		log.Println("Error finding player entity:", err)
	}
	player = entity.NewPlayer(playerX, playerY, nil)
	g.world.Camera.Follow(player, 14, 14)
	g.world.AddEntity(&player.Entity).ID = utils.PlayerUID

	g.world.Map.LoadBumpObjects(g.world.Space, "collision")
	g.world.Map.LoadEntityObjects(g.world, "entities", map[uint32]core.EntityContructor{
		115: entity.NewKnight,
	})
}

func (g *Game) Update() error {
	dt := 1.0 / 60
	g.world.Update(dt)

	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		return errors.New("Exited")
	}
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.world.Debug = !g.world.Debug
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.world.Draw(screen)
	if g.world.Debug {
		ebitenutil.DebugPrint(screen, fmt.Sprintf(`TPS: %0.2f`, ebiten.CurrentTPS()))
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth*scale, screenHeight*scale)
	ebiten.SetWindowTitle("Castle")
	ebiten.SetWindowResizable(true)

	game.init()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
