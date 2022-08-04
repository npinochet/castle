package comp

import (
	"game/core"
	"game/utils"
	"math/rand"
)

func (ac *AIComponent) IsActive() bool        { return ac.active }
func (ac *AIComponent) SetActive(active bool) { ac.active = active }

type AIAction struct {
	State     string
	Condition func(ac *AIComponent) bool
	Weight    float64
	// Timeout time.Duration // TODO: Add timeout for ocasions like "backingUp" or "Waiting"
}

type AIUpdateFunc func(ac *AIComponent) string

type AIFsm struct {
	Initial, State string
	Updates        map[string]AIUpdateFunc
	Actions        map[string][]AIAction
	Component      *AIComponent
}

func (ac *AIFsm) Update(dt float64) {
	if state := ac.State; state == "" {
		ac.State = "Ready"
	}
	if update := ac.Updates[ac.State]; update != nil {
		if actionOrState := update(ac.Component); actionOrState != "" {
			for ac.Updates[actionOrState] == nil && actionOrState != "" {
				actionOrState = ac.StateFromAction(actionOrState)
			}
			ac.State = actionOrState
		}
	}
}

func (ac *AIFsm) StateFromAction(action string) string {
	actions := ac.Actions[action]
	if len(actions) == 1 {
		return actions[0].State
	}

	type StateProb struct {
		State  string
		Weight float64
	}
	var states []StateProb

	totalWeight := 0.0
	for _, t := range actions {
		if t.Condition == nil || t.Condition(ac.Component) {
			weight := t.Weight
			if weight == 0 {
				weight = 1.0 / float64(len(actions))
			}
			totalWeight += weight
			states = append(states, StateProb{t.State, weight})
		}
	}

	r := rand.Float64() * totalWeight
	for _, nextState := range states {
		if r -= nextState.Weight; r <= 0 {
			return nextState.State
		}
	}

	return ""
}

type AIComponent struct {
	active         bool
	Fsm            *AIFsm
	Target, entity *core.Entity
}

func (ac *AIComponent) Init(entity *core.Entity) {
	ac.entity = entity
	if ac.Fsm != nil { // What happens here?
		ac.Fsm.Component = ac
	}
}

func (ac *AIComponent) Update(dt float64) {
	ac.Fsm.Update(dt)
}

func (ac *AIComponent) distFromTarget() float64 {
	if ac.Target == nil {
		return -1
	}
	tx, ty := ac.Target.Position()
	x, y := ac.entity.Position()

	return utils.Distante(x, y, tx, ty)
}

func (ac *AIComponent) InTargetRange(minDist, maxDist float64) bool {
	dist := ac.distFromTarget()
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
