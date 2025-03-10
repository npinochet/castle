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
- Change background color and characters outline color. Experiment more.
- Experiment implementing a backstepping (kind of like rolling). (think about adding I frames or not, maybe just shrink the hurtbox).
	- Add quick step! Like a Dogde/dash, but small and fast that let you dodge or approach enemies quickly
		- This ^ but only the enemies
- Clean up actor.ManageAnim and body.Vx code, make it dry with player and other Actors.
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
- Maybe separate walking from running (costs stamina), maybe double tapping the direction?
	This way enemies can run towards you and reach you, dificulting escaping.
- Think of ways to make the game more brutal
	- Add a cripple mechanic when you are low on health, you move slower
	- Add a bleed mechanic when you get hit by a heavy attack, you lose health over time
- Hold attack has some bugs, the bonus damaged carries over to the normal attack sometimes, it does not reset.
- Make hurtbox change depending on frame.

- Video playtest things: https://drive.google.com/file/d/1GZ48vG0wAzkD09A6MYnGKDKIRahOqUev/view

- Demo 2 Roadmap:
	- Hide HUD when player is over it
	- Add a pattern for a background, find something sutle that fits the game
	- Try a shader with half pixels, where the pixels can be diagonal triangles
	- Two Phases:
		- As a short term goal, finish the demo with is already done
		- As a long term goal, add the rest of the features listed here

	- Phase 1:
		- Ajust combat flow speed
		- Lower stamina for player
			- Need to create a 2 phase fight, where one you do action, second you defend while waiting for stamina to recover
		- Add an intro, make just a textbox that tells you how the castle is infamous for being a place where people go in and never come out
		- Draw
			- Enemy that floats
			- More static NPCs
		- New map
			- I struggle with this
			- Add ruble somehow? Make it more ruined
				- Add grass maybe? more cracks?
			- more hazards, like spikes
				- Maybe one more hazard?
			- Limbo map that you have to clear everytime you die 3 times (more info in mobile notes...)
			- Add a new enemy (low priority)
			- Add enemy that can only be hit from behind.
			- Small enemy like a bat - Semi-completed
			- Enemies that explode on death?
			- Small enemy that rushes into you and deals damage when touching you
			- Posible cool events/encounters brainstorm:
				- A not good hidden secret that has a chest that makes (statues move/enemies appear) for a surprise.
				- Empty room with something creepy:
					- A big mirror that shows you a reflection of you, but with a different weapon (sure copilot)
					- A big dark hollow face on the wall
					- A big dark figure that disappears when you get close

	- Phase 2:
		- Posible cool events/encounters brainstorm:
			- A huge dark monster enemy that you have to avoid, give them huge health bars, make the player learn they are a part of the ecosystem and have to survive
			- A room with poison dripping from the floor, have corpses that have mushroom head. If you touch the dripping liquid, a 5 minutes timer starts where you get a mushroom head and die
			- A big tower where the rooms are the same
			- A looping laberynth
			- Have an imposing enemy be swarmed by small rats / bats
			- A dying knight that you can kill and loot for a good item
			- A room with a big enemy that is sleeping, you can kill it or sneak past it

	-- Character ideas: https://route1rodent.itch.io/16x16-rpg-character-sprite-sheet

- Demo 3 Roadmap/ideas?:
	- Add status effects
		- One where the hud hides the current health and/or stamina
	- Add items and inventory, a menu where you can see items and status effects
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
	if err := ebiten.RunGameWithOptions(&game.Game{}, op); err != nil {
		log.Fatal(err)
	}
}
