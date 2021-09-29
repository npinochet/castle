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

/*TODO
- Add an invinsibility system, probably add to hitbox component?
- Add animation tiles on update for Tiled map
- Move player code and common components (body, animation, states, hitboxes, movement functions) to an Actor entity somehow?
- Stats component? For health, stamina, poise, attack, etc calculations and display?
- Maybe stop time while camera transition is playing, and move follower entity to border?
- IA component for enemies
- Control component?
*/

const (
	scale        = 3
	screenWidth  = 160 //320
	screenHeight = 90  //240
)

var player *entity.Player

type Game struct {
	inited bool
	world  *core.World
}

func (g *Game) init() {
	defer func() { g.inited = true }()

	camera := core.NewCamera(screenWidth, screenHeight, 9)
	g.world = core.NewWorld()
	g.world.LoadTiledMap(core.NewTiledMap("maps/test/first/test.tmx", "foreground", "background"), camera, "rooms")

	playerX, playerY, err := g.world.TiledMap.FindObjectPosition("entities", 114)
	if err != nil {
		fmt.Println("Error finding player entity:", err.Error())
	}
	player = entity.NewPlayer(playerX, playerY, nil)
	camera.Follow(&player.Entity, 14, 14)
	g.world.AddEntity(&player.Entity).Id = utils.Player

	g.world.TiledMap.LoadBumpObjects(g.world.Space, "collision", true)
	g.world.TiledMap.LoadEntityObjects(g.world, "entities", map[uint32]core.EntityContructor{
		115: entity.NewKnight,
	})
}

func (g *Game) Update() error {
	dt := 1.0 / 60
	if !g.inited {
		g.init()
	}

	g.world.Update(dt)
	g.world.Camera.Update(dt, 9)

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

var game *Game = &Game{}

func main() {
	ebiten.SetWindowSize(screenWidth*scale, screenHeight*scale)
	ebiten.SetWindowTitle("Castle")
	ebiten.SetWindowResizable(true)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
