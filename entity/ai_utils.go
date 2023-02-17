package entity

import (
	"game/comps/ai"
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
	viewRect bump.Rect
	viewDist float64

	combatDist float64
	backUpDist float64
	reactDist  float64

	PaceReact      []ai.WeightedState
	RunAttackReact []Attack

	Attacks       []Attack
	CombatOptions []ai.WeightedState
}

// nolint: gomnd, nolintlint
func DefaultAIConfig() *AIConfig {
	// TODO: Review config attributes and make them more understandable.
	return &AIConfig{
		viewDist:   100,
		combatDist: 100,
		backUpDist: 40,
		reactDist:  20,

		PaceReact:      []ai.WeightedState{{"Attack", 2}, {"Guard", 1}},
		RunAttackReact: nil,

		Attacks:       []Attack{{"Attack", 100, 20}},
		CombatOptions: []ai.WeightedState{{"Pursuit", 100}, {"Pace", 2}, {"Wait", 1}, {"RunAttack", 1}, {"Attack", 1}, {"Guard", 1}},
	}
}

// TODO: Review speed changes, some speed transitions are not working as expected.
func (a *Actor) SetDefaultAI(config *AIConfig) {
	if config == nil {
		config = DefaultAIConfig()
	}
	speed, maxSpeed := a.speed, a.Body.MaxX

	if a.AI == nil {
		a.AI = &ai.Comp{Actor: a}
		a.AddComponent(a.AI)
	}
	fsm := ai.NewFsm("Idle")

	fsm.SetAction("Idle", a.AI.IdleBuilder(config.viewRect, config.viewDist, defaultViewHeight, nil).Build())
	fsm.SetAction("Wait", a.AI.WaitBuilder(1, 1.2).SetCooldown(ai.Cooldown{1, 2}).Build())

	fsm.SetAction("Pursuit", a.AI.PursuitBuilder(config.combatDist, speed, maxSpeed, []ai.WeightedState{{"Pace", 0}}).Build())

	fsm.SetAction("Pace", a.AI.PaceBuilder(config.backUpDist, config.reactDist, speed, maxSpeed, config.PaceReact).
		SetTimeout(ai.Timeout{ai.CombatState, 2, 3}).
		Build())

	fsm.SetAction("Guard", (&ai.ActionBuilder{}).
		SetCooldown(ai.Cooldown{3, 0}).
		SetTimeout(ai.Timeout{"Pace", 1, 2}).
		SetEntry(func() { a.SetSpeed(-speed, maxSpeed/4); a.ShieldUp() }).
		SetExit(func() { a.ShieldDown() }).
		Build())

	for _, attack := range config.Attacks {
		fsm.SetAction(ai.State(attack.AnimTag), a.AI.AnimBuilder(attack.AnimTag, nil).
			SetCooldown(ai.Cooldown{1.5, 2.5}).
			AddCondition(a.AI.EnoughStamina(attack.StaminaDamage)).
			SetEntry(func() { a.Attack(attack.AnimTag, attack.Damage, attack.StaminaDamage) }).
			Build())
	}

	var runAttackStates []ai.WeightedState
	maxStamina := 0.0
	runAttackReact := config.RunAttackReact
	if len(runAttackReact) == 0 {
		runAttackReact = config.Attacks
	}
	for _, attack := range runAttackReact {
		if attack.StaminaDamage > maxStamina {
			maxStamina = attack.StaminaDamage
		}
		runAttackStates = append(runAttackStates, ai.WeightedState{ai.State(attack.AnimTag), 1})
	}

	fsm.SetAction("RunAttack", (&ai.ActionBuilder{}).
		SetCooldown(ai.Cooldown{2, 3}).
		SetTimeout(ai.Timeout{"Pace", 1, 2}).
		AddCondition(a.AI.EnoughStamina(maxStamina)).
		AddCondition(a.AI.OutRangeFunc(config.backUpDist)).
		SetEntry(a.AI.SetSpeedFunc(speed, maxSpeed)).
		AddReaction(a.AI.InRangeFunc(config.reactDist), runAttackStates).
		Build())

	a.AI.Fsm = fsm
	a.AI.SetCombatOptions(config.CombatOptions)
}
