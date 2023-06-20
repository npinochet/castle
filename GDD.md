# Game Design Document

**Working title**: Mini souls: Eternal Castle whatever

## Inspirations

* Dark Souls (Maybe a little too much).
* "El inmortal" by Jorge Luis Borges.

## Setting

### Setting introduction

A curse has been casted upon the the king's castle.
From one day to the next a barrier surrounding the whole castle appeared, keeping everyone trapped inside and, by an unknown reason, unable to die.

Many years or centuries have passed and the barrier is still intact. Everyone is in the brink of insanity after being confined together for so long without contact with the outer world. Little by little the inhabitants start losing their humanity, acting on their most primal impulses, everyone is old, hurting, tired, hungry and cold.

The player finally wakes up and finds himself inside the castle after been [asleep/dead/unconscious]? for many years. He has in his possession the only item that can truly lift the curse and bring death to a someone, but it only has a one time use and he's saving it for someone special. Most inhabitants you encounter will do anything to obtain your item.

After year or centuries of being prisoners on the castle, strange social bonds and groups started to emerge, a whole new ecosystem has been developed as people try their best to keep their little sanity left.

### Objective

* Make the world feel alive and believable, not "gamey".
* Make a small map, quality over quantity.

### Themes

* Time
* Dark Fantasy

### Nomenclature

* Without the first sin: Immortal: Restless
* Currency: Specks ~~of time~~?
* Rite of immortality: ???
* Curse of immortality: ???
* Item/Rite that can reverse the immortality: ???
* All knowing god who bringed the rite: The dreamer

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
* Kitchen.

### Possible area themes

* A dungeon full of abandoned deadly attractions, where people paid (a greedy man) for a *supposed* certain death.
* A garden/jungle: As the years passed, plants can grow limitless.
* A Library housing mindless intellectuals who worship books.

### Possible themes

* Time
  * The immortality rite is performed by freeing one's first sin, and manipulate time (freeze it). What you collect is the game (souls) are "time crystals". Which let you level up (?).

### Loose ideas

* There are humans that managed to free themselves of the first sin, though a rite, which grant them immortality, these humans became gods-like in the kingdom.
* The king (Arawn) of the castle is one of these "gods". In his governance some major event happened where the wisdom of becoming immortal leaked, making everyone in the castle immortal. That's when an outside god placed a barrier on the castle and let all it's inhabitants prisoned.
* There was a person who didn't use the wisdom to became immortal, which let him die and later became like a saint and prophet for one of the covenants which named him "the first death". This covenant seeks a way to undo the rite somehow and become sinners again. Maybe as a plot twist, this prophet did become immortal and he just faked his dead and appears in the game?

#### Setting imposed limitations

* Must have at most 3 areas, presenting one covenant/group per area.
* At least 3 enemies types per area. (3x3 = 9 enemies).
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
  * One aspect to achieve this is raising the game's difficulty (hopefully in terms of skills and not numbers), make sure the player dies a lot (not a primary objective/idea).
    * Caveats: **Don't make health sponge enemies**. Make them hit harder if I must do.
  * Can also punish the player for dying, with lost progress or other mechanics.
* Have a semi-maze like map, where you have to be aware of your surrounding to navigate it.
* To a lesser extent have, have a great emphasis on enemy AI, this will add complexity to decision making and difficulty.
* **Do not make a CastleVania: No item gates, minimal backtracking**.

### Main game mechanics

### Mechanics introduction

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
* Better armor have bigger poise bars.

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
* 2 type of attack (for now), light attack (short windup) and heavy attack (long windup), by holding and charging the attack button.

#### Mechanics imposed limitations

* No leveling up.
* Amour and weapons dictates the player's stats.
* One boss per area.

#### 2D limitations

* How to make it so more than 2 enemies can fight you?
  * Being 2D means you can be only attacked from the left or right.

## Enemies

### Firs area

#### Crawler

The first enemy, an insignificant, slow and weak obstacle, they can occasionally attack.
A good introduction for combat.
Lore:
There are poor lost souls that are too tired and spent from living.

#### Ghoul

These are well-rounded enemies that can throw rocks from higher places. They have a 2 attack combo that can trick you, but they have weak poise.
Lore:
IDK

#### Skeleman

These ones are the first real challenge a player will face. They have mid poise and can spin their swords to make a wall of hitboxes. They will force you to use the shield.
Lore:
IDK

#### Abomination

These are health walls (not too much). But can jump and fall on you. They hit hard but are slow and well telegraphed.
Lore:
They come from the mass of bodies under the dungeon. Where people believe that pulping and mutilation oneself can lead to a state of death. These are the ones that decide to leave the mass.

## Estimated progress

* Game engine: 20%
* Story: 20%
* Enemy Design: 0%
* Combat Mechanics: 30%
* Play testing: 0%
* Art: 5%
* Sound: 0%

## TODO

* Think of a story progression/objectives.
  * This can be left for later.
* Add depth to combat mechanics.
  * dash? Backoff?
  * Hold attack to do a heavy attack.
* Consider input/button weight to player controls. (what is this??) (really what is this? is this a masahiro thing?)
* Healing mechanic and limited use.
* Bonfires/Gravestones.
  * Every save point (gravestone) has an engraving which gives a little lore to the area.
* First area themes.
  * A dungeon repurposed as a cult area with followers who believe that beheading and cracking of the skull is the new death. Where you can't know if one gets to live still or really die.
  * Lead by the final boss who's a greedy leader with.
    * Named Lakim (greed in black speech), he was
* Art.
* Have an enemy which is made from diferent bodyparts stiched together, like frankenstein's monster.
* Have an enemy which is are aristocrats with deer of some other animal heads. A type fo grafted enemy.
  
Level design tips that might be useful to keep in mind:

* search castle layouts.
* keep it random and unpredictable.
* New londo ruins can be accessed through the beginning, but has unkillable enemies to prevent new players going through it.
* you can still tackle difficult areas early to get cool items.
* difficulty curves, no more than 2-3 bosses that needs to be killed in any order.

TODO list:

* Gameplay:
  * Tweak enemy behavior:
    * Enemies are too aggressive, they don't have openings, every dark souls enemy has
    * If player comes closer maybe the shield is raise? or the enemy attacks (can be tweak with aggressive parameter?)
  * Make simple test map with multiple enemies, see how they behave
  * Think about adding roll/dodge for gameplay experimentation
* Future:
  * Improve AI implementation, it it's too complex and relies too much on copy and paste
  * Bonfire
  * Use Tiled auto-mapping for backgrounds and floor tiles
    * And maybe, if it's not to difficult, aut mapping for collision objects
  * Polish combat juice, stop time when take damage (or send damage), kill animations, etc...
    * [Stop for Big Moments! [Design Specifics]](https://www.youtube.com/watch?v=OdVkEOzdCPw)
  * Add path-finding for enemies
  * Implement specks for defeating enemies
  * Think of a mechanic to encourage player push forwards the next bonfire with loads of XP, instead of backtracking to previous bonfire and cash in.
  * The shield can do poise damage if block correctly (like a parry).
  * When an enemy is hit, but does not breaks poise, it should slow down their animation for a little. To really sell the hit and make the fight more dynamic.
  * https://www.gamedeveloper.com/design/the-fundamental-pillars-of-a-combat-system

* Level Design:
  * Test when possible

Level progression:

* Start, tutorial.
* Find first bonfire. Maybe an NPC?
* Descend through dungeon.
* Find chest with tougher enemy?
* Encounter mass of body enemies.
* Encounter the source of the enemies, a big pile of body parts.
* Descend further through the dungeon.

Level design steps [How to Design Great Metroidvania Levels | Game Design](https://www.youtube.com/watch?v=bAHXYfP38CA):

* Draft the map ()
* Develop the timeline
* Develop the abilities
* Map out each room
* Test, Review, Adapt

## NOTHING WORKS

### What feels wrong

* You can overcome an enemy really easy
* Combat feels boring
  * Everything is stiff [and slow?]
  * Feels like a waiting game
  * Stamina management feels more as a inconvenience than an actual mechanic that raises stakes
* It's not difficult
  * Enemies aren't really a threat
  * Have no idea how Castlevania fixes this
* I've loose the main objective focus, It doesn't really make you feel nervous or in danger

### Actionables

* Add a quick step/dodge mechanic where you can approach or flee the enemy. Enemies can do this too, so you have to be focused and block or dodge incoming quick step attacks.
* Make a tall enemy that can't be jumped over
* Find a way to make enemies follow you and really overwhelm and bother you when you pass them through
  * For example in dark souls you can't really run through enemies some times, because they can stop you in narrow doors or geometry and overwhelm you
* Polish combat? Rethink combat, maybe change the animations a bit, make them faster?
  * Rethink anticipation frames
* Have stakes
  * You can't really care about your HP if there is nothing to loose
* Rethink stamina and blocking mechanic? Think of a way to limit player spamming attack button or hide behind shield
* I've believe the game is difficult enough, but it need to be more dynamic, like teams of enemy working together, not a hallway with enemy 1v1 the whole way
  * Still, make sure to make them die again and again

### Key elements missing from game which Dark Souls has

In MDA context, Dynamics:

* Positioning: The player must position themselves correctly in relation to the enemy to avoid attacks and exploit weaknesses.
* Adaptation: The player must adapt to each enemy's attack patterns, weaknesses, and strengths to overcome them.
