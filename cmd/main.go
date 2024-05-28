package main

import (
	"game/game"
	"game/vars"
	"log"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
)

/* TODO
- Don't cap max speed when guarding in mid-air.
- Combos for attacks.
- Change background color and characters outline color.
- Rethink Poise mechanic, is shouldn't be a bar that increses with time, it should be more like a health that resets.
- Experiment implementing a backstepping (kind of like rolling). (think about adding I frames or not, maybe just shrink the hurtbox).

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
- Experiment with shaders, change background to be more dark (maybe gradient, from blue to black?), maybe keep background color for a paralax layer details background.
- Add quick step! Like a Dogde/dash, but small and fast that let you dodge or approach enemies quickly
	- This ^ but only the enemies
- lograr 2 cosas:
	- Deadly enemies
	- Dread of loosing and urgency to get to the next checkpoint
- Add item that restores your dropped loot, but spawns a high level enemy at the loot spot. To encourage taking other routes.

- Today's TODO:
- Demo MVP Steps:
	- (Optional) Show controls on start screen
	- Build Map
	- Death drop
	- Add end room message
	- Add some polish
		- JUICE UP COMBAT, IM TALKING STOP TIME, PARTICLE EFFECTS, FLASHING BABY
	- Add Mage Enemy?
	- Add Heavy attack?
	- Make enemies respawn when yo get back to the room before after
	- Maybe add a new combo for the player?
*/

func main() {
	ebiten.SetWindowSize(vars.ScreenWidth*vars.Scale, vars.ScreenHeight*vars.Scale)
	ebiten.SetWindowTitle("Castle")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(false)

	// TODO: Prevent macOS from using Metal API and panic.
	op := &ebiten.RunGameOptions{}
	if runtime.GOOS == "darwin" {
		op.GraphicsLibrary = ebiten.GraphicsLibraryOpenGL
	}
	if err := ebiten.RunGameWithOptions(&game.Game{}, op); err != nil {
		log.Fatal(err)
	}
}
