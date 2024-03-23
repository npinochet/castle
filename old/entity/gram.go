package entity

import (
	"game/comps/basic/anim"
	"game/comps/basic/body"
	"game/comps/basic/textbox"
	"game/core"
	"game/entity/defaults"
	"game/libs/bump"
)

const (
	gramAnimFile                             = "assets/gram"
	gramWidth, gramHeight                    = 10, 12
	gramOffsetX, gramOffsetY, gramOffsetFlip = -1, -2, 0
)

type gram struct{ *defaults.Actor }

func NewGram(x, y, _, _ float64, props *core.Property) *core.Entity {
	gram := &gram{Actor: defaults.NewActor(x, y, gramWidth, gramHeight, nil)}
	gram.Anim = &anim.Comp{FilesName: gramAnimFile, OX: gramOffsetX, OY: gramOffsetY, OXFlip: gramOffsetFlip, FlipX: props.FlipX}
	gram.Body = &body.Comp{Unmovable: true}
	textbox := &textbox.Comp{
		Text: "Hewwo, I Gramr nice to mit yu, i have no idea wat i doing here, lol im so random, rawr",
		Body: gram.Body,
		Area: func() bump.Rect { return bump.NewRect(gram.X-10, gram.Y, gramWidth+20, gramHeight) },
	}
	gram.Stats.MaxPoise, gram.Stats.Poise = 100, 100
	gram.SetupComponents()
	gram.AddComponent(textbox, gram)

	return gram.Entity
}

func (g *gram) Init(entity *core.Entity) {
	hurtbox, err := g.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	g.Hitbox.PushHitbox(hurtbox, false)
}
