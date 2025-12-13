package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/stats"
	"game/comps/textbox"
	"game/core"
	"game/entity/actor"
	"game/libs/bump"
	"game/shader"
	"game/vars"
)

const (
	acedianAnimFile                                   = "acedian"
	acedianWidth, acedianHeight                       = 10, 18
	acedianOffsetX, acedianOffsetY, acedianOffsetFlip = -6, -2, 6
)

type Acedian struct {
	*core.BaseEntity
	*actor.Control
	anim                *anim.Comp
	body                *body.Comp
	hitbox              *hitbox.Comp
	stats               *stats.Comp
	light               *shader.Light
	textbox             *textbox.Comp
	hurtText, finalText string
}

func NewAcedian(x, y, _, _ float64, props *core.Properties) *Acedian {
	text := "Hi Hello"
	if props.Custom["text"] != "" {
		text = props.Custom["text"]
	}
	acedian := &Acedian{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: acedianWidth, H: acedianHeight},
		anim: &anim.Comp{
			FilesName: acedianAnimFile,
			OX:        acedianOffsetX, OY: acedianOffsetY,
			OXFlip: acedianOffsetFlip,
			FlipX:  props.FlipX,
		},
		body:   &body.Comp{Unmovable: true},
		hitbox: &hitbox.Comp{},
		stats:  &stats.Comp{MaxPoise: 500},
		light:  shader.AddLight(0, 0, 8),
		textbox: &textbox.Comp{
			Text:      text,
			Indicator: true,
		},
		hurtText:  props.Custom["hurtText"],
		finalText: props.Custom["finalText"],
	}
	acedian.textbox.Area = func() bump.Rect {
		return bump.NewRect(acedian.X-acedianWidth*2, acedian.Y-acedianHeight, acedianWidth*4, acedianHeight*2)
	}
	acedian.Add(acedian.anim, acedian.body, acedian.hitbox, acedian.stats, acedian.textbox)
	acedian.Control = actor.NewControl(acedian)

	return acedian
}

func (a *Acedian) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return a.anim, a.body, a.hitbox, a.stats, nil
}

func (a *Acedian) Update(dt float64) {
	a.light.X, a.light.Y = a.X-4, a.Y+acedianHeight/2-1
	healthPerc := a.stats.Health / a.stats.MaxHealth
	if a.hurtText != "" && healthPerc < 0.8 {
		a.textbox.NewText(a.hurtText)
	}
	if a.finalText != "" && healthPerc < 0.3 {
		a.textbox.NewText(a.finalText)
		a.light.Y -= 4
		a.light.X += 2
		a.Die(dt, 2.5)
	}
}

func (a *Acedian) Destroy() {
	a.light.Remove()
	if healthPerc := a.stats.Health / a.stats.MaxHealth; healthPerc < 0.3 {
		// TODO: Choose an safe offscreen spwan point
		vars.World.Add(NewVarg(a.X, a.Y-20, 0, 0, nil))
	}
}
