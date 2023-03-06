package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/stats"
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

type ghoul struct {
	*Actor
	rocks int
}

func NewGhoul(x, y, w, h float64, props *core.Property) *core.Entity {
	animc := &anim.Comp{FilesName: ghoulAnimFile, OX: ghoulOffsetX, OY: ghoulOffsetY, OXFlip: ghoulOffsetFlip}
	animc.FlipX = props.FlipX

	body := &body.Comp{MaxX: ghoulMaxSpeed}

	rocks, _ := strconv.Atoi(props.Custom["rocks"])
	attackTags := []string{"AttackShort", "AttackLong"}
	ghoul := &ghoul{
		Actor: NewActor(x, y, ghoulWidth, ghoulHeight, attackTags, animc, body, &stats.Comp{MaxHealth: ghoulHealth, MaxPoise: ghoulDamage}),
		rocks: rocks,
	}
	ghoul.Speed = ghoulSpeed
	ghoul.AddComponent(ghoul)

	var view bump.Rect
	if props.View != nil {
		view = bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
	}
	if props.AI == "poacher" {
		ghoul.setupPoacherAI(view)
	} else {
		ghoul.setupDefaultAI(view)
	}

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
	g.SimpleUpdate(dt)
}

func (g *ghoul) ThrowRock() {
	tag := "Throw"
	if g.Stats.Stamina <= 0 || g.Anim.State == tag || g.Anim.State == anim.StaggerTag {
		return
	}
	g.ResetState(tag)
	g.Anim.OnFrames(func(frame int) {
		if frame == ghoulThrowFrame {
			g.Stats.AddStamina(-rockDamage)
			g.World.AddEntity(NewRock(g.X-2, g.Y-4, g.Actor))
			g.rocks--
			g.Anim.OnFrames(nil)
		}
	})
}

func (g *ghoul) setupDefaultAI(view bump.Rect) {
	aiConfig := DefaultAIConfig()
	aiConfig.viewRect = view
	aiConfig.PaceReact = []ai.WeightedState{{"AttackShort", 1}, {"Wait", 0}}
	aiConfig.Attacks = []Attack{{"AttackShort", ghoulDamage, 20}, {"AttackLong", ghoulDamage, 40}}
	aiConfig.CombatOptions = []ai.WeightedState{
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
		SetCooldown(ai.Cooldown{2, 3}).
		AddCondition(g.AI.EnoughStamina(rockDamage)).
		AddCondition(func() bool { return g.rocks > 0 }).
		AddCondition(g.AI.OutRangeFunc(aiConfig.backUpDist)).
		SetEntry(func() { g.ThrowRock() }).
		Build())
}

func (g *ghoul) setupPoacherAI(view bump.Rect) {
	aiConfig := DefaultAIConfig()
	aiConfig.viewRect = view
	aiConfig.CombatOptions = []ai.WeightedState{{"Wait", 0}, {"Throw", 1}}

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
		SetCooldown(ai.Cooldown{1, 0}).
		AddCondition(func() bool { return g.rocks > 0 }).
		AddCondition(g.AI.OutRangeFunc(aiConfig.backUpDist)).
		SetEntry(func() { g.ThrowRock() }).
		Build())
}
