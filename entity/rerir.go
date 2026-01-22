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
	rerirAnimFile                               = "rerir"
	rerirWidth, rerirHeight                     = 10, 12
	rerirOffsetX, rerirOffsetY, rerirOffsetFlip = -1, -2, 6 // TODO: Finish
)

type Rerir struct {
	*core.BaseEntity
	*actor.Control
	anim   *anim.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	stats  *stats.Comp
}

func NewRerir(x, y, _, _ float64, props *core.Properties) *Rerir {
	rerir := &Rerir{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: rerirWidth, H: rerirHeight},
		anim:       &anim.Comp{FilesName: rerirAnimFile, OX: rerirOffsetX, OY: rerirOffsetY, OXFlip: rerirOffsetFlip, FlipX: props.FlipX},
		body:       &body.Comp{Unmovable: true},
		hitbox:     &hitbox.Comp{},
		stats:      &stats.Comp{},
	}
	text := "Hello, I'm Gaiseric, nice to meet you"
	if props.Custom["text"] != "" {
		text = props.Custom["text"]
	}
	textbox := &textbox.Comp{
		Text:      text,
		Indicator: true,
		Area: func() bump.Rect {
			return bump.NewRect(rerir.X-rerirWidth*2, rerir.Y-rerirHeight, rerirWidth*4, rerirHeight*2)
		},
	}
	rerir.stats.MaxPoise, rerir.stats.Poise = 100, 100
	rerir.Add(rerir.anim, rerir.body, rerir.hitbox, rerir.stats, textbox)
	rerir.Control = actor.NewControl(rerir)

	return rerir
}

func (g *Rerir) Comps() (anim *anim.Comp, body *body.Comp, hitbox *hitbox.Comp, stats *stats.Comp, ai *ai.Comp) {
	return g.anim, g.body, g.hitbox, g.stats, nil
}

func (g *Rerir) Update(_ float64) {}
