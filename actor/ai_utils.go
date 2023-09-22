package actor

import (
	"game/libs/bump"
)

var (
	defaultViewHeight = 40.0
)

type Attack struct {
	AnimTag               string
	Damage, StaminaDamage float64
}

type AIConfig struct {
	ViewRect bump.Rect
	ViewDist float64

	CombatDist float64
	BackUpDist float64
	ReactDist  float64

	PaceReact      []AIWeightedState
	RunAttackReact []Attack

	Attacks       []Attack
	CombatOptions []AIWeightedState
}

// nolint: gomnd, nolintlint
func DefaultAIConfig() *AIConfig {
	// TODO: Review config attributes and make them more understandable.
	return &AIConfig{
		ViewDist:   70,
		CombatDist: 100,
		BackUpDist: 35,
		ReactDist:  20,

		PaceReact:      []AIWeightedState{{"Attack", 2}, {"Guard", 1}},
		RunAttackReact: nil,

		Attacks:       []Attack{{"Attack", 100, 20}},
		CombatOptions: []AIWeightedState{{"Pursuit", 100}, {"Pace", 2}, {"Wait", 1}, {"RunAttack", 1}, {"Attack", 1}, {"Guard", 1}},
	}
}

// TODO: Review speed changes, some speed transitions are not working as expected
// TODO: Sometimes Idle state resets and a new target is selected from nowhere.
func (a *Actor) SetDefaultAI(config *AIConfig) {
	if config == nil {
		config = DefaultAIConfig()
	}
	speed, maxSpeed := a.Speed, a.Body.MaxX

	a.AI.Actor = a
	fsm := NewAIFsm("Idle")

	fsm.SetAction("Idle", a.AI.IdleBuilder(config.ViewRect, config.ViewDist, defaultViewHeight, nil).Build())
	fsm.SetAction("Wait", a.AI.WaitBuilder(0.5, 1).SetCooldown(AICooldown{1, 2}).Build())

	fsm.SetAction("Pursuit", a.AI.PursuitBuilder(config.CombatDist, speed, maxSpeed, []AIWeightedState{{"Pace", 0}}).Build())

	fsm.SetAction("Pace", a.AI.PaceBuilder(config.BackUpDist, config.ReactDist, speed, maxSpeed, config.PaceReact).
		SetTimeout(AITimeout{AICombatState, 1, 1.5}).
		Build())

	fsm.SetAction("Guard", (&ActionBuilder{}).
		SetCooldown(AICooldown{3, 0}).
		SetTimeout(AITimeout{"Pace", 1, 2}).
		SetEntry(func() {
			a.Speed = -speed
			a.Body.MaxX = maxSpeed / 4
			a.ShieldUp()
		}).
		SetExit(func() { a.ShieldDown() }).
		Build())

	for _, attack := range config.Attacks {
		attack := attack
		fsm.SetAction(AIState(attack.AnimTag), a.AI.AnimBuilder(attack.AnimTag, nil).
			SetCooldown(AICooldown{1.5, 2.5}).
			AddCondition(a.AI.EnoughStamina(attack.StaminaDamage)).
			SetEntry(func() { a.Attack(attack.AnimTag, attack.Damage, attack.StaminaDamage) }).
			Build())
	}

	maxStamina := 0.0
	runAttackReact := config.RunAttackReact
	if len(runAttackReact) == 0 {
		runAttackReact = config.Attacks
	}
	runAttackStates := make([]AIWeightedState, len(runAttackReact))
	for i, attack := range runAttackReact {
		if attack.StaminaDamage > maxStamina {
			maxStamina = attack.StaminaDamage
		}
		runAttackStates[i] = AIWeightedState{AIState(attack.AnimTag), 1}
	}

	fsm.SetAction("RunAttack", (&ActionBuilder{}).
		SetCooldown(AICooldown{2, 3}).
		SetTimeout(AITimeout{"Pace", 3, 0}).
		AddCondition(a.AI.EnoughStamina(maxStamina)).
		SetEntry(a.AI.SetSpeedFunc(speed, maxSpeed)).
		AddReaction(a.AI.InRangeFunc(config.ReactDist), runAttackStates).
		Build())

	a.AI.Fsm = fsm
	a.AI.SetCombatOptions(config.CombatOptions)
}
