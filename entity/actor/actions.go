package actor

import (
	"game/comps/ai"
	"game/core"
	"game/ext"
	"game/libs/bump"
	"game/old/comps/basic/anim"
)

var (
	minTargetRange = 10.0
	maxTargetRange = 20.0
	frontViewDist  = 20.0

	reactForce = 10.0
	pushForce  = 10.0
)

func IdleAction(a *Control, view *bump.Rect) *ai.Action {
	return &ai.Action{
		Name: "Idle",
		Next: func(_ float64) bool {
			if a.ai.Target != nil {
				return true
			}
			var targets []core.Entity
			if view == nil {
				_, _, _, h := a.actor.Rect()
				targets = ext.QueryFront[core.Entity](a.actor, frontViewDist, h, a.anim.FlipX)
			} else {
				targets = ext.QueryItems[core.Entity](a.actor, *view, "body")
			}
			if len(targets) > 0 {
				a.ai.Target = targets[0]

				return true
			}

			return false
		},
	}
}

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
	currentMaxSpeed := a.body.MaxX

	return &ai.Action{
		Name:  "Approach",
		Entry: func() { a.body.MaxX = maxSpeed },
		Exit:  func() { a.body.MaxX = currentMaxSpeed },
		Next: func(dt float64) bool {
			if a.ai.Target == nil || a.ai.InTargetRange(minTargetRange, -1) {
				return true
			}
			if !a.ai.InTargetRange(0, maxTargetRange) {
				// TODO: switch to running, up the speed somehow
			}
			if a.anim.State == anim.WalkTag || a.anim.State == anim.IdleTag {
				x, _ := a.actor.Position()
				tx, _ := a.ai.Target.Position()
				a.anim.FlipX = tx > x
			}
			if !a.PausingState() {
				if a.anim.FlipX {
					a.body.Vx += speed * dt
				} else {
					a.body.Vx -= speed * dt
				}
			}

			return false
		},
	}
}

func BackUpAction(a *Control, speed, maxSpeed float64) *ai.Action {
	currentMaxSpeed := a.body.MaxX

	return &ai.Action{
		Name:  "BackUp",
		Entry: func() { a.body.MaxX = maxSpeed },
		Exit:  func() { a.body.MaxX = currentMaxSpeed },
		Next: func(dt float64) bool {
			if a.ai.Target == nil || !a.ai.InTargetRange(0, maxTargetRange) {
				return true
			}
			if a.anim.State == anim.WalkTag || a.anim.State == anim.IdleTag {
				x, _ := a.actor.Position()
				tx, _ := a.ai.Target.Position()
				a.anim.FlipX = tx > x
			}
			if !a.PausingState() {
				if a.anim.FlipX {
					a.body.Vx -= speed * dt
				} else {
					a.body.Vx += speed * dt
				}
			}

			return false
		},
	}
}
