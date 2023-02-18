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
	skelemanAnimFile                                     = "assets/skeleman"
	skelemanWidth, skelemanHeight                        = 8, 12
	skelemanOffsetX, skelemanOffsetY, skelemanOffsetFlip = -12, -5, 20
	skelemanSpeed                                        = 100
	skelemanMaxSpeed                                     = 20
	skelemanDamage                                       = 20
	skelemanPoise                                        = 30
)

type skeleman struct {
	*Actor
}

func NewSkeleman(x, y, w, h float64, props map[string]string) *core.Entity {
	animc := &anim.Comp{FilesName: skelemanAnimFile, OX: skelemanOffsetX, OY: skelemanOffsetY, OXFlip: skelemanOffsetFlip}
	animc.FlipX = props[core.HorizontalProp] == "true"

	body := &body.Comp{MaxX: skelemanMaxSpeed}

	skeleman := &skeleman{
		Actor: NewActor(x, y, skelemanWidth, skelemanHeight, body, animc, &stats.Comp{MaxPoise: skelemanPoise}, []string{"AttackShort", "AttackLong"}),
	}
	skeleman.Speed = skelemanSpeed
	skeleman.AddComponent(skeleman)

	var view bump.Rect
	if viewStrings := strings.Split(props[core.ViewProp], ","); len(viewStrings) > 1 {
		view.X, _ = strconv.ParseFloat(viewStrings[0], 64)
		view.Y, _ = strconv.ParseFloat(viewStrings[1], 64)
		view.W, _ = strconv.ParseFloat(viewStrings[2], 64)
		view.H, _ = strconv.ParseFloat(viewStrings[3], 64)
	}

	skeleman.setupAI(view)

	return &skeleman.Entity
}

func (g *skeleman) Init(entity *core.Entity) {
	hurtbox, err := g.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	g.Hitbox.PushHitbox(hurtbox, false)
}

func (g *skeleman) Update(dt float64) {
	g.SimpleUpdate(dt)
}

func (g *skeleman) setupAI(view bump.Rect) {
	aiConfig := DefaultAIConfig()
	aiConfig.viewRect = view
	aiConfig.PaceReact = []ai.WeightedState{{"AttackShort", 1}}
	aiConfig.Attacks = []Attack{{"AttackShort", skelemanDamage, 20}, {"AttackLong", skelemanDamage / 2, 40}}
	aiConfig.CombatOptions = []ai.WeightedState{{"Pursuit", 100}, {"Pace", 2}, {"Wait", 1}, {"RunAttack", 1}, {"AttackLong", 1}, {"AttackShort", 1}}

	g.Speed, g.Body.MaxX = ghoulSpeed, ghoulMaxSpeed
	g.SetDefaultAI(aiConfig)
}
