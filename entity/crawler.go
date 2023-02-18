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
	crawlerAnimFile                                   = "assets/crawler"
	crawlerWidth, crawlerHeight                       = 11, 8
	crawlerOffsetX, crawlerOffsetY, crawlerOffsetFlip = -4, -4, 10
	crawlerSpeed                                      = 100
	crawlerMaxSpeed                                   = 20
	crawlerDamage                                     = 20
)

type crawler struct {
	*Actor
}

func NewCrawler(x, y, w, h float64, props map[string]string) *core.Entity {
	animc := &anim.Comp{FilesName: crawlerAnimFile, OX: crawlerOffsetX, OY: crawlerOffsetY, OXFlip: crawlerOffsetFlip}
	animc.FlipX = props[core.HorizontalProp] == "true"

	body := &body.Comp{MaxX: crawlerMaxSpeed}

	crawler := &crawler{
		Actor: NewActor(x, y, crawlerWidth, crawlerHeight, body, animc, &stats.Comp{MaxPoise: crawlerDamage}, []string{"Attack"}),
	}
	crawler.Speed = crawlerSpeed
	crawler.AddComponent(crawler)

	var view bump.Rect
	if viewStrings := strings.Split(props[core.ViewProp], ","); len(viewStrings) > 1 {
		view.X, _ = strconv.ParseFloat(viewStrings[0], 64)
		view.Y, _ = strconv.ParseFloat(viewStrings[1], 64)
		view.W, _ = strconv.ParseFloat(viewStrings[2], 64)
		view.H, _ = strconv.ParseFloat(viewStrings[3], 64)
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
	aiConfig.PaceReact = []ai.WeightedState{{"Attack", 1}}
	aiConfig.Attacks = []Attack{{"Attack", crawlerDamage, 20}}
	aiConfig.CombatOptions = []ai.WeightedState{{"Pursuit", 100}, {"Pace", 2}, {"Wait", 1}, {"RunAttack", 1}, {"Attack", 1}}

	c.Speed, c.Body.MaxX = ghoulSpeed, ghoulMaxSpeed
	c.SetDefaultAI(aiConfig)
}
