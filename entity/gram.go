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
	gramAnimFile                             = "gram"
	gramWidth, gramHeight                    = 10, 12
	gramOffsetX, gramOffsetY, gramOffsetFlip = -1, -2, 0
)

type Gram struct {
	*core.BaseEntity
	*actor.Control
	anim   *anim.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	stats  *stats.Comp
}

func NewGram(x, y, _, _ float64, props *core.Properties) *Gram {
	gram := &Gram{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: gramWidth, H: gramHeight},
		anim:       &anim.Comp{FilesName: gramAnimFile, OX: gramOffsetX, OY: gramOffsetY, OXFlip: gramOffsetFlip, FlipX: props.FlipX},
		body:       &body.Comp{Unmovable: true},
		hitbox:     &hitbox.Comp{},
		stats:      &stats.Comp{},
	}
	textbox := &textbox.Comp{
		Text: "Hewwo, I Gramr nice to mit yu, i have no idea wat i doing here, lol im so random, rawr",
		Area: func() bump.Rect {
			return bump.NewRect(gram.X-gramWidth*2, gram.Y-gramHeight, gramWidth*4, gramHeight*2)
		},
		Indicator: true,
	}
	gram.stats.MaxPoise, gram.stats.Poise = 100, 100
	gram.Add(gram.anim, gram.body, gram.hitbox, gram.stats, textbox)
	gram.Control = actor.NewControl(gram)

	return gram
}

func (g *Gram) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return g.anim, g.body, g.hitbox, g.stats, nil
}

func (g *Gram) Update(_ float64) {}
