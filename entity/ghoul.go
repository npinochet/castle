package entity

import (
	"game/actor"
	"game/core"
	"game/libs/bump"
	"strconv"
)

const (
	ghoulAnimFile                               = "assets/ghoul"
	ghoulWidth, ghoulHeight                     = 9, 13
	ghoulOffsetX, ghoulOffsetY, ghoulOffsetFlip = -6.5, -4, 14
	ghoulSpeed, ghoulMaxSpeed                   = 100, 30
	ghoulHealth                                 = 70
	ghoulDamage                                 = 18
	ghoulThrowFrame                             = 2
)

type Ghoul struct {
	actor.Actor
	rocks int
}

func NewGhoul(x, y, _, _ float64, props *core.Property) core.Entity {
	rocks, _ := strconv.Atoi(props.Custom["rocks"])
	ghoul := &Ghoul{
		Actor: actor.NewActor(x, y, ghoulWidth, ghoulHeight, []string{"AttackShort", "AttackLong"}),
		rocks: rocks,
	}
	ghoul.Speed = ghoulSpeed
	ghoul.Body.MaxX = ghoulMaxSpeed
	ghoul.Stats.MaxPoise = ghoulDamage
	ghoul.Stats.MaxHealth = ghoulHealth
	ghoul.Anim.FilesName = ghoulAnimFile
	ghoul.Anim.OX, ghoul.Anim.OY = ghoulOffsetX, ghoulOffsetY
	ghoul.Anim.OXFlip = ghoulOffsetFlip
	ghoul.Anim.FlipX = props.FlipX

	var view bump.Rect
	if props.View != nil {
		view = bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
	}
	if props.AI == "poacher" {
		ghoul.setupPoacherAI(view)
	} else {
		ghoul.setupDefaultAI(view)
	}

	return ghoul
}

func (g *Ghoul) Update(dt float64) {
	g.Actor.Update(dt)
	g.BasicUpdate(dt)
}

func (g *Ghoul) ThrowRock() {
	tag := "Throw"
	if g.Stats.Stamina <= 0 || g.Anim.State == tag || g.Anim.State == actor.StaggerTag {
		return
	}
	g.ResetState(tag)
	g.Anim.OnFrames(func(frame int) {
		if frame == ghoulThrowFrame {
			g.Stats.AddStamina(-rockDamage)
			g.World.AddEntity(NewRock(g.X-2, g.Y-4, &g.Actor))
			g.rocks--
			g.Anim.OnFrames(nil)
		}
	})
}

func (g *Ghoul) setupDefaultAI(view bump.Rect) {
	aiConfig := actor.DefaultAIConfig()
	aiConfig.ViewRect = view
	aiConfig.PaceReact = []actor.AIWeightedState{{"AttackShort", 1}, {"Wait", 0}}
	aiConfig.Attacks = []actor.Attack{{"AttackShort", ghoulDamage, 20}, {"AttackLong", ghoulDamage, 40}}
	aiConfig.CombatOptions = []actor.AIWeightedState{
		{"Pursuit", 100},
		{"Pace", 2},
		{"Wait", 1},
		{"RunAttack", 1},
		{"AttackLong", 1},
		{"AttackShort", 1},
		{"Throw", 1.5},
	}

	g.Speed, g.Body.MaxX = ghoulSpeed, ghoulMaxSpeed
	g.SetDefaultAI(aiConfig)

	g.AI.Fsm.SetAction("Throw", g.AI.AnimBuilder("Throw", nil).
		SetCooldown(actor.AICooldown{2, 3}).
		AddCondition(g.AI.EnoughStamina(rockDamage)).
		AddCondition(func() bool { return g.rocks > 0 }).
		AddCondition(g.AI.OutRangeFunc(aiConfig.BackUpDist)).
		SetEntry(func() { g.ThrowRock() }).
		Build())
}

func (g *Ghoul) setupPoacherAI(view bump.Rect) {
	aiConfig := actor.DefaultAIConfig()
	aiConfig.ViewRect = view
	aiConfig.CombatOptions = []actor.AIWeightedState{{"Wait", 0}, {"Throw", 1}}

	g.Speed, g.Body.MaxX = ghoulSpeed, ghoulMaxSpeed
	g.SetDefaultAI(aiConfig)

	g.AI.Fsm.SetAction("Wait", g.AI.WaitBuilder(2, 0).
		AddReaction(func() bool {
			if g.rocks <= 0 {
				g.setupDefaultAI(view)

				return true
			}

			return false
		}, nil).
		Build())
	g.AI.Fsm.SetAction("Throw", g.AI.AnimBuilder("Throw", nil).
		SetCooldown(actor.AICooldown{1, 0}).
		AddCondition(func() bool { return g.rocks > 0 }).
		AddCondition(g.AI.OutRangeFunc(aiConfig.BackUpDist)).
		SetEntry(func() { g.ThrowRock() }).
		Build())
}
