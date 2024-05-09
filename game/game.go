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
	"game/maps"
	"game/utils"
	"game/vars"
	"image/color"
	"log"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const playerID = 25

var (
	backgroundColor = color.RGBA{50, 60, 57, 255}
	entityBinds     = map[uint32]core.EntityContructor{
		26:  toEntityContructor(entity.NewKnight),
		27:  toEntityContructor(entity.NewGhoul),
		28:  toEntityContructor(entity.NewSkeleman),
		29:  toEntityContructor(entity.NewCrawler),
		87:  toEntityContructor(entity.NewGram),
		149: toEntityContructor(entity.NewChest),
		150: toEntityContructor(entity.NewGrave),
		151: toEntityContructor(entity.NewDoor),
	}
	deathTransition Transition
)

func Load() {
	worldMap := core.NewMap("intro/intro.tmx", "foreground", "background", maps.IntroFS)
	vars.World = core.NewWorld(float64(vars.ScreenWidth), float64(vars.ScreenHeight))
	vars.World.SetMap(worldMap, "rooms")
	worldMap.LoadBumpObjects(vars.World.Space, "collisions")
	Reset()
}

func Reset() {
	saveData, err := LoadSave()
	if err != nil {
		log.Panicln("error loading save:", err)
	}
	ApplySaveData(saveData)

	vars.World.Speed = 1
	vars.World.RemoveAllEntities()
	vars.World.Add(vars.Player)
	vars.World.Camera.Follow(vars.Player)
	vars.World.Map.LoadEntityObjects(vars.World, "entities", entityBinds)
	runtime.GC()
}

type Game struct{}

func (g *Game) Update() error {
	dt := 1.0 / 60
	vars.World.Update(dt)

	if core.Get[*stats.Comp](vars.Player).Health <= 0 {
		if deathTransition == nil {
			deathTransition = &DeathTransition{}
			deathTransition.Init()
		}
	}
	if deathTransition != nil {
		if done := deathTransition.Update(dt); done {
			deathTransition = nil
		}
	}

	if vars.Debug {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return errors.New("exited")
		}
		debugControls()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColor)
	vars.World.Draw(screen)
	if deathTransition != nil {
		deathTransition.Draw(screen)
	}

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
