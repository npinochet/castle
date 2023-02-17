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
		Actor: NewActor(x, y, crawlerWidth, crawlerHeight, body, animc, &stats.Comp{MaxPoise: crawlerDamage}),
	}
	crawler.speed = crawlerSpeed
	crawler.AddComponent(crawler)

	var view bump.Rect
	if viewStrings := strings.Split(props[core.ViewProp], ","); len(viewStrings) > 1 {
		view.X, _ = strconv.ParseFloat(viewStrings[0], 64)
		view.Y, _ = strconv.ParseFloat(viewStrings[1], 64)
		view.W, _ = strconv.ParseFloat(viewStrings[2], 64)
		view.H, _ = strconv.ParseFloat(viewStrings[3], 64)
	}

	crawler.setupAI(view)
	crawler.speed = 0

	return &crawler.Entity
}

func (g *crawler) Init(entity *core.Entity) {
	hurtbox, err := g.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	g.Hitbox.PushHitbox(hurtbox, false)
}

func (g *crawler) Update(dt float64) {
	g.ManageAnim([]string{"Attack"})
	if g.Anim.State == anim.WalkTag && g.speed == 0 {
		g.Anim.SetState(anim.IdleTag)
	}
	if g.AI.Target != nil {
		if g.Anim.State == anim.WalkTag || g.Anim.State == anim.IdleTag {
			g.Anim.FlipX = g.AI.Target.X > g.X
		}
	}
	if g.Anim.State != "Attack" && g.Anim.State != anim.StaggerTag {
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

func (g *crawler) setupAI(view bump.Rect) {
	aiConfig := DefaultAIConfig()
	aiConfig.viewRect = view
	aiConfig.PaceReact = []ai.WeightedState{{"Attack", 1}}
	aiConfig.Attacks = []Attack{{"Attack", crawlerDamage, 20}}
	aiConfig.CombatOptions = []ai.WeightedState{{"Pursuit", 100}, {"Pace", 2}, {"Wait", 1}, {"RunAttack", 1}, {"Attack", 1}}

	g.speed, g.Body.MaxX = ghoulSpeed, ghoulMaxSpeed
	g.SetDefaultAI(aiConfig)
}
