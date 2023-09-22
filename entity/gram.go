package entity

import (
	"game/actor"
	"game/core"
	"game/libs/bump"
)

const (
	gramAnimFile                             = "assets/gram"
	gramWidth, gramHeight                    = 10, 12
	gramOffsetX, gramOffsetY, gramOffsetFlip = -1, -2, 0
)

type Gram struct {
	actor.Actor
	actor.Textbox
}

func NewGram(x, y, _, _ float64, props *core.Property) core.Entity {
	gram := &Gram{
		Actor: actor.NewActor(x, y, gramWidth, gramHeight, nil),
	}
	gram.Body.Unmovable = true
	gram.Stats.MaxPoise, gram.Stats.Poise = 100, 100
	gram.Anim.FilesName = gramAnimFile
	gram.Anim.OX, gram.Anim.OY = gramOffsetX, gramOffsetY
	gram.Anim.OXFlip = gramOffsetFlip
	gram.Anim.FlipX = props.FlipX
	gram.Textbox = actor.Textbox{
		Text: "Hewwo, I Gramr nice to mit yu, i have no idea wat i doing here, lol im so random, rawr",
		Area: func() bump.Rect { return bump.NewRect(gram.X-10, gram.Y, gramWidth+20, gramHeight) },
	}

	return gram
}

func (g *Gram) Update(dt float64) {
	g.Actor.Update(dt)
	g.Textbox.Update(&g.Actor, dt)
}
