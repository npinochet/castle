package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/stats"
	"game/core"
	"game/libs/bump"
	"strconv"
	"strings"
)

const (
	ghoulAnimFile                               = "assets/ghoul"
	ghoulWidth, ghoulHeight                     = 9, 13
	ghoulOffsetX, ghoulOffsetY, ghoulOffsetFlip = -6.5, -4, 14
	ghoulSpeed                                  = 100
	ghoulMaxSpeed                               = 20
	ghoulDamage                                 = 20
	ghoulThrowFrame                             = 2
)

type ghoul struct {
	*Actor
	rocks int
}

func NewGhoul(x, y, w, h float64, props map[string]string) *core.Entity {
	animc := &anim.Comp{FilesName: ghoulAnimFile, OX: ghoulOffsetX, OY: ghoulOffsetY, OXFlip: ghoulOffsetFlip}
	animc.FlipX = props[core.HorizontalProp] == "true"

	body := &body.Comp{MaxX: ghoulMaxSpeed}

	rocks, _ := strconv.Atoi(props["rocks"])
	ghoul := &ghoul{
		Actor: NewActor(x, y, ghoulWidth, ghoulHeight, body, animc, &stats.Comp{MaxPoise: ghoulDamage + 1}),
		rocks: rocks,
	}
	ghoul.speed = ghoulSpeed
	ghoul.AddComponent(ghoul)

	var view bump.Rect
	if viewStrings := strings.Split(props[core.ViewProp], ","); len(viewStrings) > 1 {
		view.X, _ = strconv.ParseFloat(viewStrings[0], 64)
		view.Y, _ = strconv.ParseFloat(viewStrings[1], 64)
		view.W, _ = strconv.ParseFloat(viewStrings[2], 64)
		view.H, _ = strconv.ParseFloat(viewStrings[3], 64)
	}
	if props["ai"] == "poacher" {
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
	g.ManageAnim([]string{"AttackShort", "AttackLong"})
	if g.Anim.State == anim.WalkTag && g.speed == 0 {
		g.Anim.SetState(anim.IdleTag)
	}
	if g.AI.Target != nil {
		if g.Anim.State == anim.WalkTag || g.Anim.State == anim.IdleTag {
			g.Anim.FlipX = g.AI.Target.X > g.X
		}
	}
	if (g.Anim.State != "AttackShort" && g.Anim.State != "AttackLong") && g.Anim.State != anim.StaggerTag {
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

func (g *ghoul) ThrowRock() {
	tag := "Throw"
	if g.Stats.Stamina <= 0 || g.Anim.State == tag || g.Anim.State == anim.StaggerTag {
		return
	}
	g.Anim.SetState(tag)
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
	aiConfig.PaceReact = []ai.WeightedState{{"AttackShort", 1}}
	aiConfig.Attacks = []Attack{{"AttackShort", ghoulDamage, 20}, {"AttackLong", ghoulDamage, 40}}
	aiConfig.CombatOptions = []ai.WeightedState{{"Pursuit", 10}, {"Pace", 0.5}, {"Wait", 0.125}, {"RunAttack", 0.125}, {"AttackLong", 0.125}, {"AttackShort", 0.125}, {"Throw", 0.3}}

	g.speed, g.Body.MaxX = ghoulSpeed, ghoulMaxSpeed
	g.SetDefaultAI(aiConfig)

	g.AI.Fsm.SetAction("Throw", g.AI.AnimBuilder("Throw", nil).
		SetCooldown(ai.Cooldown{2, 3}).
		AddCondition(g.AI.EnoughStamina(0.2)).
		AddCondition(func() bool { return g.rocks > 0 }).
		AddCondition(g.AI.OutRangeFunc(aiConfig.backUpDist)).
		SetEntry(func() { g.ThrowRock() }).
		Build())
}

func (g *ghoul) setupPoacherAI(view bump.Rect) {
	aiConfig := DefaultAIConfig()
	aiConfig.viewRect = view
	aiConfig.CombatOptions = []ai.WeightedState{{"Wait", 0}, {"Throw", 1}}

	g.speed, g.Body.MaxX = ghoulSpeed, ghoulMaxSpeed
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
