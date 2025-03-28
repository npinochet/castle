package entity

import (
	"bytes"
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/gated"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/entity/actor"
	"game/libs/bump"
	"game/vars"
)

const (
	knightAnimFile                                 = "knight"
	knightWidth, knightHeight                      = 8, 11
	knightOffsetX, knightOffsetY, knightOffsetFlip = -10, -3, 17
	knightDamage                                   = 25
	knightHealth, knightPosie                      = 180, 25
	knightExp                                      = 50
)

type Knight struct {
	*core.BaseEntity
	*actor.Control
	anim   *anim.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	stats  *stats.Comp
	ai     *ai.Comp
	gates  *gated.Comp
}

func NewKnight(x, y, _, _ float64, props *core.Properties) *Knight {
	knight := &Knight{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: knightWidth, H: knightHeight},
		anim: &anim.Comp{
			FilesName: knightAnimFile,
			OX:        knightOffsetX, OY: knightOffsetY,
			OXFlip: knightOffsetFlip,
			FlipX:  props.FlipX,
		},
		body:   &body.Comp{},
		hitbox: &hitbox.Comp{},
		stats:  &stats.Comp{MaxHealth: knightHealth, MaxPoise: knightPosie, Exp: knightExp},
		ai:     &ai.Comp{},
		gates:  &gated.Comp{Props: props.Custom},
	}
	knight.Add(knight.anim, knight.body, knight.hitbox, knight.stats, knight.ai, knight.gates)
	knight.Control = actor.NewControl(knight)

	var view *bump.Rect
	if props.View != nil {
		viewRect := bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
		view = &viewRect
	}
	knight.ai.SetAct(func() { knight.aiScript(view) })

	return knight
}

func (k *Knight) Init() {
	k.Control.Init()
	k.turnImageRed()
}

func (k *Knight) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return k.anim, k.body, k.hitbox, k.stats, k.ai
}

func (k *Knight) Update(dt float64) {
	if k.stats.Health <= 0 {
		k.gates.Open()
	}
	k.SimpleUpdate(dt)
}

func (k *Knight) DashAction() *ai.Action {
	minDist := 20.0
	closeEnough := k.ai.InTargetRange(0, minDist)

	return &ai.Action{
		Name: "Dash",
		Next: func(_ float64) bool { return true },
		Entry: func() {
			if k.PausingState() {
				return
			}
			if !closeEnough {
				k.ai.Add(5, actor.AttackAction(k.Control, "Attack", knightDamage))
			} else {
				k.ai.Add(4, actor.AnimAction(k.Control, "Throw", func() { k.ThrowRock() }))
			}
			speed := k.body.MaxX * 4
			if !closeEnough != k.anim.FlipX {
				speed *= -1
			}
			k.body.Vx = speed
		},
	}
}

func (k *Knight) ThrowRock() {
	tag := "Attack"
	if k.anim.State == tag || k.PausingState() {
		return
	}
	k.anim.SetState(tag)
	k.anim.OnFrame(2, func() { vars.World.Add(NewRock(k.X-2, k.Y-4, k)) })
}

//nolint:mnd
func (k *Knight) aiScript(view *bump.Rect) {
	speed := 100.0
	maxSpeed := vars.DefaultMaxX
	secondPhase := k.stats.Health <= k.stats.MaxHealth*0.8
	if secondPhase {
		speed = 300.0
		maxSpeed = vars.DefaultMaxX * 3
		k.body.MaxX = maxSpeed
		k.anim.Data.PlaySpeed = 1.5
	}
	k.ai.Add(0, actor.IdleAction(k.Control, view))
	k.ai.Add(1, actor.EntryAction(func() { k.gates.Close() }))
	k.ai.Add(0, actor.ApproachAction(k.Control, speed, maxSpeed, 0))
	k.ai.Add(0.1, actor.WaitAction())

	choices := ai.Choices{
		{2, func() { k.ai.Add(5, actor.AttackAction(k.Control, "Attack", knightDamage)) }},
		{0.5, func() { k.ai.Add(1, actor.BackUpAction(k.Control, speed, 0)) }},
		{1, func() { k.ai.Add(0.2, actor.WaitAction()) }},
	}
	if secondPhase {
		choices = append(choices, ai.Choice{1, func() { k.ai.Add(1, actor.ShieldAction(k.Control)) }})
		choices = append(choices, ai.Choice{1, func() { k.ai.Add(1, k.DashAction()) }})
	}
	choices.Play()
}

func (k *Knight) turnImageRed() {
	image := k.anim.Image
	size := image.Bounds().Size()
	pixels := make([]byte, size.X*size.Y*4)
	image.ReadPixels(pixels)
	image.WritePixels(bytes.ReplaceAll(pixels, []byte{91, 110, 225}, []byte{172, 50, 50}))
}
