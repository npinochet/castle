package entity

import (
	"game/comps/ai"
	"game/comps/basic/anim"
	"game/comps/basic/stats"
	"game/core"
	"game/entity/defaults"
	"game/libs/bump"
)

const (
	crawlerAnimFile                                   = "assets/crawler"
	crawlerWidth, crawlerHeight                       = 11, 8
	crawlerOffsetX, crawlerOffsetY, crawlerOffsetFlip = -4, -4, 10
	crawlerSpeed                                      = 100
	crawlerHealth, crawlerDamage                      = 30, 20
)

type crawler struct{ *defaults.Actor }

func NewCrawler(x, y, w, h float64, props *core.Property) *core.Entity {
	crawler := &crawler{Actor: defaults.NewActor(x, y, crawlerWidth, crawlerHeight, []string{"Attack"})}
	crawler.Anim = &anim.Comp{
		FilesName: crawlerAnimFile,
		OX:        crawlerOffsetX,
		OY:        crawlerOffsetY,
		OXFlip:    crawlerOffsetFlip,
		FlipX:     props.FlipX,
	}
	crawler.Stats = &stats.Comp{MaxPoise: crawlerDamage, MaxHealth: crawlerHealth}
	crawler.Control.Speed = crawlerSpeed

	var view bump.Rect
	if props.View != nil {
		view = bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
	}
	if props.AI != "none" {
		crawler.setupAI(view)
	} else {
		crawler.Control.Speed = 0
	}
	crawler.SetupComponents()
	crawler.AddComponent(crawler)

	return crawler.Entity
}

func (c *crawler) Init(_ *core.Entity) {
	hurtbox, err := c.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	c.Hitbox.PushHitbox(hurtbox, false)
}

func (c *crawler) Update(dt float64) {
	c.Control.SimpleUpdate(dt, c.AI.Target)
}

func (c *crawler) setupAI(view bump.Rect) {
	config := defaults.DefaultAIConfig()
	config.ViewRect = view
	config.PaceReact = []ai.WeightedState{{"Attack", 1}, {"Wait", 0}}
	config.Attacks = []defaults.Attack{{"Attack", crawlerDamage}}
	config.CombatOptions = []ai.WeightedState{{"Pursuit", 100}, {"Pace", 2}, {"Wait", 1}, {"RunAttack", 1}, {"Attack", 1}}

	c.Control.Speed, c.Body.MaxX = ghoulSpeed, ghoulMaxSpeed
	c.SetDefaultAI(config)
}
