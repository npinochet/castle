package entity

import (
	"fmt"
	"game/comp"
	"game/core"
	"game/utils"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const knightAnimFile = "assets/knight"

func (k *Knight) IsActive() bool        { return k.Active }
func (k *Knight) SetActive(active bool) { k.Active = active }

type Knight struct {
	*Actor
	speed  float64
	player *core.Entity
}

func NewKnight(x, y float64, props map[string]interface{}) *core.Entity {
	speed := 180.0
	anim := &comp.AsepriteComponent{FilesName: knightAnimFile, X: -2, Y: -3}
	body := &comp.BodyComponent{W: 10, H: 11, MaxX: 35}

	knight := &Knight{
		Actor: NewActor(x, y, body, anim, &comp.StatsComponent{MaxPoise: 10}, 20, 20),
		speed: speed,
	}
	// test

	minDist, attackCloseMax, attackMaxDist, maxDist := 20.0, 30.0, 40.0, 50.0
	fms := &comp.AIFsm{Initial: "Ready", State: "Ready"}
	fms.Actions = map[string][]comp.AIAction{
		"Decide": {
			{State: "Waiting", Weight: 0.5},
			{State: "Approaching", Weight: 2, Condition: func(ac *comp.AIComponent) bool {
				return ac.InTargetRange(minDist, -1) && knight.Stats.StaminaPercent() > 0.1
			}}, // maybe add Timeout: 1.0?
			{State: "BackingUp", Weight: 0.5, Condition: func(ac *comp.AIComponent) bool { return ac.InTargetRange(0, maxDist) }},
			{State: "Attack", Weight: 2, Condition: func(ac *comp.AIComponent) bool {
				return ac.InTargetRange(0, attackMaxDist) && knight.Stats.StaminaPercent() > 0.1
			}},
		},
		"Attack": {
			{State: "AttackClose", Condition: func(ac *comp.AIComponent) bool {
				return ac.InTargetRange(0, attackCloseMax) && knight.Stats.StaminaPercent() > 0.1
			}},
			{State: "AttackFast", Condition: func(ac *comp.AIComponent) bool {
				return ac.InTargetRange(0, attackCloseMax) && knight.Stats.StaminaPercent() > 0.1
			}},
		},
	}
	fms.Updates = map[string]comp.AIUpdateFunc{
		"Ready": func(ac *comp.AIComponent) string { return "Decide" },
		"Approaching": func(ac *comp.AIComponent) string {
			// Add timeout
			if knight.AI.InTargetRange(minDist, -1) {
				knight.speed = speed

				return ""
			}

			return "Decide"
		},
		"BackingUp": func(ac *comp.AIComponent) string {
			// Add timeout
			if knight.AI.InTargetRange(0, maxDist) {
				knight.speed = -speed

				return ""
			}

			return "Decide"
		},
		"Waiting": func(ac *comp.AIComponent) string {
			// If timeout, Decide Again
			knight.speed = 0

			return "Decide"
		},
		"AttackClose": func(ac *comp.AIComponent) string {
			knight.Attack()

			return "Attacking"
		},
		"AttackFast": func(ac *comp.AIComponent) string {
			knight.Attack()

			return "Attacking"
		},
		"Attacking": func(ac *comp.AIComponent) string {
			if knight.Anim.State == animAttackTag {
				return ""
			}

			return "Decide"
		},
	}
	knight.Actor.AI = &comp.AIComponent{Fsm: fms}
	knight.AddComponent(knight.AI)
	// end test
	knight.AddComponent(knight)

	return &knight.Entity
}

func (k *Knight) Init(entity *core.Entity) {
	hurtbox, _, _ := k.Anim.GetFrameHitboxes()
	k.Hitbox.PushHitbox(hurtbox.X, hurtbox.Y, hurtbox.W, hurtbox.H, false)
	k.player = k.World.GetEntityByID(utils.PlayerUID)
	k.AI.Target = k.player
}

func (k *Knight) Update(dt float64) {
	k.ManageAnim()
	if k.Anim.State == animWalkTag && !(k.speed != 0) {
		k.Anim.SetState(animIdleTag)
	}
	if k.Anim.State == animWalkTag || k.Anim.State == animIdleTag {
		k.Anim.FlipX = k.player.X < k.X
	}
	if k.Anim.State != animAttackTag && k.Anim.State != animStaggerTag {
		if k.Anim.FlipX {
			k.Body.Vx -= k.speed * dt
		} else {
			k.Body.Vx += k.speed * dt
		}
	}
}

func (k *Knight) DebugDraw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(`State: %s`, k.AI.Fsm.State), 0, 10)
}
