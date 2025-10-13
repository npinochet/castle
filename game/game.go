package game

import (
	"fmt"
	"game/assets"
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/entity"
	"game/entity/actor"
	"game/maps"
	"game/shader"
	"game/utils"
	"game/vars"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	playerID = 25
	torchGID = 378
)

/*
package main

type Entity interface{ Name() string }
type Constructor func(name string) Entity

type Player struct{ name string }

func (p *Player) Name() string      { return p.name }
func NewPlayer(name string) *Player { return &Player{name} }

var entities = map[string]Constructor{}

func AddEntity(entityID string, constructor Constructor) { entities[entityID] = constructor }

func main() { AddEntity("Hero", NewPlayer) }
*/

var (
	backgroundColor = color.RGBA{50, 60, 57, 255}
	pixelScreen     = ebiten.NewImage(vars.ScreenWidth, vars.ScreenHeight)
	pipeline        = core.NewPipeline()
	entityBinds     = map[uint32]core.EntityContructor{
		26: toEntityContructor(entity.NewKnight),
		27: toEntityContructor(entity.NewGhoul),
		28: toEntityContructor(entity.NewSkeleman),
		29: toEntityContructor(entity.NewCrawler),
		30: toEntityContructor(entity.NewRat),
		31: toEntityContructor(entity.NewBat),
		32: toEntityContructor(entity.NewEnt),
		87: toEntityContructor(entity.NewGram),
		88: toEntityContructor(entity.NewFerragus),
		89: toEntityContructor(entity.NewOscar),
		90: toEntityContructor(entity.NewAcedian),
		//89:  toEntityContructor(entity.New??),
		149: toEntityContructor(entity.NewChest),
		150: toEntityContructor(entity.NewGrave),
		151: toEntityContructor(entity.NewDoor),
		152: toEntityContructor(entity.NewSpike),
		153: toEntityContructor(entity.NewFakeWall),
	}
	restartTransition, deathTransition Transition
)

type Transition interface {
	Init()
	Update(dt float64) bool
	Draw(screen *ebiten.Image)
}

func toEntityContructor[T core.Entity](contructor func(x, y, w, h float64, p *core.Properties) T) core.EntityContructor {
	return func(x, y, w, h float64, p *core.Properties) core.Entity { return contructor(x, y, w, h, p) }
}

func Load() {
	actor.DieParticle = func(e core.Entity) core.Entity { return entity.NewFlake(e) }
	//worldMap := core.NewMap("intro/intro.tmx", 1, maps.IntroFS, vars.PipelineScreenTag, vars.PipelineNormalMapTag)
	worldMap := core.NewMap("intro/playground_imp.tmx", 1, maps.IntroFS, vars.PipelineScreenTag, vars.PipelineNormalMapTag)
	vars.World = core.NewWorld(float64(vars.ScreenWidth), float64(vars.ScreenHeight))
	vars.World.SetMap(worldMap, "rooms")
	worldMap.LoadBumpObjects(vars.World.Space, "collisions")
	shader.Load(worldMap, []uint32{torchGID, 931, 993})
	Reset()
}

func Reset() {
	saveData, err := LoadSave()
	if err != nil {
		log.Panicln("error loading save:", err)
	}

	vars.World.Speed = 1
	vars.World.RemoveAll()
	vars.World.Map.LoadEntityObjects(vars.World, "entities", entityBinds)
	LoadMapEvents(vars.World.Map)
	vars.World.Update(0)
	ApplySaveData(saveData)
	vars.World.Add(vars.Player)
	vars.World.Camera.Follow(vars.Player)
	vars.World.Update(0)
}

type Game struct{ loaded bool }

func (g *Game) Update() error {
	if !g.loaded {
		g.loaded = true
		Load()
	}
	dt := 1.0 / 60
	vars.World.Update(dt)
	shader.Update(dt)
	if vars.SaveGame {
		vars.SaveGame = false
		if err := Save(); err != nil {
			return err
		}
	}
	if restartTransition == nil && vars.ResetGame {
		restartTransition = &RestartTransition{}
		restartTransition.Init()
	}
	if restartTransition != nil {
		vars.ResetGame = false
		if done := restartTransition.Update(dt); done {
			restartTransition = nil
		}
	}

	if deathTransition == nil && core.Get[*stats.Comp](vars.Player).Health <= 0 {
		deathTransition = &DeathTransition{}
		deathTransition.Init()
	}
	if deathTransition != nil {
		if done := deathTransition.Update(dt); done {
			deathTransition = nil
		}
	}

	if vars.Debug {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return ebiten.Termination
		}
		debugControls()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	pixelScreen.Fill(backgroundColor)
	vars.World.Draw(pipeline)
	pipeline.Compose(vars.PipelineScreenTag, pixelScreen)
	shader.DrawLights(pipeline, pixelScreen)
	pipeline.DisposeAll()

	if restartTransition != nil {
		restartTransition.Draw(pixelScreen)
	}
	if deathTransition != nil {
		deathTransition.Draw(pixelScreen)
	}

	op := &ebiten.DrawImageOptions{}
	fps := fmt.Sprintf("%0.2f", ebiten.ActualFPS())
	w, _ := utils.TextSize(fps, assets.NanoFont)
	op.GeoM.Translate(float64(vars.ScreenWidth-w-1), 1)
	utils.DrawText(pixelScreen, fps, assets.NanoFont, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(vars.Scale), float64(vars.Scale))
	screen.DrawImage(pixelScreen, op)
	shader.DrawPhosphore(pipeline, screen)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return vars.ScreenWidth * vars.Scale, vars.ScreenHeight * vars.Scale
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
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		shader.Lights = !shader.Lights
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		shader.Phosphore = !shader.Phosphore
	}
}
