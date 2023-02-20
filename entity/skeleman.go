package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/stats"
	"game/core"
	"game/libs/bump"
)

const (
	skelemanAnimFile                                     = "assets/skeleman"
	skelemanWidth, skelemanHeight                        = 8, 12
	skelemanOffsetX, skelemanOffsetY, skelemanOffsetFlip = -12, -5, 20
	skelemanSpeed                                        = 100
	skelemanDamage                                       = 20
	skelemanPoise                                        = 30
)

type skeleman struct {
	*Actor
}

func NewSkeleman(x, y, w, h float64, props *core.Property) *core.Entity {
	animc := &anim.Comp{FilesName: skelemanAnimFile, OX: skelemanOffsetX, OY: skelemanOffsetY, OXFlip: skelemanOffsetFlip}
	animc.FlipX = props.FlipX

	attackTags := []string{"AttackShort", "AttackLong"}
	skeleman := &skeleman{
		Actor: NewActor(x, y, skelemanWidth, skelemanHeight, attackTags, animc, nil, &stats.Comp{MaxPoise: skelemanPoise}),
	}
	skeleman.Speed = skelemanSpeed
	skeleman.AddComponent(skeleman)

	var view bump.Rect
	if props.View != nil {
		view = bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
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
	aiConfig.PaceReact = []ai.WeightedState{{"AttackShort", 1}, {"Wait", 0}}
	aiConfig.Attacks = []Attack{{"AttackShort", skelemanDamage, 20}, {"AttackLong", skelemanDamage / 2, 40}}
	aiConfig.CombatOptions = []ai.WeightedState{
		{"Pursuit", 100},
		{"Pace", 2},
		{"Wait", 1},
		{"RunAttack", 1},
		{"AttackLong", 1},
		{"AttackShort", 1},
	}

	g.Speed, g.Body.MaxX = ghoulSpeed, ghoulMaxSpeed
	g.SetDefaultAI(aiConfig)
}
