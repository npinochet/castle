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
	"strconv"
)

const (
	ghoulAnimFile                               = "ghoul"
	ghoulWidth, ghoulHeight                     = 9, 13
	ghoulOffsetX, ghoulOffsetY, ghoulOffsetFlip = -6.5, -4, 14
	ghoulSpeed, ghoulMaxSpeed                   = 100, 30
	ghoulHealth                                 = 70
	ghoulDamage                                 = 18
	ghoulExp                                    = 20
	ghoulThrowFrame                             = 2
)

type Ghoul struct {
	*core.BaseEntity
	*actor.Control
	anim   *anim.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	stats  *stats.Comp
	ai     *ai.Comp
	rocks  int
}

func NewGhoul(x, y, _, _ float64, props *core.Properties) *Ghoul {
	rocks, _ := strconv.Atoi(props.Custom["rocks"])
	ghoul := &Ghoul{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: ghoulWidth, H: ghoulHeight},
		anim:       &anim.Comp{FilesName: ghoulAnimFile, OX: ghoulOffsetX, OY: ghoulOffsetY, OXFlip: ghoulOffsetFlip, FlipX: props.FlipX},
		body:       &body.Comp{MaxX: ghoulMaxSpeed},
		hitbox:     &hitbox.Comp{},
		stats:      &stats.Comp{MaxPoise: ghoulDamage, MaxHealth: ghoulHealth, Exp: ghoulExp},
		ai:         &ai.Comp{},
		rocks:      rocks,
	}
	ghoul.Add(ghoul.anim, ghoul.body, ghoul.hitbox, ghoul.stats, ghoul.ai)
	ghoul.Control = actor.NewControl(ghoul)

	var view *bump.Rect
	if props.View != nil {
		viewRect := bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
		view = &viewRect
	}
	ghoul.ai.SetAct(func() { ghoul.aiScript(view, props.Custom["ai"] == "poacher") })

	return ghoul
}

func (g *Ghoul) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return g.anim, g.body, g.hitbox, g.stats, g.ai
}

func (g *Ghoul) Update(dt float64) {
	g.SimpleUpdate(dt)
}

func (g *Ghoul) ThrowRock() {
	tag := "Throw"
	if g.anim.State == tag || g.PausingState() || g.rocks <= 0 {
		return
	}
	g.anim.SetState(tag)
	g.anim.OnFrame(ghoulThrowFrame, func() {
		vars.World.Add(NewRock(g.X-2, g.Y-4, g))
		g.rocks--
	})
}

// nolint: nolintlint, gomnd
func (g *Ghoul) aiScript(view *bump.Rect, poacher bool) {
	g.ai.Add(0, actor.IdleAction(g.Control, view))
	if poacher && g.rocks > 0 {
		g.ai.Add(1, actor.BackUpAction(g.Control, ghoulSpeed, 0))
		g.ai.Add(3, actor.AnimAction(g.Control, "Throw", func() { g.ThrowRock() }))

		return
	}
	g.ai.Add(0, actor.ApproachAction(g.Control, ghoulSpeed, vars.DefaultMaxX))
	g.ai.Add(0.1, actor.WaitAction())

	ai.Choices{
		{2, func() { g.ai.Add(5, actor.AttackAction(g.Control, "AttackShort", ghoulDamage)) }},
		{2, func() { g.ai.Add(5, actor.AttackAction(g.Control, "AttackLong", ghoulDamage)) }},
		{0.5, func() { g.ai.Add(1, actor.BackUpAction(g.Control, ghoulSpeed, 0)) }},
		{1, func() { g.ai.Add(1, actor.WaitAction()) }},
	}.Play()
}
