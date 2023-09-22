package entity

import (
	"game/actor"
	"game/core"
)

const knightAnimFile = "assets/knight"

type Knight struct{ actor.Actor }

func NewKnight(x, y, _, _ float64, props *core.Property) core.Entity {
	knight := &Knight{
		Actor: actor.NewActor(x, y, playerWidth, playerHeight, []string{actor.AttackTag}),
	}
	knight.Speed = 100
	knight.Stats.MaxPoise = 25
	knight.Anim.FilesName = knightAnimFile
	knight.Anim.OX, knight.Anim.OY = playerOffsetX, playerOffsetY
	knight.Anim.OXFlip = playerOffsetFlip
	knight.Anim.FlipX = props.FlipX
	knight.SetDefaultAI(nil)

	return knight
}

func (k *Knight) Update(dt float64) {
	k.Actor.Update(dt)
	k.BasicUpdate(dt)
}
