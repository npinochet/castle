# Game Design Document
**Working title**: Mini souls: Eternal Castle whatever
# Inspirations
* Dark Souls (Maybe a little too much).
* "El inmortal", by Jorge Luis Borges.

## Setting
### Introduction
A curse has been casted upon the the king's castle.
From one day to the next a barrier surrounding the whole castle appeared, keeping everyone trapped inside and, by an unknown reason, unable to die.

Many years or centuries have passed and the barrier is still intact. Everyone is in the brink of insanity after being confined together for so long without contact with the outer world. Little by little the inhabitants start losing their humanity, acting on their most primal impulses, everyone is old, hurting, tired, hungry and cold.

The player finally wakes up and finds himself inside the castle after been [asleep/dead/unconscious]? for many years. He has in his possession the only item that can truly lift the curse and bring death to a someone, but it only has a one time use and he's saving it for someone special. Most inhabitants you encounter will do anything to obtain your item.

After year or centuries of being prisoners on the castle, strange social bonds and groups started to emerge, a whole new ecosystem has been developed as people try their best to keep their little sanity left.

### Objective
* Make the world feel alive and believable, not "gamey".
* Make a small map, quality over quantity.

### Possible covenants or social groups
* Those who give up on life and embrace the suffering, always seeking and hoping to end it all.
* Strong minded guards who keep being loyal to their king and will for years to come.
* Those who start to see the curse as a gift and start worshiping the unknown deity who brought this gift upon them.
* People who succumber into their basic needs and pleasures:
    * Greed: The ones who still clings to social values of the outside and want to hoard all the gold in the castle.
    * Gluttony: The ones who still clings to the old ways of life and seek to eat all the food in the castle.

### Possible castle areas
* Dungeon.
* Graveyard.
* Throne room (This is just a room, can't be an area really).
* Garden.

### Possible area themes
* A dungeon full of deadly attractions, where people can pay (to a greedy man) for a *supposed* certain death.
* A garden/jungle: As the years passed, plants can grow limitless.
* A Library housing mindless intellectuals who worship books

#### Imposed limitations
* Must have at most 3 areas, presenting one covenant/group per area.
* At least 3 enemies types per area. (3*3 = 9 enemies).
* There must be a central kind of central area hub.
* There must be a small tutorial area that directly connect to the central hub.
* Areas must connect to each other and have unlockable shortcuts, these can be elevator of ladders.
* Maybe at one point be captured by a cult and be "teleported" as an introduction to another area?
    * It would be really cool if I can replicate the wondrous felling of Seath's encounter in the Dark Soul's "The Duke's Archives".

### Possible first area: *The jester way*
The dungeons of the castle has been taken by the castle's jester, which he used to make a series of deadly attractions (with the dungeons torture devices) and charge an entry fee to *surely end yourself*. At the end one can find the jester over a useless pile of gold.
* Can or not have a lot of traps and saw blades and stuff. May or not be inspired by Sen's Fortress xD.

#### Area Enemies
* Soulless wanderers: Kind of like Dark Souls basic hollows.
* Jester minions: Disfigured (from all the failed suicides) followers of the jester attractions.
* Have to think of one more here :v
* Boss: The Jester. Duh.


## Mechanics
A 2D deliberate hack and slash game focused on animations and stamina management.

### Objectives
* To capture the filling of dread that comes from going deeper and deeper into the game without finding a checkpoint
    * Try to replicate the feeling of crossing the bridge in undead parish to the church and pressing on in Dark Souls for the first time.
    * One aspect to achieve this is raising the game's difficulty (hopefully in terms of skills and not numbers), make sure the player dies a lot.
        * Caveats: **Don't make health sponge enemies**. Make them hit harder if I must do.
    * Can also punish the player for dying, with lost progress or other mechanics.
* Have a semi-maze like map, where you have to be aware of your surrounding to navigate it.
* **Do not make a CastleVania: No item gates, minimal backtracking**.

### Main game mechanics
#### Introduction
Most enemies will be humanoid and share the same skills, abilities and mechanics of the player.

#### Stamina
The player will have a stamina bar which:
* Depletes with every attack.
* Recovers by idling.
* Recover rate is slower when guarding.
* If stamina reaches less than zero, the player staggers for a moment making him vulnerable.

#### Poise
The player will have a poise bar which depletes with every taken hit:
* Independent of guarding, every taken hit will deplete the poise bar.
* Recovers by idling.
* Recover rate is slower when guarding.
* If the poise bar reaches zero, the player's animation (can be attacking) is interrupted and he staggers for a moment.
* Better armour have bigger poise bars.

#### Guard
The player can guard with his shield anytime he wants:
* Player moves slower when guarding.
* Every hit taken while guarding will deplete the stamina bar.
* Better shield drain less stamina per hit.

#### Checkpoints
To add dread and difficulty, the player can only save at designated points (like Bonfires):
* If the player dies, he loses his currency, he must go back for them to retrieve it.
* If the player dies 3 times in a row starting on the same checkpoint, the checkpoint is deactivated and the player spawns on the last previous activated checkpoint. Where he needs to traverse the map again to activate the deactivated checkpoint.
    * This can add dread if the player is on it's last life before the checkpoint deactivates, progress is at stake, making the player play more carefully.
    * Needs play testing, it can sound a bit annoying.

#### Misc
* Player can jump, but it's **not a platformer**.
* There may be a jumping attack.
* Tiled base maps, with slopes and ladders.
* Probably have a few enemies than can shoot from a distance.
* The player can side step (dodge) to the left or right to avoid an attack (no invincibility frames) (uses stamina).
* 2 type of attack (for now), light attack (short windup) and heavy attack (long windup).

#### Imposed limitations
* No leveling up.
* Amour and weapons dictates the player's stats.
* One boss per area.

#### 2D limitations
* How to make it so more than 2 enemies can fight you?
    * Being 2D means you can be only attacked from the left or right.

## Estimated progress
* Game engine: 20%
* Story: 20%
* Enemy Design: 0%
* Combat Mechanics: 30%
* Play testing: 0%
* Art: 5%
* Sound: 0%

## TODO
* Think of a fitting currency for the setting.
* Think of the item that can lift the curse.
* Think of a story progression.
* Add depth to combat mechanics.