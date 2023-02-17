package entity

import (
	"game/comps/anim"
	"game/comps/body"
	"game/comps/textbox"
	"game/core"
	"game/libs/bump"
)

const (
	gramAnimFile                             = "assets/gram"
	gramWidth, gramHeight                    = 10, 12
	gramOffsetX, gramOffsetY, gramOffsetFlip = -1, -2, 0
)

type gram struct {
	*Actor
}

func NewGram(x, y, w, h float64, props map[string]string) *core.Entity {
	animc := &anim.Comp{FilesName: gramAnimFile, OX: gramOffsetX, OY: gramOffsetY, OXFlip: gramOffsetFlip}
	animc.FlipX = props[core.HorizontalProp] == "true"

	body := &body.Comp{Unmovable: true}
	gram := &gram{Actor: NewActor(x, y, gramWidth, gramHeight, body, animc, nil)}
	textbox := &textbox.Comp{
		Text: "Hewwo, I Gramr nice to mit yu, i have no idea wat i doing here, lol im so random, rawr",
		Body: body,
		Area: func() bump.Rect {
			return bump.NewRect(gram.X-10, gram.Y, gramWidth+20, gramHeight)
		},
	}
	gram.AddComponent(textbox, gram)
	gram.Stats.MaxPoise, gram.Stats.Poise = 100, 100

	return &gram.Entity
}

func (g *gram) Init(entity *core.Entity) {
	hurtbox, err := g.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	g.Hitbox.PushHitbox(hurtbox, false)
}
