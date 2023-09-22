package entity

import (
	"game/actor"
	"game/core"
	"game/libs/bump"
)

const (
	crawlerAnimFile                                   = "assets/crawler"
	crawlerWidth, crawlerHeight                       = 11, 8
	crawlerOffsetX, crawlerOffsetY, crawlerOffsetFlip = -4, -4, 10
	crawlerSpeed                                      = 100
	crawlerHealth, crawlerDamage                      = 30, 20
)

type Crawler struct{ actor.Actor }

func NewCrawler(x, y, _, _ float64, props *core.Property) core.Entity {
	crawler := &Crawler{
		Actor: actor.NewActor(x, y, crawlerWidth, crawlerHeight, []string{"Attack"}),
	}
	crawler.Speed = crawlerSpeed
	crawler.Stats.MaxPoise = crawlerDamage
	crawler.Stats.MaxHealth = crawlerHealth
	crawler.Anim.FilesName = crawlerAnimFile
	crawler.Anim.OX, crawler.Anim.OY = crawlerOffsetX, crawlerOffsetY
	crawler.Anim.OXFlip = crawlerOffsetFlip
	crawler.Anim.FlipX = props.FlipX

	var view bump.Rect
	if props.View != nil {
		view = bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
	}

	if props.AI != "none" {
		crawler.setupAI(view)
	} else {
		crawler.Speed = 0
	}

	return crawler
}

func (c *Crawler) Update(dt float64) {
	c.Actor.Update(dt)
	c.BasicUpdate(dt)
}

func (c *Crawler) setupAI(view bump.Rect) {
	aiConfig := actor.DefaultAIConfig()
	aiConfig.ViewRect = view
	aiConfig.PaceReact = []actor.AIWeightedState{{"Attack", 1}, {"Wait", 0}}
	aiConfig.Attacks = []actor.Attack{{"Attack", crawlerDamage, 20}}
	aiConfig.CombatOptions = []actor.AIWeightedState{{"Pursuit", 100}, {"Pace", 2}, {"Wait", 1}, {"RunAttack", 1}, {"Attack", 1}}

	c.Speed, c.Body.MaxX = ghoulSpeed, ghoulMaxSpeed
	c.SetDefaultAI(aiConfig) // errors, must run after init()
}
