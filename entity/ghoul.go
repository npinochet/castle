package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/stats"
	"game/core"
)

const (
	ghoulAnimFile                               = "assets/ghoul"
	ghoulWidth, ghoulHeight                     = 9, 13
	ghoulOffsetX, ghoulOffsetY, ghoulOffsetFlip = -6.5, -4, 14
	ghoulSpeed                                  = 100
	ghoulMaxSpeed                               = 20
	ghoulDamage                                 = 20
)

type ghoul struct {
	*Actor
}

func NewGhoul(x, y, w, h float64, props map[string]string) *core.Entity {
	animc := &anim.Comp{FilesName: ghoulAnimFile, OX: ghoulOffsetX, OY: ghoulOffsetY, OXFlip: ghoulOffsetFlip}
	body := &body.Comp{W: ghoulWidth, H: ghoulHeight, MaxX: ghoulMaxSpeed}

	ghoul := &ghoul{
		Actor: NewActor(x, y, body, animc, &stats.Comp{MaxPoise: ghoulDamage}, ghoulDamage, ghoulDamage),
	}
	ghoul.speed = ghoulSpeed
	ghoul.AddComponent(ghoul)
	ghoul.Anim.Fsm.Transitions["AttackShort"] = anim.IdleTag
	ghoul.Anim.Fsm.Transitions["AttackLong"] = anim.IdleTag
	ghoul.setupAI()

	return &ghoul.Entity
}

func (g *ghoul) Init(entity *core.Entity) {
	hurtbox, err := g.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	g.Hitbox.PushHitbox(hurtbox, false)
}

func (g *ghoul) Update(dt float64) {
	g.ManageAnim([]string{"AttackShort", "AttackLong"})
	if g.Anim.State == anim.WalkTag && g.speed == 0 {
		g.Anim.SetState(anim.IdleTag)
	}
	if g.AI.Target != nil {
		if g.Anim.State == anim.WalkTag || g.Anim.State == anim.IdleTag {
			g.Anim.FlipX = g.AI.Target.X > g.X
		}
	}
	if (g.Anim.State != "AttackShort" && g.Anim.State != "AttackoLong") && g.Anim.State != anim.StaggerTag {
		if g.Anim.FlipX {
			g.Body.Vx += g.speed * dt
		} else {
			g.Body.Vx -= g.speed * dt
		}
	}

	if g.Stats.Health <= 0 {
		g.Remove()
	}
}

func (g *ghoul) setupAI() {
	aiConfig := DefaultAIConfig()
	aiConfig.attackDisable = true
	g.SetDefaultAI(aiConfig, []ai.WeightedState{{"AttackShort", 1}})

	combatOptions := []ai.WeightedState{{"Pursuit", 10}, {"Pace", 0.5}, {"Wait", 0.125}, {"RunAttack", 0.125}, {"AttackLong", 0.125}, {"AttackShort", 0.125}}
	g.AI.SetCombatOptions(combatOptions)
	g.AI.Fsm.SetAction("AttackShort", g.AI.AnimBuilder("AttackShort", nil).
		SetCooldown(ai.Cooldown{2, 3}).
		AddCondition(g.AI.EnoughStamina(0.2)).
		SetEntry(func() { g.Attack("AttackShort") }).
		Build())
	g.AI.Fsm.SetAction("AttackLong", g.AI.AnimBuilder("AttackLong", nil).
		SetCooldown(ai.Cooldown{2, 3}).
		AddCondition(g.AI.EnoughStamina(0.4)).
		SetEntry(func() { g.Attack("AttackLong") }).
		Build())
	g.AI.Fsm.SetAction("RunAttack", (&ai.ActionBuilder{}).
		SetCooldown(ai.Cooldown{2, 3}).
		SetTimeout(ai.Timeout{"Pace", 1, 2}).
		AddCondition(g.AI.EnoughStamina(aiConfig.minAttackStamina)).
		AddCondition(g.AI.OutRangeFunc(aiConfig.backUpDist)).
		SetEntry(g.AI.SetSpeedFunc(ghoulSpeed, ghoulMaxSpeed)).
		AddReaction(g.AI.InRangeFunc(aiConfig.reactDist), []ai.WeightedState{{"AttackLong", 1}, {"AttackShort", 1}}).
		Build())
}
