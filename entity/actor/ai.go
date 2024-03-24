package actor

import (
	"game/comps/ai"
	"game/old/comps/basic/anim"
)

var (
	minTargetRange  = 10.0
	maxWalkingRange = 20.0

	reactForce = 10.0
	pushForce  = 10.0
)

func WaitAction() *ai.Action {
	return &ai.Action{Name: "Wait", Next: func(_ float64) bool { return false }}
}

func AttackAction(a *Control, tag string, damage float64) *ai.Action {
	return &ai.Action{
		Name:  tag,
		Entry: func() { a.Attack(tag, damage, damage, reactForce, pushForce) },
		Next:  func(_ float64) bool { return a.anim.State != tag },
	}
}

func ApproachAction(a *Control, speed, maxSpeed float64) *ai.Action {
	currentSpeed, currentMaxSpeed := a.speed, a.body.MaxX

	return &ai.Action{
		Name:  "Approach",
		Entry: func() { a.speed, a.body.MaxX = speed, maxSpeed },
		Exit:  func() { a.speed, a.body.MaxX = currentSpeed, currentMaxSpeed },
		Next: func(dt float64) bool {
			if a.ai.Target == nil || a.ai.InTragetRange(minTargetRange, -1) {
				return true
			}
			if !a.ai.InTragetRange(0, maxWalkingRange) {
				// TODO: switch to running, up the speed somehow
			}
			if a.anim.State == anim.WalkTag || a.anim.State == anim.IdleTag {
				x, _ := a.actor.Position()
				tx, _ := a.ai.Target.Position()
				a.anim.FlipX = tx > x
			}
			if !a.PausingState() {
				if a.anim.FlipX {
					a.body.Vx += a.speed * dt
				} else {
					a.body.Vx -= a.speed * dt
				}
			}

			return false
		},
	}
}
