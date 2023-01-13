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
	gramDamage                               = 20
)

type gram struct {
	*Actor
}

func NewGram(x, y, w, h float64, props map[string]string) *core.Entity {
	animc := &anim.Comp{FilesName: gramAnimFile, OX: gramOffsetX, OY: gramOffsetY, OXFlip: gramOffsetFlip}
	animc.FlipX = props[core.HorizontalProp] == "true"

	body := &body.Comp{W: gramWidth, H: gramHeight, Unmovable: true}
	textbox := &textbox.Comp{
		Text: "Hewwo, I gram nice to mit yu, i have no idea wat i doing here, lol im so random, rawr",
		Body: body,
		Area: func() bump.Rect {
			rect := body.Rect()
			rect.X -= 10
			rect.W += 20

			return rect
		},
	}
	gram := &gram{Actor: NewActor(x, y, body, animc, nil, gramDamage, gramDamage)}
	gram.AddComponent(textbox, gram)

	return &gram.Entity
}

func (g *gram) Init(entity *core.Entity) {
	hurtbox, err := g.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	g.Hitbox.PushHitbox(hurtbox, false)
}
