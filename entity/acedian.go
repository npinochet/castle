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
)

const (
	acedianAnimFile                                   = "acedian"
	acedianWidth, acedianHeight                       = 10, 18
	acedianOffsetX, acedianOffsetY, acedianOffsetFlip = -6, -2, 6
)

type Acedian struct {
	*core.BaseEntity
	*actor.Control
	anim   *anim.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	stats  *stats.Comp
	light  *shader.Light
}

func NewAcedian(x, y, _, _ float64, props *core.Properties) *Acedian {
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
		stats:  &stats.Comp{},
		light:  shader.AddLight(0, 0, 8),
	}
	text := "Hi Hello"
	if props.Custom["text"] != "" {
		text = props.Custom["text"]
	}
	textbox := &textbox.Comp{
		Text:      text,
		Indicator: true,
		Area: func() bump.Rect {
			return bump.NewRect(acedian.X-acedianWidth*2, acedian.Y-acedianHeight, acedianWidth*4, acedianHeight*2)
		},
	}
	acedian.Add(acedian.anim, acedian.body, acedian.hitbox, acedian.stats, textbox)
	acedian.Control = actor.NewControl(acedian)

	return acedian
}

func (a *Acedian) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return a.anim, a.body, a.hitbox, a.stats, nil
}

func (a *Acedian) Update(_ float64) {
	a.light.X, a.light.Y = a.X-4, a.Y+acedianHeight/2-1
}
