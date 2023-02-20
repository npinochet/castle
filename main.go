package main

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
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
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
- Change background color and characters outline color.
- Rethink Poise mechanic, is shouldn't be a bar that increses with time, it should be more like a health that resets.
- Implement estus flasks.
- Implement backstepping (kind of life rolling). (think about adding I frames or not, maybe just shrink the hurtbox).
- Consider scapping core.Entity all together, use interface{} (pointer) as entities and use Actor for everything.
	Every Comp will have an actor referencing it's owner.


- Clean up actor.ManageAnim and body.Vx code, make it sry with player and other Actors.
- Add a Timeout system for AI states.
- Clean up AI code, Make a default AI behaviour for actors if none are present. Make it tweekable with other params maybe.
- Think of movement accion or states the anim component can have.
- Sometimes the enemy can cut off the stagger animation somehow.
- Can not jump when going down slope, body.Ground is mostly false, this can be solved with coyote time.

-- Dark Souls Combat Findings
- When guard breaks while guarding (stamina < 0) the stagger animation is longer than poise break.
- Poise break are really small, just to interrupt animation.
- When using a big shield (stability high) and guarding, an enemy attack can be deflect.
- When blocking an attack, a little stagger animation is played.
- Stagger animation can be reset if hit again.
	- Only the player can be stun locked. -> poise is reset only after stagger animation finishes.
- No invinsibility frames after getting hit.
	- Each enemy can hit the player after being in contact with the hitbox once.
	- If the hitbox gets away from the player hurtbox in one frame and then it overlaps again on the next frame, it should hit again.
- Add teams to actor, the AI should only target player and not other enemies (unless hit by enemy).
- Maybe replace FSM with behaviour tree (ref: https://github.com/askft/go-behave)
*/

const (
	scale                     = 4
	screenWidth, screenHeight = 160, 96 // 320, 240.

	playerID = 25
)

var (
	game       = &Game{}
	player     *entity.Player
	canRestart = true
)

type Game struct {
	world *core.World
}

func (g *Game) init() {
	g.world = core.NewWorld(screenWidth, screenHeight)
	g.world.SetMap(core.NewMap("maps/intro/intro.tmx", "foreground", "background"), "rooms")

	obj, err := g.world.Map.FindObjectFromTileID(playerID, "entities")
	if err != nil {
		log.Println("Error finding player entity:", err)
	}
	player = entity.NewPlayer(obj.X, obj.Y, nil)
	g.world.Camera.Follow(player)
	g.world.AddEntity(&player.Entity)

	g.world.Map.LoadBumpObjects(g.world.Space, "collisions")
	g.world.Map.LoadEntityObjects(g.world, "entities", map[uint32]core.EntityContructor{
		26: entity.NewKnight,
		27: entity.NewGhoul,
		28: entity.NewSkeleman,
		29: entity.NewCrawler,
		87: entity.NewGram,
	})
}

func (g *Game) Update() error {
	if !canRestart {
		return nil
	}

	dt := 1.0 / 60
	g.world.Update(dt)

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return errors.New("Exited")
	}
	debugControls()

	if player.Stats.Health <= 0 && canRestart {
		canRestart = false
		time.AfterFunc(2, func() {
			game.init()
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
	g.world.Draw(screen)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenWidth-16, 1)
	utils.DrawText(screen, fmt.Sprintf(`%0.2f`, ebiten.ActualFPS()), assets.TinyFont, op)
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
