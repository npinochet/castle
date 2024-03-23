package entity

import (
	"game/comps/ai"
	"game/comps/basic/anim"
	"game/comps/basic/body"
	"game/comps/basic/stats"
	"game/core"
	"game/entity/defaults"
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

type ghoul struct {
	*defaults.Actor
	rocks int
}

func NewGhoul(x, y, _, _ float64, props *core.Property) *core.Entity {
	rocks, _ := strconv.Atoi(props.Custom["rocks"])
	attackTags := []string{"AttackShort", "AttackLong"}
	ghoul := &ghoul{
		Actor: defaults.NewActor(x, y, ghoulWidth, ghoulHeight, attackTags),
		rocks: rocks,
	}
	ghoul.Anim = &anim.Comp{FilesName: ghoulAnimFile, OX: ghoulOffsetX, OY: ghoulOffsetY, OXFlip: ghoulOffsetFlip, FlipX: props.FlipX}
	ghoul.Body = &body.Comp{MaxX: ghoulMaxSpeed}
	ghoul.Stats = &stats.Comp{MaxHealth: ghoulHealth, MaxPoise: ghoulDamage}
	ghoul.Control.Speed = ghoulSpeed

	var view bump.Rect
	if v := props.View; v != nil {
		view = bump.NewRect(v.X, v.Y, v.Width, v.Height)
	}
	if props.AI == "poacher" {
		ghoul.setupPoacherAI(view)
	} else {
		ghoul.setupDefaultAI(view)
	}
	ghoul.SetupComponents()
	ghoul.AddComponent(ghoul)

	return ghoul.Entity
}

func (g *ghoul) Init(_ *core.Entity) {
	hurtbox, err := g.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	g.Hitbox.PushHitbox(hurtbox, false)
}

func (g *ghoul) Update(dt float64) {
	g.Control.SimpleUpdate(dt, g.AI.Target)
}

func (g *ghoul) ThrowRock() {
	tag := "Throw"
	if g.Anim.State == tag || g.Anim.State == anim.StaggerTag {
		return
	}
	g.Control.ResetState(tag)
	g.Anim.OnFrame(ghoulThrowFrame, func() {
		g.World.AddEntity(NewRock(g.X-2, g.Y-4, g.Actor))
		g.rocks--
	})
}

func (g *ghoul) setupDefaultAI(view bump.Rect) {
	config := defaults.DefaultAIConfig()
	config.ViewRect = view
	config.PaceReact = []ai.WeightedState{{"AttackShort", 1}, {"Wait", 0}}
	config.Attacks = []defaults.Attack{{"AttackShort", ghoulDamage}, {"AttackLong", ghoulDamage}}
	config.CombatOptions = []ai.WeightedState{
		{"Pursuit", 100},
		{"Pace", 2},
		{"Wait", 1},
		{"RunAttack", 1},
		{"AttackLong", 1},
		{"AttackShort", 1},
		{"Throw", 1.5},
	}
	g.SetDefaultAI(config)

	g.AI.Fsm.SetAction("Throw", g.AnimBuilder("Throw", nil).
		SetCooldown(ai.Cooldown{2, 3}).
		AddCondition(func() bool { return g.rocks > 0 }).
		AddCondition(g.OutRangeFunc(config.BackUpDist)).
		SetEntry(func() { g.ThrowRock() }).
		Build())
}

func (g *ghoul) setupPoacherAI(view bump.Rect) {
	config := defaults.DefaultAIConfig()
	config.ViewRect = view
	config.CombatOptions = []ai.WeightedState{{"Wait", 0}, {"Throw", 1}}
	g.SetDefaultAI(config)

	g.AI.Fsm.SetAction("Wait", g.WaitBuilder(2, 0).
		AddReaction(func() bool {
			if g.rocks <= 0 {
				g.setupDefaultAI(view)

				return true
			}

			return false
		}, nil).
		Build())
	g.AI.Fsm.SetAction("Throw", g.AnimBuilder("Throw", nil).
		SetCooldown(ai.Cooldown{1, 0}).
		AddCondition(func() bool { return g.rocks > 0 }).
		AddCondition(g.OutRangeFunc(config.BackUpDist)).
		SetEntry(func() { g.ThrowRock() }).
		Build())
}
