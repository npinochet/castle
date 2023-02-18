package entity

import (
	"game/comps/anim"
	"game/comps/body"
	"game/comps/stats"
	"game/core"
)

const knightAnimFile = "assets/knight"

type Knight struct {
	*Actor
}

func NewKnight(x, y, w, h float64, props map[string]string) *core.Entity {
	speed := 100.0
	animc := &anim.Comp{FilesName: knightAnimFile, OX: playerOffsetX, OY: playerOffsetY, OXFlip: playerOffsetFlip}
	animc.FlipX = props[core.HorizontalProp] == "true"
	body := &body.Comp{MaxX: 35}

	knight := &Knight{
		Actor: NewActor(x, y, playerWidth, playerHeight, body, animc, &stats.Comp{MaxPoise: 25}, []string{anim.AttackTag}),
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
