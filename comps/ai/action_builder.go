package ai

import (
	"game/libs/bump"
	"math/rand"
)

type reaction struct {
	condition func() bool
	states    []WeightedState
}

type ActionBuilder struct {
	timeout     Timeout
	cooldown    Cooldown
	conditions  [](func() bool)
	entry, exit func()
	reacts      []reaction
}

func (ab *ActionBuilder) SetTimeout(timeout Timeout) *ActionBuilder {
	if timeout.Target == "" {
		timeout.Target = CombatState
	}
	ab.timeout = timeout

	return ab
}

func (ab *ActionBuilder) SetCooldown(cooldown Cooldown) *ActionBuilder {
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

func (ab *ActionBuilder) AddReaction(condition func() bool, states []WeightedState) *ActionBuilder {
	ab.reacts = append(ab.reacts, reaction{condition, append(states, WeightedState{CombatState, 0})})

	return ab
}

func (ab *ActionBuilder) Build() *Action {
	return &Action{
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
		Next: func() []WeightedState {
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

func (c *Comp) IdleBuilder(view bump.Rect, viewDist, height float64, nextStates []WeightedState) *ActionBuilder {
	body, anim := c.Actor.GetBody(), c.Actor.GetAnim()
	builder := &ActionBuilder{}
	builder.SetEntry(c.SetSpeedFunc(0, 0))
	if view.W != 0 && view.H != 0 {
		builder.AddReaction(func() bool {
			if targets := body.QueryRect(view); len(targets) > 0 {
				c.Target = targets[0]
			}

			return c.Target != nil
		}, nextStates)
	} else {
		builder.AddReaction(func() bool {
			if targets := body.QueryFront(viewDist, height, anim.FlipX); len(targets) > 0 {
				c.Target = targets[0]
			}

			return c.Target != nil
		}, nextStates)
	}

	return builder
}

func (c *Comp) PursuitBuilder(combatDist, speed, maxSpeed float64, nextStates []WeightedState) *ActionBuilder {
	builder := &ActionBuilder{}
	builder.AddCondition(c.OutRangeFunc(combatDist))
	builder.SetEntry(c.SetSpeedFunc(speed, maxSpeed))
	builder.AddReaction(c.InRangeFunc(combatDist), nextStates)

	return builder
}

func (c *Comp) WaitBuilder(duration, maxDuration float64) *ActionBuilder {
	builder := &ActionBuilder{}
	builder.SetTimeout(Timeout{CombatState, duration, maxDuration})
	builder.SetEntry(c.SetSpeedFunc(0, 0))

	return builder
}

func (c *Comp) PaceBuilder(backUpDist, reactDist, speed, maxSpeed float64, react []WeightedState) *ActionBuilder {
	builder := &ActionBuilder{}
	builder.SetEntry(func() {
		s, ms := speed, maxSpeed
		ms = (ms / 2) * (1 + rand.Float64())
		if c.InTargetRange(0, backUpDist) {
			s *= -1
		}
		c.Actor.SetSpeed(s, ms)
	})
	if len(react) > 0 {
		builder.AddReaction(c.InRangeFunc(reactDist), react)
	}

	return builder
}

func (c *Comp) AnimBuilder(animTag string, nextStates []WeightedState) *ActionBuilder {
	anim := c.Actor.GetAnim()
	builder := &ActionBuilder{}
	builder.AddReaction(func() bool { return anim.State != animTag }, nextStates)

	return builder
}
