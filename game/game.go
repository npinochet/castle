package game

import (
	"errors"
	"fmt"
	"game/assets"
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/entity"
	"game/utils"
	"game/vars"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const playerID = 25

var (
	game        = &Game{}
	canRestart  = true
	entityBinds = map[uint32]core.EntityContructor{
		26: toEntityContructor(entity.NewKnight),
		27: toEntityContructor(entity.NewGhoul),
		28: toEntityContructor(entity.NewSkeleman),
		29: toEntityContructor(entity.NewCrawler),
		87: toEntityContructor(entity.NewGram),
	}
)

func toEntityContructor[T core.Entity](contructor func(float64, float64, float64, float64, *core.Properties) T) core.EntityContructor {
	return func(x, y, w, h float64, props *core.Properties) core.Entity { return contructor(x, y, w, h, props) }
}

type Game struct{}

func (g *Game) Load() {
	worldMap := core.NewMap("maps/intro/intro.tmx", "foreground", "background")
	vars.World = core.NewWorld(float64(vars.ScreenWidth), float64(vars.ScreenHeight))
	vars.World.SetMap(worldMap, "rooms")
	worldMap.LoadBumpObjects(vars.World.Space, "collisions")
	if err := Save(); err != nil {
		log.Panicln("error saving game:", err)
	}
}

func (g *Game) Reset() {
	saveData, err := LoadSave()
	if err != nil {
		log.Panicln("error loading save:", err)
	}
	vars.Player = entity.NewPlayer(saveData.PlayerData.X, saveData.PlayerData.Y)

	vars.World.RemoveAllEntities()
	vars.World.Add(vars.Player)
	vars.World.Camera.Follow(vars.Player)
	vars.World.Map.LoadEntityObjects(vars.World, "entities", entityBinds)
}

func (g *Game) Update() error {
	if !canRestart {
		return nil
	}
	dt := 1.0 / 60
	vars.World.Update(dt)

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return errors.New("Exited")
	}
	if vars.Debug {
		debugControls()
	}

	if core.Get[*stats.Comp](vars.Player).Health <= 0 && canRestart {
		canRestart = false
		time.AfterFunc(5, func() {
			game.Reset()
			canRestart = true
		})
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{50, 60, 57, 255}) // default background color.
	if !canRestart {
		return
	}
	vars.World.Draw(screen)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(vars.ScreenWidth-16), 1)
	utils.DrawText(screen, fmt.Sprintf(`%0.2f`, ebiten.ActualFPS()), assets.TinyFont, op)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return vars.ScreenWidth, vars.ScreenHeight
}

func debugControls() {
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		body.DebugDraw = !body.DebugDraw
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		hitbox.DebugDraw = !hitbox.DebugDraw
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		ai.DebugDraw = !ai.DebugDraw
	}
	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		stats.DebugDraw = !stats.DebugDraw
	}
	if inpututil.IsKeyJustPressed(ebiten.Key5) {
		anim.DebugDraw = !anim.DebugDraw
	}
}
