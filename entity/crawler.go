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
			OX:        crawlerOffsetX, OY: crawlerOffsetY,
			OXFlip: crawlerOffsetFlip,
			FlipX:  props.FlipX,
		},
		body:   &body.Comp{},
		hitbox: &hitbox.Comp{},
		stats:  &stats.Comp{MaxHealth: crawlerHealth, MaxPoise: crawlerDamage},
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

func (c *Crawler) Update(_ float64) {
	c.SimpleUpdate()
}

// nolint: nolintlint, gomnd
func (c *Crawler) aiScript(view *bump.Rect) {
	c.ai.Add(0, actor.IdleAction(c.Control, view))
	c.ai.Add(0, actor.ApproachAction(c.Control, crawlerSpeed, vars.DefaultMaxX))

	ai.Choice{
		{1, func() { c.ai.Add(5, actor.AttackAction(c.Control, "Attack", crawlerDamage)) }},
		{1, func() { c.ai.Add(1.5, actor.BackUpAction(c.Control, crawlerSpeed, 0)) }},
		{1, func() { c.ai.Add(1, actor.WaitAction()) }},
	}.Play()
}
