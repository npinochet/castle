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
	"time"
)

const (
	ghoulAnimFile                               = "ghoul"
	ghoulWidth, ghoulHeight                     = 9, 13
	ghoulOffsetX, ghoulOffsetY, ghoulOffsetFlip = -6.5, -4, 14
	ghoulSpeed, ghoulMaxSpeed                   = 100, 40
	ghoulHealth                                 = 70
	ghoulDamage, ghoulPoise                     = 18, 21
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
		stats:      &stats.Comp{MaxHealth: ghoulHealth, MaxPoise: ghoulPoise, Exp: ghoulExp},
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

func (g *Ghoul) Update(dt float64) { g.SimpleUpdate(dt) }

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

func (g *Ghoul) jumpAttackAction() *ai.Action {
	speed := float64(skelemanSpeed)

	return &ai.Action{
		Name: "JumpAttack",
		Entry: func() {
			if g.PausingState() {
				return
			}
			g.body.MaxX = ghoulMaxSpeed * 2
			time.AfterFunc(1*time.Millisecond, func() { g.Control.Attack("AttackShort", skelemanDamage, 0, 10, 10) })
			g.body.Vy = -skelemanSpeed / 2
			g.body.Ground = false
			if g.anim.FlipX {
				g.body.Vx += ghoulMaxSpeed * 2
			} else {
				g.body.Vx -= ghoulMaxSpeed * 2
				speed *= -1
			}
		},
		Next: func(dt float64) bool {
			if !g.PausingState() {
				g.body.Vx += speed * dt
			}

			return g.body.Ground && g.anim.State != "AttackShort"
		},
		Exit: func() { g.body.MaxX = ghoulMaxSpeed },
	}
}

//nolint:mnd
func (g *Ghoul) aiScript(view *bump.Rect, poacher bool) {
	g.ai.Add(0, actor.IdleAction(g.Control, view))
	if poacher && g.rocks > 0 {
		g.ai.Add(1, actor.BackUpAction(g.Control, ghoulSpeed, 0))
		g.ai.Add(3, actor.AnimAction(g.Control, "Throw", func() { g.ThrowRock() }))
		g.ai.Add(0.8, actor.WaitAction())

		return
	}
	ai.Choices{
		{0.2, func() {
			g.ai.Add(0, actor.ApproachAction(g.Control, ghoulSpeed, ghoulMaxSpeed, 10))
			g.ai.Add(0.1, actor.WaitAction())
			g.ai.Add(5, g.jumpAttackAction())
			g.ai.Add(0.2, actor.WaitAction())
		}},
		{1, func() {
			g.ai.Add(0, actor.ApproachAction(g.Control, ghoulSpeed, ghoulMaxSpeed, 0))
			g.ai.Add(0.1, actor.WaitAction())
			ai.Choices{
				{2, func() { g.ai.Add(5, actor.AttackAction(g.Control, "AttackShort", ghoulDamage)) }},
				{2, func() { g.ai.Add(5, actor.AttackAction(g.Control, "AttackLong", ghoulDamage)) }},
				{0.5, func() { g.ai.Add(1, actor.BackUpAction(g.Control, ghoulSpeed, 0)) }},
				{1, func() { g.ai.Add(0.1, actor.WaitAction()) }},
			}.Play()
		}},
	}.Play()
}
