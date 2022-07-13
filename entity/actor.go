package entity

import (
	"game/comp"
	"game/core"
	"game/libs/bump"
)

const defaultForce = 1800

func (a *Actor) IsActive() bool        { return a.Active }
func (a *Actor) SetActive(active bool) { a.Active = active }

type Actor struct {
	core.Entity
	body       *comp.BodyComponent
	hitbox     *comp.HitboxComponent
	anim       *comp.AsepriteComponent
	stats      *comp.StatsComponent
	reactForce float64
}

func NewActor(x, y float64, body *comp.BodyComponent, anim *comp.AsepriteComponent, stats *comp.StatsComponent) *Actor {
	if stats == nil {
		stats = &comp.StatsComponent{}
	}

	actor := &Actor{
		Entity: core.Entity{X: x, Y: y},
		hitbox: &comp.HitboxComponent{},
		body:   body, anim: anim, stats: stats,
		reactForce: defaultForce,
	}
	actor.AddComponent(actor.body, actor.hitbox, actor.anim, actor.stats)
	actor.hitbox.HurtFunc = func(otherHc *comp.HitboxComponent, col bump.Collision, damange float64) {
		actor.Hurt(*otherHc.EntX, damange, nil)
	}
	actor.hitbox.BlockFunc = func(otherHc *comp.HitboxComponent, col bump.Collision, damange float64) {
		actor.Block(*otherHc.EntX, damange, nil)
	}

	return actor
}

func (a *Actor) ManageAnim(idle, walk, attack, stagger string) {
	state := a.anim.State
	a.body.Friction = !(state == walk && a.body.Vx != 0)
	a.stats.SetActive(state != attack && state != stagger)

	if state == idle || state == walk {
		nextState := idle
		if a.body.Vx != 0 {
			nextState = walk
		}
		a.anim.SetState(nextState)
	}
}

func (a *Actor) Attack(state string, stamina, damage float64) {
	if a.stats.Stamina <= 0 {
		return
	}
	force := a.reactForce
	if a.anim.FlipX {
		force *= -1
	}
	a.anim.SetState(state)

	once := false
	a.anim.OnFrames(func(frame int) {
		if _, hitbox, _ := a.anim.GetFrameHitboxes(); hitbox != nil {
			if !once {
				once = true
				a.body.Vx += force * 1.0 / 60
				a.stats.AddStamina(-stamina)
			}
			if a.hitbox.Hit(hitbox.X, hitbox.Y, hitbox.W, hitbox.H, damage) {
				// p.Stagger(dt, force) when shield has too much defense?
				a.body.Vx -= (force / 2) * 1.0 / 60
			}
		}
	})
}

func (a *Actor) Stagger(state string, force float64) {
	if a.anim.State == state {
		return
	}
	a.anim.SetState(state)
	a.body.Vx = -force
}

func (a *Actor) Hurt(otherX float64, damage float64, poiseBreak func(force float64)) {
	force := a.reactForce
	if a.X > otherX {
		force *= -1
	}
	a.body.Vx -= (force / 2) * 1.0 / 60
	a.stats.AddPoise(-damage)
	a.stats.AddHealth(-damage)
	if a.stats.Poise <= 0 && poiseBreak != nil {
		poiseBreak(force)
	}
}

func (a *Actor) Block(otherX float64, damage float64, blockBreak func(force float64)) {
	force := a.reactForce
	if a.X > otherX {
		force *= -1
	}
	a.body.Vx -= (force / 2) * 1.0 / 60
	a.stats.AddStamina(-damage)
	if a.stats.Stamina < 0 && blockBreak != nil {
		blockBreak(force)
	}
}
