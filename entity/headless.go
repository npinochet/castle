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
	"strings"
)

const (
	headlessAnimFile                                     = "headless"
	headlessWidth, headlessHeight                        = 9, 12
	headlessOffsetX, headlessOffsetY, headlessOffsetFlip = -17, -11, 29

	headlessSpeed, headlessMaxSpeed = 100, 50
	headlessHealth                  = 100 //420
	headlessDamage                  = 18
	headlessExp                     = 500
	headlessPoise                   = 40

	headlessThrowFrame = 1
)

type Headless struct {
	*core.BaseEntity
	*actor.Control
	anim      *anim.Comp
	body      *body.Comp
	hitbox    *hitbox.Comp
	stats     *stats.Comp
	ai        *ai.Comp
	bareHands bool
}

func NewHeadless(x, y, _, _ float64, props *core.Properties) *Headless {
	headless := &Headless{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: headlessWidth, H: headlessHeight},
		anim: &anim.Comp{
			FilesName: headlessAnimFile,
			OX:        headlessOffsetX, OY: headlessOffsetY,
			OXFlip: headlessOffsetFlip,
			FlipX:  props.FlipX,
		},
		body:   &body.Comp{},
		hitbox: &hitbox.Comp{},
		stats:  &stats.Comp{MaxHealth: headlessHealth, MaxPoise: headlessPoise, Exp: headlessExp},
		ai:     &ai.Comp{},
	}
	headless.Add(headless.anim, headless.body, headless.hitbox, headless.stats, headless.ai)
	headless.Control = actor.NewControl(headless)

	var view *bump.Rect
	if props.View != nil {
		viewRect := bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
		view = &viewRect
	}
	headless.ai.SetAct(func() { headless.aiScript(view) })

	return headless
}

func (h *Headless) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return h.anim, h.body, h.hitbox, h.stats, h.ai
}

func (h *Headless) Update(dt float64) {
	h.SimpleUpdate(dt)
	if h.bareHands && h.anim.State != "Throw" && !strings.HasSuffix(h.anim.State, "Bare") {
		//TODO: Finish long and short punch animation and remove next line
		h.anim.State = strings.ReplaceAll(strings.ReplaceAll(h.anim.State, "Short", ""), "Long", "")
		h.anim.SetState(h.anim.State + "Bare")
	}
}

func (h *Headless) ThrowWeapon() {
	tag := "Throw"
	if h.anim.State == tag || h.PausingState() {
		return
	}
	h.anim.SetState(tag)
	h.anim.OnFrame(headlessThrowFrame, func() {
		vars.World.Add(NewHalberd(h.X-2, h.Y-4, h))
		h.bareHands = true
	})
}

//nolint:mnd
func (h *Headless) aiScript(view *bump.Rect) {
	shouldThrowWeapon := !h.bareHands && h.stats.Health < h.stats.MaxHealth/2
	h.ai.Add(0, actor.IdleAction(h.Control, view))
	if shouldThrowWeapon {
		h.ai.Add(0.1, actor.WaitAction())
		h.ai.Add(3, actor.AnimAction(h.Control, "Throw", func() { h.ThrowWeapon() }))

		return
	}
	h.ai.Add(0, actor.ApproachAction(h.Control, headlessSpeed, headlessMaxSpeed, 0))
	h.ai.Add(0.1, actor.WaitAction())

	ai.Choices{
		{2, func() { h.ai.Add(5, actor.AttackAction(h.Control, "AttackShort", headlessDamage)) }},
		{2, func() { h.ai.Add(5, actor.AttackAction(h.Control, "AttackLong", headlessDamage)) }},
		{1, func() { h.ai.Add(1, actor.BackUpAction(h.Control, headlessSpeed, 0)) }},
		{1, func() { h.ai.Add(0.8, actor.WaitAction()) }},
	}.Play()
}
