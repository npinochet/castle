package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/entity/actor"
	"game/libs/bump"
	"game/vars"
	"math/rand"
)

const (
	crawlerAnimFile                                   = "assets/crawler"
	crawlerWidth, crawlerHeight                       = 11, 8
	crawlerOffsetX, crawlerOffsetY, crawlerOffsetFlip = -4, -4, 10
	crawlerHealth, crawlerDamage                      = 30, 20
	crawlerSpeed                                      = 100
)

type Crawler struct {
	*core.BaseEntity
	*actor.Control
	anim   *anim.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	stats  *stats.Comp
	ai     *ai.Comp
}

func NewCrawler(x, y, _, _ float64, props *core.Properties) *Crawler {
	crawler := &Crawler{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: crawlerWidth, H: crawlerHeight},
		anim: &anim.Comp{
			FilesName: crawlerAnimFile,
			OX:        crawlerOffsetX,
			OY:        crawlerOffsetY,
			OXFlip:    crawlerOffsetFlip,
			FlipX:     props.FlipX,
		},
		body:   &body.Comp{},
		hitbox: &hitbox.Comp{},
		stats:  &stats.Comp{MaxPoise: crawlerDamage, MaxHealth: crawlerHealth},
		ai:     &ai.Comp{},
	}
	crawler.Add(crawler.anim, crawler.body, crawler.hitbox, crawler.stats, crawler.ai)
	crawler.Control = actor.NewControl(crawler)

	var view *bump.Rect
	if props.View != nil {
		viewRect := bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
		view = &viewRect
	}
	crawler.ai.SetAct(func() { crawler.aiScript(view) })

	return crawler
}

func (c *Crawler) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return c.anim, c.body, c.hitbox, c.stats, c.ai
}

func (c *Crawler) Init() {
	c.Control.Init()
}

func (c *Crawler) Update(_ float64) {
	if c.stats.Health <= 0 {
		c.Remove()

		return
	}
}

func (c *Crawler) aiScript(view *bump.Rect) {
	c.ai.Add(0, actor.IdleAction(c.Control, view))
	c.ai.Add(0, actor.ApproachAction(c.Control, crawlerSpeed, vars.DefaultMaxX))

	if fate := rand.Float64(); fate > 0.66 {
		c.ai.Add(5, actor.AttackAction(c.Control, "Attack", crawlerDamage))
	} else if fate > 0.33 {
		c.ai.Add(1.5, actor.BackUpAction(c.Control, crawlerSpeed, 0))
	} else {
		c.ai.Add(1, actor.WaitAction())
	}
}
