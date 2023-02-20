package entity

import (
	"game/comps/anim"
	"game/comps/stats"
	"game/core"
)

const knightAnimFile = "assets/knight"

type Knight struct {
	*Actor
}

func NewKnight(x, y, w, h float64, props *core.Property) *core.Entity {
	speed := 100.0
	animc := &anim.Comp{FilesName: knightAnimFile, OX: playerOffsetX, OY: playerOffsetY, OXFlip: playerOffsetFlip}
	animc.FlipX = props.FlipX

	knight := &Knight{
		Actor: NewActor(x, y, playerWidth, playerHeight, []string{anim.AttackTag}, animc, nil, &stats.Comp{MaxPoise: 25}),
	}
	knight.Speed = speed
	knight.SetDefaultAI(nil)
	knight.AddComponent(knight)

	return &knight.Entity
}

func (k *Knight) Init(entity *core.Entity) {
	hurtbox, err := k.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	k.Hitbox.PushHitbox(hurtbox, false)
}

func (k *Knight) Update(dt float64) {
	k.SimpleUpdate(dt)
}
