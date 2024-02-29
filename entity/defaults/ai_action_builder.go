package defaults

import (
	"game/comps/ai"
	"game/libs/bump"
	"game/utils"
	"math/rand"
)

type reaction struct {
	condition func() bool
	states    []ai.WeightedState
}

type ActionBuilder struct {
	timeout     ai.Timeout
	cooldown    ai.Cooldown
	conditions  [](func() bool)
	entry, exit func()
	reacts      []reaction
}

func (ab *ActionBuilder) SetTimeout(timeout ai.Timeout) *ActionBuilder {
	if timeout.Target == "" {
		timeout.Target = ai.CombatState
	}
	ab.timeout = timeout

	return ab
}

func (ab *ActionBuilder) SetCooldown(cooldown ai.Cooldown) *ActionBuilder {
	ab.cooldown = cooldown

	return ab
}

func (ab *ActionBuilder) AddCondition(condition func() bool) *ActionBuilder {
	ab.conditions = append(ab.conditions, condition)

	return ab
}

func (ab *ActionBuilder) SetEntry(entry func()) *ActionBuilder {
	ab.entry = entry

	return ab
}

func (ab *ActionBuilder) SetExit(exit func()) *ActionBuilder {
	ab.exit = exit

	return ab
}

func (ab *ActionBuilder) AddReaction(condition func() bool, states []ai.WeightedState) *ActionBuilder {
	ab.reacts = append(ab.reacts, reaction{condition, append(states, ai.WeightedState{ai.CombatState, 0})})

	return ab
}

func (ab *ActionBuilder) Build() *ai.Action {
	return &ai.Action{
		Timeout: ab.timeout, Cooldown: ab.cooldown,
		Entry: ab.entry, Exit: ab.exit,
		Condition: func() bool {
			for _, c := range ab.conditions {
				if !c() {
					return false
				}
			}

			return true
		},
		Next: func() []ai.WeightedState {
			for _, r := range ab.reacts {
				if r.condition() {
					return r.states
				}
			}

			return nil
		},
	}
}

// Preset Actions.

func (a *Actor) IdleBuilder(view bump.Rect, viewDist, height float64, nextStates []ai.WeightedState) *ActionBuilder {
	builder := &ActionBuilder{}
	builder.SetEntry(a.SetSpeedFunc(0, 0))
	if view.W != 0 && view.H != 0 {
		builder.AddReaction(func() bool {
			if targets := a.Body.QueryEntites(view, true); len(targets) > 0 {
				a.AI.Target = targets[0]
			}

			return a.AI.Target != nil
		}, nextStates)
	} else {
		builder.AddReaction(func() bool {
			if targets := a.Body.QueryFront(viewDist, height, a.Anim.FlipX, true); len(targets) > 0 {
				a.AI.Target = targets[0]
			}

			return a.AI.Target != nil
		}, nextStates)
	}

	return builder
}

func (a *Actor) PursuitBuilder(combatDist, speed, maxSpeed float64, nextStates []ai.WeightedState) *ActionBuilder {
	builder := &ActionBuilder{}
	builder.AddCondition(a.OutRangeFunc(combatDist))
	builder.SetEntry(a.SetSpeedFunc(speed, maxSpeed))
	builder.AddReaction(a.InRangeFunc(combatDist), nextStates)

	return builder
}

func (a *Actor) WaitBuilder(duration, maxDuration float64) *ActionBuilder {
	builder := &ActionBuilder{}
	builder.SetTimeout(ai.Timeout{ai.CombatState, duration, maxDuration})
	builder.SetEntry(a.SetSpeedFunc(0, 0))

	return builder
}

func (a *Actor) PaceBuilder(backUpDist, reactDist, speed, maxSpeed float64, react []ai.WeightedState) *ActionBuilder {
	builder := &ActionBuilder{}
	builder.SetEntry(func() {
		s, ms := speed, maxSpeed
		ms = (ms / 2) * (1 + rand.Float64())
		if a.InTargetRange(0, backUpDist) {
			s *= -1
			ms /= 2
		}
		a.SetSpeedFunc(s, ms)()
	})
	if len(react) > 0 {
		builder.AddReaction(a.InRangeFunc(reactDist), react)
	}

	return builder
}

func (a *Actor) AnimBuilder(animTag string, nextStates []ai.WeightedState) *ActionBuilder {
	builder := &ActionBuilder{}
	builder.AddReaction(func() bool { return a.Anim.State != animTag }, nextStates)

	return builder
}

// Utils.

func (a *Actor) InTargetRange(minDist, maxDist float64) bool {
	dist := a.distFromTarget()
	if dist < 0 {
		return false
	}
	in := dist >= minDist
	out := true
	if maxDist > 0 {
		out = dist <= maxDist
	}

	return in && out
}

func (a *Actor) InRangeFunc(maxDist float64) func() bool {
	return func() bool { return a.InTargetRange(0, maxDist) }
}

func (a *Actor) OutRangeFunc(maxDist float64) func() bool {
	return func() bool { return !a.InTargetRange(0, maxDist) }
}

func (a *Actor) SetSpeedFunc(speed, maxSpeed float64) func() {
	return func() { a.Control.SetSpeed(speed, maxSpeed) }
}

func (a *Actor) distFromTarget() float64 {
	if a.AI.Target == nil {
		return -1
	}
	tx, ty := a.AI.Target.Position()
	x, y := a.Entity.Position()

	return utils.Distante(x, y, tx, ty)
}
