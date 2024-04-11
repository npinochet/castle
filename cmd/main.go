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
	"game/vars"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

/* TODO
- JUICE UP COMBAT, IM TALKING STOP TIME, PARTICLE EFFECTS, FLASHING BABY
- Add animation tiles on update for Tiled map.
- Maybe stop time while camera transition is playing, and move follower entity to border?
- Don't cap max speed when guarding in mid-air.
- Combos for attacks.
- Change background color and characters outline color.
- Rethink Poise mechanic, is shouldn't be a bar that increses with time, it should be more like a health that resets.
- Experiment implementing a backstepping (kind of life rolling). (think about adding I frames or not, maybe just shrink the hurtbox).

- Clean up actor.ManageAnim and body.Vx code, make it sry with player and other Actors.
- Sometimes the enemy can cut off the stagger animation somehow.
- Cannot jump when going down slope, body.Ground is mostly false, this can be solved with coyote time.

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
- Add enemy that can only be hit from behind.
- Add enemy that jumps around.
- Add ability to use a over heal with a consumable to boost attack damage
- Experiment with partial blocking (a block does not negate all damage) and a system where you can attack back for a short period and
	gain the lost health
- Experiment hiding the enemy health bar, even for bosses.


- Today's TODO:
	- Experiment with ideas above.

*/

const (
	playerID = 25
)

var (
	game       = &Game{}
	player     *entity.Player
	canRestart = true
)

type Game struct{}

func (g *Game) init() {
	obj, err := vars.World.Map.FindObjectFromTileID(playerID, "entities")
	if err != nil {
		log.Println("main: error finding player entity:", err)
	}
	player = entity.NewPlayer(obj.X, obj.Y, nil)
	vars.World.Camera.Follow(player)
	vars.World.Add(player)
	entity.PlayerRef = player

	vars.World.Map.LoadBumpObjects(vars.World.Space, "collisions")
	vars.World.Map.LoadEntityObjects(vars.World, "entities", map[uint32]core.EntityContructor{
		26: func(x, y, w, h float64, props *core.Properties) core.Entity {
			return entity.NewKnight(x, y, w, h, props)
		},
		27: func(x, y, w, h float64, props *core.Properties) core.Entity {
			return entity.NewGhoul(x, y, w, h, props)
		},
		28: func(x, y, w, h float64, props *core.Properties) core.Entity {
			return entity.NewSkeleman(x, y, w, h, props)
		},
		29: func(x, y, w, h float64, props *core.Properties) core.Entity {
			return entity.NewCrawler(x, y, w, h, props)
		},
		87: func(x, y, w, h float64, props *core.Properties) core.Entity {
			return entity.NewGram(x, y, w, h, props)
		},
	})
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

	if core.Get[*stats.Comp](player).Health <= 0 && canRestart {
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
	vars.World.Draw(screen)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(vars.ScreenWidth-16), 1)
	utils.DrawText(screen, fmt.Sprintf(`%0.2f`, ebiten.ActualFPS()), assets.TinyFont, op)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return vars.ScreenWidth, vars.ScreenHeight
}

func main() {
	ebiten.SetWindowSize(vars.ScreenWidth*vars.Scale, vars.ScreenHeight*vars.Scale)
	ebiten.SetWindowTitle("Castle")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	// ebiten.SetVsyncEnabled(false)

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
