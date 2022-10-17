package ai

import (
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

type Action struct {
	Timeout     Timeout
	Condition   func() bool
	Next        func() []WeightedState
	Entry, Exit func()
	// TODO: recovery time? or just depend on stamina bars?
}

type Fsm struct {
	Comp           *Comp
	Actions        map[State]*Action
	State, Initial State
	timer          float64
	timeoutTarget  State
}

type actionIndex struct {
	action *Action
	index  int
}

func (f *Fsm) update(dt float64) {
	if f.State == "" {
		f.State = f.Initial
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
}

func (f *Fsm) setState(states []WeightedState) {
	if action := f.Actions[f.State]; action != nil && action.Exit != nil {
		action.Exit()
	}
	f.timeoutTarget = ""
	f.State = f.selectState(states)
	if action := f.Actions[f.State]; action != nil && action.Entry != nil {
		action.Entry()
	}
}

func (f *Fsm) selectState(states []WeightedState) State {
	actions := make([]*Action, len(states))
	for i, s := range states {
		action := f.Actions[s.State]
		if action == nil {
			log.Panicf("AI: no action found for state %s\n", s.State)
		}
		if s.Weight == 0 {
			s.Weight = 1.0 / float64(len(states))
		}
		actions[i] = action
	}

	var selected []actionIndex
	totalWeight := 0.0
	for i, a := range actions {
		if a.Condition == nil || a.Condition() {
			totalWeight += states[i].Weight
			selected = append(selected, actionIndex{a, i})
		}
	}

	r := rand.Float64() * totalWeight
	for _, a := range selected {
		if r -= states[a.index].Weight; r <= 0 {
			action := a.action
			if action.Timeout.Duration > 0 {
				f.timer = action.Timeout.Duration
				if action.Timeout.MaxDuration > action.Timeout.Duration {
					f.timer += rand.Float64() * (action.Timeout.MaxDuration - action.Timeout.Duration)
				}
				f.timeoutTarget = action.Timeout.Target
			}

			return states[a.index].State
		}
	}

	return ""
}
