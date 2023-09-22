package actor

import (
	"game/libs/bump"
	"math/rand"
)

type reaction struct {
	condition func() bool
	states    []AIWeightedState
}

type ActionBuilder struct {
	timeout     AITimeout
	cooldown    AICooldown
	conditions  [](func() bool)
	entry, exit func()
	reacts      []reaction
}

func (ab *ActionBuilder) SetTimeout(timeout AITimeout) *ActionBuilder {
	if timeout.Target == "" {
		timeout.Target = AICombatState
	}
	ab.timeout = timeout

	return ab
}

func (ab *ActionBuilder) SetCooldown(cooldown AICooldown) *ActionBuilder {
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

func (ab *ActionBuilder) AddReaction(condition func() bool, states []AIWeightedState) *ActionBuilder {
	ab.reacts = append(ab.reacts, reaction{condition, append(states, AIWeightedState{AICombatState, 0})})

	return ab
}

func (ab *ActionBuilder) Build() *AIAction {
	return &AIAction{
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
		Next: func() []AIWeightedState {
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

func (c *AI) IdleBuilder(view bump.Rect, viewDist, height float64, nextStates []AIWeightedState) *ActionBuilder {
	body, anim := c.Actor.Body, c.Actor.Anim
	builder := &ActionBuilder{}
	builder.SetEntry(c.SetSpeedFunc(0, 0))
	if view.W != 0 && view.H != 0 {
		builder.AddReaction(func() bool {
			if targets := body.QueryActors(c.Actor, view, true); len(targets) > 0 {
				c.Target = targets[0]
			}

			return c.Target != nil
		}, nextStates)
	} else {
		builder.AddReaction(func() bool {
			if targets := body.QueryFront(c.Actor, viewDist, height, anim.FlipX, true); len(targets) > 0 {
				c.Target = targets[0]
			}

			return c.Target != nil
		}, nextStates)
	}

	return builder
}

func (c *AI) PursuitBuilder(combatDist, speed, maxSpeed float64, nextStates []AIWeightedState) *ActionBuilder {
	builder := &ActionBuilder{}
	builder.AddCondition(c.OutRangeFunc(combatDist))
	builder.SetEntry(c.SetSpeedFunc(speed, maxSpeed))
	builder.AddReaction(c.InRangeFunc(combatDist), nextStates)

	return builder
}

func (c *AI) WaitBuilder(duration, maxDuration float64) *ActionBuilder {
	builder := &ActionBuilder{}
	builder.SetTimeout(AITimeout{AICombatState, duration, maxDuration})
	builder.SetEntry(c.SetSpeedFunc(0, 0))

	return builder
}

func (c *AI) PaceBuilder(backUpDist, reactDist, speed, maxSpeed float64, react []AIWeightedState) *ActionBuilder {
	builder := &ActionBuilder{}
	builder.SetEntry(func() {
		s, ms := speed, maxSpeed
		ms = (ms / 2) * (1 + rand.Float64())
		if c.InTargetRange(0, backUpDist) {
			s *= -1
			ms /= 2
		}
		c.SetSpeedFunc(s, ms)()
	})
	if len(react) > 0 {
		builder.AddReaction(c.InRangeFunc(reactDist), react)
	}

	return builder
}

func (c *AI) AnimBuilder(animTag string, nextStates []AIWeightedState) *ActionBuilder {
	anim := c.Actor.Anim
	builder := &ActionBuilder{}
	builder.AddReaction(func() bool { return anim.State != animTag }, nextStates)

	return builder
}
