package entity

import (
	"game/comps/ai"
	"game/comps/anim"
)

type AIConfig struct {
	canBlock         bool
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
		canBlock:         false,
		viewDist:         100,
		combatDist:       100,
		backUpDist:       40,
		reactDist:        20,
		minAttackStamina: 0.2,
	}
}

// TODO: Review speed changes, some speed transitions are not working as expected.
func (a *Actor) SetDefaultAI(config *AIConfig) []ai.WeightedState {
	if config == nil {
		config = defaultConfig()
	}
	speed, maxSpeed := a.speed, a.Body.MaxX
	options := []ai.WeightedState{{"Pursuit", 1}, {"Pace", 0.5}, {"Wait", 0.16}, {"RunAttack", 0.16}, {"Attack", 0.16}}
	react := []ai.WeightedState{{"Attack", 1}, {"Pace", 0.8}}
	if config.canBlock {
		options = append(options, ai.WeightedState{"Guard", 0.1})
		react = append(react, ai.WeightedState{"Guard", 0.5})
	}

	fsm := ai.NewFsm("Idle")
	a.AI = &ai.Comp{Actor: a, Fsm: fsm}
	a.AddComponent(a.AI)

	a.AI.SetCombatOptions(options)
	fsm.SetAction("Idle", a.AI.IdleBuilder(config.viewDist, 40, nil).Build())
	fsm.SetAction("Wait", a.AI.WaitBuilder(1, 1.2).SetCooldown(ai.Cooldown{1, 2}).Build())
	fsm.SetAction("Pursuit", a.AI.PursuitBuilder(config.combatDist, speed, maxSpeed, []ai.WeightedState{{"Pace", 0}}).Build())

	fsm.SetAction("Pace", a.AI.PaceBuilder(config.backUpDist, config.reactDist, speed, maxSpeed, react).
		SetCooldown(ai.Cooldown{1, 2}).
		SetTimeout(ai.Timeout{ai.CombatState, 1, 2}).
		Build())

	fsm.SetAction("Attack", a.AI.AnimBuilder(anim.AttackTag, nil).
		SetCooldown(ai.Cooldown{2, 3}).
		AddCondition(a.AI.EnoughStamina(config.minAttackStamina)).
		SetEntry(func() { a.Attack(anim.AttackTag) }).
		Build())

	fsm.SetAction("Guard", (&ai.ActionBuilder{}).
		SetCooldown(ai.Cooldown{3, 0}).
		SetTimeout(ai.Timeout{"Pace", 1, 2}).
		SetEntry(func() { a.SetSpeed(-speed, maxSpeed/4); a.ShieldUp() }).
		SetExit(func() { a.ShieldDown() }).
		Build())

	fsm.SetAction("RunAttack", (&ai.ActionBuilder{}).
		SetCooldown(ai.Cooldown{2, 3}).
		SetTimeout(ai.Timeout{"Pace", 1, 2}).
		AddCondition(a.AI.EnoughStamina(config.minAttackStamina)).
		AddCondition(a.AI.OutRangeFunc(config.backUpDist)).
		SetEntry(a.AI.SetSpeedFunc(speed, maxSpeed)).
		AddReaction(a.AI.InRangeFunc(config.reactDist), []ai.WeightedState{{anim.AttackTag, 0}}).
		Build())

	return options
}
