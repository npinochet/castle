package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/libs/bump"
)

type AIConfig struct {
	viewRect         bump.Rect
	viewDist         float64
	combatDist       float64
	backUpDist       float64
	reactDist        float64
	minAttackStamina float64
}

// nolint: gomnd, nolintlint
func DefaultAIConfig() *AIConfig {
	// TODO: Review config attributes and make them more understandable.
	return &AIConfig{
		viewDist:         100,
		combatDist:       100,
		backUpDist:       40,
		reactDist:        20,
		minAttackStamina: 0.2,
	}
}

// TODO: Review speed changes, some speed transitions are not working as expected.
func (a *Actor) SetDefaultAI(config *AIConfig, react []ai.WeightedState) {
	if config == nil {
		config = DefaultAIConfig()
	}
	speed, maxSpeed := a.speed, a.Body.MaxX
	options := []ai.WeightedState{{"Pursuit", 10}, {"Pace", 0.5}, {"Wait", 0.125}, {"RunAttack", 0.125}, {"Attack", 0.125}, {"Guard", 0.125}}
	if react == nil {
		react = []ai.WeightedState{{"Attack", 1}, {"Guard", 0.5}}
	}

	if a.AI == nil {
		a.AI = &ai.Comp{Actor: a}
		a.AddComponent(a.AI)
	}
	fsm := ai.NewFsm("Idle")

	fsm.SetAction("Idle", a.AI.IdleBuilder(config.viewRect, config.viewDist, 40, nil).Build())
	fsm.SetAction("Wait", a.AI.WaitBuilder(1, 1.2).SetCooldown(ai.Cooldown{1, 2}).Build())

	fsm.SetAction("Pursuit", a.AI.PursuitBuilder(config.combatDist, speed, maxSpeed, []ai.WeightedState{{"Pace", 0}}).Build())

	fsm.SetAction("Pace", a.AI.PaceBuilder(config.backUpDist, config.reactDist, speed, maxSpeed, react).
		SetTimeout(ai.Timeout{ai.CombatState, 2, 3}).
		Build())

	fsm.SetAction("Guard", (&ai.ActionBuilder{}).
		SetCooldown(ai.Cooldown{3, 0}).
		SetTimeout(ai.Timeout{"Pace", 1, 2}).
		SetEntry(func() { a.SetSpeed(-speed, maxSpeed/4); a.ShieldUp() }).
		SetExit(func() { a.ShieldDown() }).
		Build())

	fsm.SetAction("Attack", a.AI.AnimBuilder(anim.AttackTag, nil).
		SetCooldown(ai.Cooldown{2, 3}).
		AddCondition(a.AI.EnoughStamina(config.minAttackStamina)).
		SetEntry(func() { a.Attack(anim.AttackTag) }).
		Build())
	fsm.SetAction("RunAttack", (&ai.ActionBuilder{}).
		SetCooldown(ai.Cooldown{2, 3}).
		SetTimeout(ai.Timeout{"Pace", 1, 2}).
		AddCondition(a.AI.EnoughStamina(config.minAttackStamina)).
		AddCondition(a.AI.OutRangeFunc(config.backUpDist)).
		SetEntry(a.AI.SetSpeedFunc(speed, maxSpeed)).
		AddReaction(a.AI.InRangeFunc(config.reactDist), []ai.WeightedState{{anim.AttackTag, 0}}).
		Build())

	a.AI.Fsm = fsm
	a.AI.SetCombatOptions(options)
}

func (a *Actor) AddAttackAIOption(options []ai.WeightedState, state ai.State, weight float64, attackTag string, minStamina float64) {
	options = append(options, ai.WeightedState{state, weight})
	a.AI.SetCombatOptions(options)

	a.AI.Fsm.SetAction(state, a.AI.AnimBuilder(attackTag, nil).
		SetCooldown(ai.Cooldown{2, 3}).
		AddCondition(a.AI.EnoughStamina(minStamina)).
		SetEntry(func() { a.Attack(attackTag) }).
		Build())
}
