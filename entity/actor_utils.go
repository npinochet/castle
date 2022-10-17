package entity

import (
	"game/comps/ai"
	"game/comps/anim"
)

type AIConfig struct {
	minDist, maxDist              float64
	attackCloseMax, attackMaxDist float64
	minAttackStamina              float64
}

// nolint: gomnd, nolintlint
func defaultConfig() *AIConfig {
	return &AIConfig{
		minDist:          20,
		attackCloseMax:   30,
		attackMaxDist:    40,
		maxDist:          50,
		minAttackStamina: 0.2,
	}
}

func (a *Actor) NewDefaultAI(config *AIConfig) *ai.Comp {
	if config == nil {
		config = defaultConfig()
	}
	speed := a.speed

	decide := []ai.WeightedState{{"Pursuit", 2}, {"BackUp", 0.5}, {"DecideAttack", 2}, {"Guard", 0.1}}
	attack := []ai.WeightedState{{"AttackClose", 1}, {"AttackFast", 1}}

	actions := map[ai.State]*ai.Action{
		"Act": {Next: func() []ai.WeightedState { return decide }},
		"DecideAttack": {
			Condition: func() bool {
				return a.AI.InTargetRange(0, config.attackMaxDist) && a.Stats.Stamina > config.minAttackStamina
			},
			Next: func() []ai.WeightedState { return attack },
		},
		"AttackClose": {
			Condition: func() bool {
				return a.AI.InTargetRange(0, config.attackCloseMax) && a.Stats.Stamina > config.minAttackStamina
			},
			Entry: func() { a.Attack() },
			Next: func() []ai.WeightedState {
				if a.Anim.State != anim.AttackTag {
					return []ai.WeightedState{{"Think", 0}}
				}

				return nil
			},
		},
		"AttackFast": {
			Condition: func() bool {
				return a.AI.InTargetRange(0, config.attackCloseMax) && a.Stats.Stamina > config.minAttackStamina
			},
			Entry: func() { a.Attack() },
			Next: func() []ai.WeightedState {
				if a.Anim.State != anim.AttackTag {
					return []ai.WeightedState{{"Think", 0}}
				}

				return nil
			},
		},
		"Guard": {
			Timeout:   ai.Timeout{"Think", 1, 0},
			Condition: func() bool { return a.AI.InTargetRange(0, config.maxDist) },
			Entry: func() {
				a.speed = 0
				a.ShieldUp()
			},
			Exit: func() { a.ShieldDown() },
		},
		"Idle": {
			Entry: func() { a.speed = 0 },
			Next: func() []ai.WeightedState {
				if a.AI.InTargetRange(0, config.maxDist) {
					return decide
				}

				return nil
			},
		},
		"Pursuit": {
			Timeout: ai.Timeout{"Act", 2, 0},
			Condition: func() bool {
				return a.AI.InTargetRange(config.minDist, -1) && a.Stats.Stamina > config.minAttackStamina
			},
			Entry: func() { a.speed = speed },
			Next: func() []ai.WeightedState {
				if !a.AI.InTargetRange(config.minDist, -1) {
					return decide
				}

				return nil
			},
		},
		"BackUp": {
			Timeout:   ai.Timeout{"Act", 2, 0},
			Condition: func() bool { return a.AI.InTargetRange(0, config.attackMaxDist) },
			Entry:     func() { a.speed = -speed },
			Next: func() []ai.WeightedState {
				if !a.AI.InTargetRange(0, config.maxDist) {
					return []ai.WeightedState{{"Think", 0}}
				}

				return nil
			},
		},
		"Think": {
			Timeout: ai.Timeout{"Act", 1, 2},
			Entry:   func() { a.speed = 0 },
		},
	}

	return &ai.Comp{Fsm: &ai.Fsm{Initial: "Idle", Actions: actions}}
}
