package main

import (
	"errors"
	"fmt"
	"game/core"
	"game/entity"
	"game/utils"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

/* TODO
- Add animation tiles on update for Tiled map.
- Maybe stop time while camera transition is playing, and move follower entity to border?
- AI component for enemies.
- Don't cap max speed when guarding in mid-air.
- Add slopes.
- Combos for attacks.
- Think of a system to manage animations.
- Make more enemies, make some of them shoot arrows.
- Make actor default params presets.

- Clean up actor.ManageAnim and body.Vx code, make it sry with player and other Actors.
- Add a Timeout system for AI states.
- Clean up AI code, Make a default AI behaviour for actors if none are present. Make it tweekable with other params maybe.
*/

const (
	scale                     = 4
	screenWidth, screenHeight = 160, 96 // 320, 240.
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
	g.world.SetMap(core.NewMap("maps/intro/intro.tmx", "foreground", "background"), "rooms")

	playerX, playerY, err := g.world.Map.FindObjectPosition("entities", 25)
	if err != nil {
		log.Println("Error finding player entity:", err)
	}
	player = entity.NewPlayer(playerX, playerY, nil)
	g.world.Camera.Follow(player, 14, 14)
	g.world.AddEntity(&player.Entity).ID = utils.PlayerUID

	g.world.Map.LoadBumpObjects(g.world.Space, "collisions")
	g.world.Map.LoadEntityObjects(g.world, "entities", map[uint32]core.EntityContructor{
		26: entity.NewKnight,
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
	screen.Fill(color.RGBA{50, 60, 57, 255}) // default background color.
	g.world.Draw(screen)
	if g.world.Debug {
		ebitenutil.DebugPrint(screen, fmt.Sprintf(`%0.2f`, ebiten.CurrentTPS()))
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth*scale, screenHeight*scale)
	ebiten.SetWindowTitle("Castle")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game.init()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
