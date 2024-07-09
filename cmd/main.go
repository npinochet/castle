package main

import (
	"game/game"
	"game/vars"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

/* TODO
- Don't cap max speed when guarding in mid-air.
- Combos for attacks.
- Change background color and characters outline color.
- Experiment implementing a backstepping (kind of like rolling). (think about adding I frames or not, maybe just shrink the hurtbox).
	- Add quick step! Like a Dogde/dash, but small and fast that let you dodge or approach enemies quickly
		- This ^ but only the enemies
- Clean up actor.ManageAnim and body.Vx code, make it dry with player and other Actors.
- Add enemy that can only be hit from behind.
- Add enemy that jumps around.
- Experiment with partial blocking (a block does not negate all damage)
	and a system where you can attack back for a short period and gain the lost health
- Experiment hiding the enemy health bar, even for bosses.
- Experiment with shaders, change background to be more dark (maybe gradient, from blue to black?),
	maybe keep background color for a paralax layer details background.
- Lograr 2 cosas:
	- Deadly enemies
	- Dread of loosing and urgency to get to the next checkpoint
- Add item that restores your dropped loot, but spawns a high level enemy at the loot spot.
	To encourage taking other routes.
- Maybe separate waling from running (costs stamina), maybe double tapping the direction?
	This way enemies can run towards you and reach you, dificulting escaping.

- Video playtest things: https://drive.google.com/file/d/1GZ48vG0wAzkD09A6MYnGKDKIRahOqUev/view
	- Pixel font too hard to read

- Demo 2 Roadmap:
	- Ajust combat flow speed
	- Lower stamina for player
		- Need to create a 2 phase fight, where one you do action, second you defend while waiting for stamina to recover
	- New map

	- Add a new enemy (low priority)
*/

func main() {
	ebiten.SetWindowSize(vars.ScreenWidth*vars.Scale, vars.ScreenHeight*vars.Scale)
	ebiten.SetWindowTitle("Castle")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(!vars.Debug)

	// TODO: Prevent macOS from using Metal API and panic.
	op := &ebiten.RunGameOptions{}
	if runtime.GOOS == "darwin" {
		op.GraphicsLibrary = ebiten.GraphicsLibraryOpenGL
	}
	if os.Args[len(os.Args)-1] == "test" {
		go func() {
			time.Sleep(2 * time.Second)
			os.Exit(0)
		}()
	}

	if err := ebiten.RunGameWithOptions(&game.Game{}, op); err != nil {
		log.Fatal(err)
	}
}
