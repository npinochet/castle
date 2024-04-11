package actor

import (
	"game/comps/ai"
	"game/core"
	"game/ext"
	"game/libs/bump"
	"game/old/comps/basic/anim"
)

var (
	minTargetRange = 20.0
	maxTargetRange = 30.0
	frontViewDist  = 60.0

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
				_, _, w, h := a.actor.Rect()
				targets = ext.QueryFront[core.Entity](a.actor, frontViewDist, h, a.anim.FlipX)

				view := &bump.Rect{X: frontViewDist, Y: h, W: frontViewDist, H: h * 2}
				if a.anim.FlipX {
					view.X += frontViewDist + w
				}
				a.ai.DebugRect = view
			} else {
				targets = ext.QueryItems[core.Entity](a.actor, *view, "body")
				a.ai.DebugRect = view
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
	return &ai.Action{Name: "Wait"}
}

func AnimAction(a *Control, tag string, entry func()) *ai.Action {
	return &ai.Action{
		Name:  tag,
		Entry: entry,
		Next:  func(_ float64) bool { return a.anim.State != tag },
	}
}

func AttackAction(a *Control, tag string, damage float64) *ai.Action {
	return &ai.Action{
		Name:  tag,
		Entry: func() { a.Attack(tag, damage, damage, reactForce, pushForce) },
		Next:  func(_ float64) bool { return a.anim.State != tag },
	}
}

func ShieldAction(a *Control) *ai.Action {
	return &ai.Action{
		Name:  "Shield",
		Entry: func() { a.ShieldUp() },
		Exit:  func() { a.ShieldDown() },
	}
}

func ShieldBackUpAction(a *Control, speed, maxSpeed float64) *ai.Action {
	currentMaxSpeed := a.body.MaxX

	return &ai.Action{
		Name: "ShieldBackUp",
		Entry: func() {
			a.body.MaxX = maxSpeed
			a.ShieldUp()
		},
		Exit: func() {
			a.body.MaxX = currentMaxSpeed
			a.ShieldDown()
		},
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

func ApproachAction(a *Control, speed, maxSpeed float64) *ai.Action {
	currentMaxSpeed := a.body.MaxX

	return &ai.Action{
		Name:  "Approach",
		Entry: func() { a.body.MaxX = maxSpeed },
		Exit:  func() { a.body.MaxX = currentMaxSpeed },
		Next: func(dt float64) bool {
			if a.ai.Target == nil || a.ai.InTargetRange(0, minTargetRange) {
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
