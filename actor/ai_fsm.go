package actor

import (
	"log"
	"math/rand"
)

type AIState string

type AIWeightedState struct {
	State  AIState
	Weight float64
}

type AITimeout struct {
	Target                AIState
	Duration, MaxDuration float64
}

type AICooldown struct {
	Duration, MaxDuration float64
}

type AIAction struct {
	Timeout     AITimeout
	Cooldown    AICooldown
	Condition   func() bool
	Next        func() []AIWeightedState
	Entry, Exit func()
}

type AIFsm struct {
	Actions        map[AIState]*AIAction
	State, Initial AIState
	timer          float64
	timeoutTarget  AIState
	cooldowns      map[AIState]float64
}

func NewAIFsm(initial AIState) *AIFsm {
	return &AIFsm{Initial: initial, Actions: map[AIState]*AIAction{}}
}

func (f *AIFsm) SetAction(state AIState, action *AIAction) *AIFsm {
	f.Actions[state] = action

	return f
}

func (f *AIFsm) update(dt float64) {
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
			f.setState([]AIWeightedState{{f.timeoutTarget, 1}})
		}
	}
	if f.cooldowns == nil {
		f.cooldowns = map[AIState]float64{}
	}
	for state, timer := range f.cooldowns {
		f.cooldowns[state] -= dt
		if timer <= 0 {
			delete(f.cooldowns, state)
		}
	}
}

func (f *AIFsm) setState(states []AIWeightedState) {
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
	nextState := f.selectState(states)
	if nextState == "" {
		return
	}
	f.State = nextState
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

func (f *AIFsm) selectState(states []AIWeightedState) AIState {
	actions := make([]*AIAction, len(states))
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

	return ""
}
