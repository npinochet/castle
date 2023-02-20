package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/stats"
	"game/core"
	"game/libs/bump"
)

const (
	crawlerAnimFile                                   = "assets/crawler"
	crawlerWidth, crawlerHeight                       = 11, 8
	crawlerOffsetX, crawlerOffsetY, crawlerOffsetFlip = -4, -4, 10
	crawlerSpeed                                      = 100
	crawlerDamage                                     = 20
)

type crawler struct {
	*Actor
}

func NewCrawler(x, y, w, h float64, props *core.Property) *core.Entity {
	animc := &anim.Comp{FilesName: crawlerAnimFile, OX: crawlerOffsetX, OY: crawlerOffsetY, OXFlip: crawlerOffsetFlip}
	animc.FlipX = props.FlipX

	crawler := &crawler{
		Actor: NewActor(x, y, crawlerWidth, crawlerHeight, []string{"Attack"}, animc, nil, &stats.Comp{MaxPoise: crawlerDamage}),
	}
	crawler.Speed = crawlerSpeed
	crawler.AddComponent(crawler)

	var view bump.Rect
	if props.View != nil {
		view = bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
	}
	crawler.setupAI(view)

	return &crawler.Entity
}

func (c *crawler) Init(entity *core.Entity) {
	hurtbox, err := c.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	c.Hitbox.PushHitbox(hurtbox, false)
}

func (c *crawler) Update(dt float64) {
	c.SimpleUpdate(dt)
}

func (c *crawler) setupAI(view bump.Rect) {
	aiConfig := DefaultAIConfig()
	aiConfig.viewRect = view
	aiConfig.PaceReact = []ai.WeightedState{{"Attack", 1}, {"Wait", 0}}
	aiConfig.Attacks = []Attack{{"Attack", crawlerDamage, 20}}
	aiConfig.CombatOptions = []ai.WeightedState{{"Pursuit", 100}, {"Pace", 2}, {"Wait", 1}, {"RunAttack", 1}, {"Attack", 1}}

	c.Speed, c.Body.MaxX = ghoulSpeed, ghoulMaxSpeed
	c.SetDefaultAI(aiConfig)
}