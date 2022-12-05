package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"math/rand"
)

type AIConfig struct {
	viewDist         float64
	combatDist       float64
	backUpDist       float64
	reactDist        float64
	minAttackStamina float64
}

// nolint: gomnd, nolintlint
func defaultConfig() *AIConfig {
	// TODO: Review config attributes and make them more understandable.
	return &AIConfig{
		viewDist:         100,
		combatDist:       100,
		backUpDist:       40,
		reactDist:        20,
		minAttackStamina: 0.2,
	}
}

// TODO: Review states, sometimes the state is empty and falls back to idle.
func (a *Actor) NewDefaultAI(config *AIConfig) *ai.Comp {
	if config == nil {
		config = defaultConfig()
	}
	maxX := a.Body.MaxX
	speed := a.speed
	act := []ai.WeightedState{{"Wait", 2}, {"Pace", 0.5}, {"Attack", 1}, {"RunAttack", 1}, {"Guard", 0.1}}
	react := []ai.WeightedState{{"Attack", 1}, {"Guard", 0.5}, {"Wait", 0.8}}

	actions := map[ai.State]*ai.Action{
		"Act": {Next: func() []ai.WeightedState { return act }},
		"Idle": {
			Entry: func() { a.speed = 0 },
			Next: func() []ai.WeightedState {
				if targets := a.Body.QueryFront(config.viewDist, 40, a.Anim.FlipX); len(targets) > 0 {
					a.AI.Target = targets[0]
				}
				if a.AI.Target != nil {
					return []ai.WeightedState{{"Pursuit", 1}, {"Pace", 0}}
				}

				return nil
			},
		},
		"Wait": {
			Cooldown: ai.Cooldown{1, 2},
			Timeout:  ai.Timeout{"Act", 1, 1.2},
			Entry:    func() { a.speed = 0 },
		},
		"Pursuit": {
			Condition: func() bool { return !a.AI.InTargetRange(0, config.combatDist) },
			Entry: func() {
				a.speed = speed
				a.Body.MaxX = maxX
			},
			Next: func() []ai.WeightedState {
				if a.AI.InTargetRange(0, config.combatDist) {
					return []ai.WeightedState{{"Pace", 0}}
				}

				return nil
			},
		},
		"Pace": {
			Cooldown:  ai.Cooldown{1, 2},
			Timeout:   ai.Timeout{"Act", 1, 2},
			Condition: func() bool { return a.AI.InTargetRange(0, config.combatDist) },
			Entry: func() {
				div := 2 + rand.Float64()*1
				a.Body.MaxX = maxX / div
				a.speed = speed
				if a.AI.InTargetRange(0, config.backUpDist) {
					a.speed *= -1
				}
			},
			Next: func() []ai.WeightedState {
				if !a.AI.InTargetRange(0, config.combatDist) {
					return []ai.WeightedState{{"Pursuit", 1}}
				}
				if a.AI.InTargetRange(0, config.reactDist) {
					return react
				}

				return nil
			},
		},
		"Attack": {
			Cooldown:  ai.Cooldown{2, 3},
			Condition: func() bool { return a.Stats.Stamina > config.minAttackStamina },
			Entry:     func() { a.Attack() },
			Next: func() []ai.WeightedState {
				if a.Anim.State != anim.AttackTag {
					return []ai.WeightedState{{"Pace", 0}}
				}

				return nil
			},
		},
		"RunAttack": {
			Cooldown: ai.Cooldown{2, 3},
			Condition: func() bool {
				return a.Stats.Stamina > config.minAttackStamina && !a.AI.InTargetRange(0, config.reactDist)
			},
			Entry: func() {
				a.speed = speed
				a.Body.MaxX = maxX
			},
			Next: func() []ai.WeightedState {
				if a.AI.InTargetRange(0, config.reactDist) {
					return []ai.WeightedState{{"Attack", 0}}
				}

				return nil
			},
		},
		"Guard": {
			Cooldown: ai.Cooldown{3, 0},
			Timeout:  ai.Timeout{"Pace", 1, 2},
			Entry: func() {
				a.speed = -speed
				a.Body.MaxX = maxX / 4
				a.ShieldUp()
			},
			Exit: func() { a.ShieldDown() },
		},
	}

	return &ai.Comp{Fsm: &ai.Fsm{Initial: "Idle", Actions: actions}}
}
