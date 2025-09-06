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
)

const (
	oscarAnimFile                               = "oscar"
	oscarWidth, oscarHeight                     = 7, 12
	oscarOffsetX, oscarOffsetY, oscarOffsetFlip = -3, -1, 6
)

type Oscar struct {
	*core.BaseEntity
	*actor.Control
	anim     *anim.Comp
	body     *body.Comp
	hitbox   *hitbox.Comp
	stats    *stats.Comp
	textbox  *textbox.Comp
	deadText string
}

func NewOscar(x, y, _, _ float64, props *core.Properties) *Oscar {
	text := "Hello"
	if props.Custom["text"] != "" {
		text = props.Custom["text"]
	}

	oscar := &Oscar{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: oscarWidth, H: oscarHeight},
		anim: &anim.Comp{
			FilesName: oscarAnimFile,
			OX:        oscarOffsetX, OY: oscarOffsetY,
			OXFlip: oscarOffsetFlip,
			FlipX:  props.FlipX,
		},
		textbox: &textbox.Comp{
			Text:      text,
			Indicator: true,
			Area: func() bump.Rect {
				return bump.NewRect(x-oscarWidth*2, y-oscarHeight, oscarWidth*4, oscarHeight*2)
			},
		},
		body:     &body.Comp{Unmovable: true, Solid: true},
		hitbox:   &hitbox.Comp{},
		stats:    &stats.Comp{MaxHealth: 200, Health: 80},
		deadText: props.Custom["deadText"],
	}

	oscar.stats.MaxPoise, oscar.stats.Poise = 100, 100
	oscar.Add(oscar.anim, oscar.body, oscar.hitbox, oscar.stats, oscar.textbox)
	oscar.Control = actor.NewControl(oscar)

	return oscar
}

func (g *Oscar) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return g.anim, g.body, g.hitbox, g.stats, nil
}

func (g *Oscar) Update(dt float64) {
	if g.stats.Health <= 0 {
		g.anim.SetState("Stagger")
		g.textbox.NewText(g.deadText)
		g.textbox.Indicator = false
	}
}
