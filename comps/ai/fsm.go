package ai

import (
	"fmt"
	"log"
	"math/rand"
)

type State string

type WeightedState struct {
	State
	Weight float64
}

type Timeout struct {
	Target                State
	Duration, MaxDuration float64
}

type Cooldown struct {
	Duration, MaxDuration float64
}

type Action struct {
	Timeout     Timeout
	Cooldown    Cooldown
	Condition   func() bool
	Next        func() []WeightedState
	Entry, Exit func()
}

type Fsm struct {
	Comp           *Comp
	Actions        map[State]*Action
	State, Initial State
	timer          float64
	timeoutTarget  State
	cooldowns      map[State]float64
}

func (f *Fsm) update(dt float64) {
	if f.State == "" {
		f.State = f.Initial
		if action := f.Actions[f.State]; action != nil {
			if action.Entry != nil {
				action.Entry()
			}
		}
	}
	if action := f.Actions[f.State]; action != nil && action.Next != nil {
		if actions := action.Next(); actions != nil {
			f.setState(actions)
		}
	}
	if f.timeoutTarget != "" {
		f.timer -= dt
		if f.timer <= 0 {
			f.setState([]WeightedState{{f.timeoutTarget, 0}})
			f.timeoutTarget = ""
		}
	}
	if f.cooldowns == nil {
		f.cooldowns = map[State]float64{}
	}
	for state, timer := range f.cooldowns {
		f.cooldowns[state] -= dt
		if timer <= 0 {
			delete(f.cooldowns, state)
		}
	}
}

func (f *Fsm) setState(states []WeightedState) {
	if action := f.Actions[f.State]; action != nil {
		if action.Cooldown.Duration > 0 {
			f.cooldowns[f.State] = action.Cooldown.Duration
			if action.Cooldown.MaxDuration > action.Cooldown.Duration {
				f.cooldowns[f.State] += rand.Float64() * (action.Cooldown.MaxDuration - action.Cooldown.Duration)
			}
		}
		if action.Exit != nil {
			action.Exit()
		}
	}
	f.State = f.selectState(states)
	f.timeoutTarget = ""
	if action := f.Actions[f.State]; action != nil {
		if action.Timeout.Duration > 0 {
			f.timer = action.Timeout.Duration
			if action.Timeout.MaxDuration > action.Timeout.Duration {
				f.timer += rand.Float64() * (action.Timeout.MaxDuration - action.Timeout.Duration)
			}
			f.timeoutTarget = action.Timeout.Target
		}
		if action.Entry != nil {
			action.Entry()
		}
	}
}

func (f *Fsm) selectState(states []WeightedState) State {
	actions := make([]*Action, len(states))
	for i, s := range states {
		action := f.Actions[s.State]
		if action == nil {
			log.Panicf("AI: no action found for state %s\n", s.State)
		}
		if s.Weight < 0 {
			s.Weight = 1.0 / float64(len(states))
		}
		actions[i] = action
	}

	var selected []int
	totalWeight := 0.0
	for i, a := range actions {
		if (a.Condition == nil || a.Condition()) && f.cooldowns[states[i].State] <= 0 {
			totalWeight += states[i].Weight
			selected = append(selected, i)
		}
	}

	r := rand.Float64() * totalWeight
	for _, i := range selected {
		if r -= states[i].Weight; r <= 0 {
			return states[i].State
		}
	}

	if len(selected) == 0 && len(states) > 0 {
		fmt.Println(f.State, states) // TODO: Prevent this! Dumb cooldowns, maybe behaviour trees really are the way.
	}

	return ""
}
