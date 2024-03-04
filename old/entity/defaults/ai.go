package defaults

import (
	"game/comps/ai"
	"game/libs/bump"
)

const defaultViewHeight = 40.0

type Attack struct {
	AnimTag string
	Damage  float64
}

type AIConfig struct {
	ViewRect bump.Rect
	ViewDist float64

	CombatDist float64
	BackUpDist float64
	ReactDist  float64

	PaceReact      []ai.WeightedState
	RunAttackReact []Attack

	Attacks       []Attack
	CombatOptions []ai.WeightedState
}

// nolint: gomnd, nolintlint
func DefaultAIConfig() *AIConfig {
	// TODO: Review config attributes and make them more understandable.
	return &AIConfig{
		ViewDist:   70,
		CombatDist: 100,
		BackUpDist: 35,
		ReactDist:  20,

		PaceReact:      []ai.WeightedState{{"Attack", 2}, {"Guard", 1}},
		RunAttackReact: nil,

		Attacks:       []Attack{{"Attack", 100}},
		CombatOptions: []ai.WeightedState{{"Pursuit", 100}, {"Pace", 2}, {"Wait", 1}, {"RunAttack", 1}, {"Attack", 1}, {"Guard", 1}},
	}
}

// TODO: Review speed changes, some speed transitions are not working as expected
// TODO: Sometimes Idle state resets and a new target is selected from nowhere.
func (a *Actor) SetDefaultAI(config *AIConfig) {
	if config == nil {
		config = DefaultAIConfig()
	}
	speed, maxSpeed := a.Control.Speed, a.Body.MaxX
	fsm := ai.NewFsm("Idle")

	fsm.SetAction("Idle", a.IdleBuilder(config.ViewRect, config.ViewDist, defaultViewHeight, nil).Build())
	fsm.SetAction("Wait", a.WaitBuilder(0.5, 1).SetCooldown(ai.Cooldown{1, 2}).Build())

	fsm.SetAction("Pursuit", a.PursuitBuilder(config.CombatDist, speed, maxSpeed, []ai.WeightedState{{"Pace", 0}}).Build())

	fsm.SetAction("Pace", a.PaceBuilder(config.BackUpDist, config.ReactDist, speed, maxSpeed, config.PaceReact).
		SetTimeout(ai.Timeout{ai.CombatState, 1, 1.5}).
		Build())

	fsm.SetAction("Guard", (&ActionBuilder{}).
		SetCooldown(ai.Cooldown{3, 0}).
		SetTimeout(ai.Timeout{"Pace", 1, 2}).
		SetEntry(func() { a.Control.SetSpeed(-speed, maxSpeed/4); a.Control.ShieldUp() }).
		SetExit(func() { a.Control.ShieldDown() }).
		Build())

	for _, attack := range config.Attacks {
		attack := attack
		fsm.SetAction(ai.State(attack.AnimTag), a.AnimBuilder(attack.AnimTag, nil).
			SetCooldown(ai.Cooldown{1.5, 2.5}).
			SetEntry(func() { a.Control.Attack(attack.AnimTag, attack.Damage, 0) }).
			Build())
	}

	runAttackReact := config.RunAttackReact
	if len(runAttackReact) == 0 {
		runAttackReact = config.Attacks
	}
	runAttackStates := make([]ai.WeightedState, len(runAttackReact))
	for i, attack := range runAttackReact {
		runAttackStates[i] = ai.WeightedState{ai.State(attack.AnimTag), 1}
	}

	fsm.SetAction("RunAttack", (&ActionBuilder{}).
		SetCooldown(ai.Cooldown{2, 3}).
		SetTimeout(ai.Timeout{"Pace", 3, 0}).
		SetEntry(a.SetSpeedFunc(speed, maxSpeed)).
		AddReaction(a.InRangeFunc(config.ReactDist), runAttackStates).
		Build())

	a.AI.Fsm = fsm
	a.AI.SetCombatOptions(config.CombatOptions)
}
