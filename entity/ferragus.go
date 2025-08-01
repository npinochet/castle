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
	ferragusAnimFile                                     = "ferragus"
	ferragusWidth, ferragusHeight                        = 8, 15
	ferragusOffsetX, ferragusOffsetY, ferragusOffsetFlip = -2, -1, 6
)

type Ferragus struct {
	*core.BaseEntity
	*actor.Control
	anim   *anim.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	stats  *stats.Comp
}

func NewFerragus(x, y, _, _ float64, props *core.Properties) *Ferragus {
	ferragus := &Ferragus{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: ferragusWidth, H: ferragusHeight},
		anim: &anim.Comp{
			FilesName: ferragusAnimFile,
			OX:        ferragusOffsetX, OY: ferragusOffsetY,
			OXFlip: ferragusOffsetFlip,
			FlipX:  props.FlipX,
		},
		body:   &body.Comp{Unmovable: true},
		hitbox: &hitbox.Comp{},
		stats:  &stats.Comp{},
	}
	text := "Hello"
	if props.Custom["text"] != "" {
		text = props.Custom["text"]
	}
	textbox := &textbox.Comp{
		Text:      text,
		Indicator: true,
		Area: func() bump.Rect {
			return bump.NewRect(ferragus.X-ferragusWidth*2, ferragus.Y-ferragusHeight, ferragusWidth*4, ferragusHeight*2)
		},
	}
	ferragus.stats.MaxPoise, ferragus.stats.Poise = 100, 100
	ferragus.Add(ferragus.anim, ferragus.body, ferragus.hitbox, ferragus.stats, textbox)
	ferragus.Control = actor.NewControl(ferragus)

	return ferragus
}

func (g *Ferragus) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return g.anim, g.body, g.hitbox, g.stats, nil
}

func (g *Ferragus) Update(_ float64) {}
