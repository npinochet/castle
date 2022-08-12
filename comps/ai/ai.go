package ai

import (
	"game/core"
	"game/utils"
	"math/rand"
)

func (c *Comp) IsActive() bool        { return c.active }
func (c *Comp) SetActive(active bool) { c.active = active }

type Timeout struct {
	Target                string
	Duration, MaxDuration float64
}

type Action struct {
	State     string
	Condition func(c *Comp) bool
	Weight    float64
	Timeout   Timeout
}

type UpdateFunc func() []Action

type Fsm struct {
	Initial, State string
	Updates        map[string]UpdateFunc
	Comp           *Comp
	timer          float64
	target         string
}

func (c *Fsm) Update(dt float64) {
	if c.State == "" {
		c.State = c.Initial
	}
	if update := c.Updates[c.State]; update != nil {
		if actions := update(); actions != nil {
			c.State = c.StateFromActions(actions)
		}
	}
	if c.target != "" {
		c.timer -= dt
		if c.timer <= 0 {
			c.State = c.target
			c.target = ""
		}
	}
}

func (c *Fsm) StateFromActions(actions []Action) string {
	c.target = ""

	var selected []Action
	if len(actions) == 1 {
		selected = append(selected, actions[0])
	}

	totalWeight := 0.0
	for _, a := range actions {
		if a.Condition == nil || a.Condition(c.Comp) {
			weight := a.Weight
			if weight == 0 {
				weight = 1.0 / float64(len(actions))
			}
			totalWeight += weight
			selected = append(selected, a)
		}
	}

	r := rand.Float64() * totalWeight
	for _, a := range selected {
		weight := a.Weight
		if weight == 0 {
			weight = 1.0 / float64(len(actions))
		}
		if r -= weight; r <= 0 {
			if a.Timeout.Duration > 0 {
				c.timer = a.Timeout.Duration
				if a.Timeout.MaxDuration > a.Timeout.Duration {
					c.timer += rand.Float64() * (a.Timeout.MaxDuration - a.Timeout.Duration)
				}
				c.target = a.Timeout.Target
			}

			return a.State
		}
	}

	return ""
}

type Comp struct {
	active         bool
	Fsm            *Fsm
	Target, entity *core.Entity
}

func (c *Comp) Init(entity *core.Entity) {
	c.entity = entity
	if c.Fsm != nil { //What happens here?
		c.Fsm.Comp = c
	}
}

func (c *Comp) Update(dt float64) {
	c.Fsm.Update(dt)
}

func (c *Comp) InTargetRange(minDist, maxDist float64) bool {
	dist := c.distFromTarget()
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

func (c *Comp) distFromTarget() float64 {
	if c.Target == nil {
		return -1
	}
	tx, ty := c.Target.Position()
	x, y := c.entity.Position()

	return utils.Distante(x, y, tx, ty)
}
