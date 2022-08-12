package entity

import (
	"fmt"
	"game/comps/ai"
	"game/comps/anim"
)

type Config struct {
	minDist, maxDist              float64
	attackCloseMax, attackMaxDist float64
	minAttackStamina              float64
}

// nolint: gomnd
func defaultConfig() *Config {
	return &Config{minDist: 20,
		attackCloseMax:   30,
		attackMaxDist:    40,
		maxDist:          50,
		minAttackStamina: 0.2,
	}
}

func NewDefaultAI(actor *Actor, config *Config) *ai.Comp {
	if config == nil {
		config = defaultConfig()
	}
	speed := actor.speed
	fms := &ai.Fsm{Initial: "Ready"}

	think := []ai.Action{{State: "Thinking", Timeout: ai.Timeout{"Ready", 0.4, 0.8}}}
	decide := []ai.Action{
		{State: "Approaching", Weight: 2, Timeout: ai.Timeout{"Ready", 2.0, 0}, Condition: inRangeStamina(config.minDist, -1, config.minAttackStamina, actor)},
		{State: "BackingUp", Weight: 0.1, Timeout: ai.Timeout{"Ready", 2.0, 0}, Condition: inRange(0, config.attackMaxDist, nil)},
		{State: "Attack", Weight: 2, Condition: inRangeStamina(0, config.attackMaxDist, config.minAttackStamina, actor)},
		{State: "Guarding", Weight: 0.2, Timeout: ai.Timeout{"Unguarding", 1.0, 0}, Condition: inRange(0, config.maxDist, nil)},
	}
	attack := []ai.Action{
		{State: "AttackClose", Condition: inRangeStamina(0, config.attackCloseMax, config.minAttackStamina, actor)},
		{State: "AttackFast", Condition: inRangeStamina(0, config.attackCloseMax, config.minAttackStamina, actor)},
	}

	fms.Updates = map[string]ai.UpdateFunc{
		"Ready":  func() []ai.Action { return decide },
		"Attack": func() []ai.Action { return attack },
		"Approaching": func() []ai.Action {
			if actor.AI.InTargetRange(config.minDist, -1) {
				actor.speed = speed

				return nil
			}

			return decide
		},
		"BackingUp": func() []ai.Action {
			if actor.AI.InTargetRange(0, config.maxDist) {
				actor.speed = -speed

				return nil
			}

			return think
		},
		"Thinking": func() []ai.Action {
			actor.speed = 0

			return nil
		},
		"AttackClose": func() []ai.Action {
			actor.Attack()

			return waitAnim(fms, actor, anim.AttackTag, think)
		},
		"AttackFast": func() []ai.Action {
			actor.Attack()

			return waitAnim(fms, actor, anim.AttackTag, think)
		},
		"Guarding": func() []ai.Action {
			actor.speed = 0
			actor.ShieldUp()

			return nil
		},
		"Unguarding": func() []ai.Action {
			actor.ShieldDown()

			return think
		},
	}

	return &ai.Comp{Fsm: fms}
}

func inRange(min, max float64, misc func(c *ai.Comp) bool) func(c *ai.Comp) bool {
	return func(c *ai.Comp) bool { return c.InTargetRange(min, max) && (misc == nil || misc(c)) }
}

func inRangeStamina(min, max, minStamina float64, actor *Actor) func(c *ai.Comp) bool {
	return inRange(min, max, func(c *ai.Comp) bool { return actor.Stats.Stamina > minStamina })
}

func waitAnim(fsm *ai.Fsm, actor *Actor, tag string, after []ai.Action) []ai.Action {
	state := fmt.Sprintf("waitAnim%s", tag)
	if _, ok := fsm.Updates[state]; !ok {
		fsm.Updates[state] = func() []ai.Action {
			if actor.Anim.State == anim.AttackTag {
				return nil
			}

			return after
		}
	}

	return []ai.Action{{State: state}}
}
