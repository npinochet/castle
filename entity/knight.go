package entity

import (
	"game/comps/basic/anim"
	"game/comps/basic/stats"
	"game/core"
	"game/entity/defaults"
)

const knightAnimFile = "assets/knight"

type Knight struct{ *defaults.Actor }

func NewKnight(x, y, _, _ float64, props *core.Property) *core.Entity {
	speed := 100.0
	knight := &Knight{Actor: defaults.NewActor(x, y, playerWidth, playerHeight, []string{anim.AttackTag})}
	knight.Anim = &anim.Comp{FilesName: knightAnimFile, OX: playerOffsetX, OY: playerOffsetY, OXFlip: playerOffsetFlip, FlipX: props.FlipX}
	knight.Stats = &stats.Comp{MaxPoise: 25}
	knight.Control.Speed = speed
	knight.SetupComponents()
	knight.AddComponent(knight)

	knight.SetDefaultAI(nil)

	return knight.Entity
}

func (k *Knight) Init(_ *core.Entity) {
	hurtbox, err := k.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	k.Hitbox.PushHitbox(hurtbox, false)
}

func (k *Knight) Update(dt float64) {
	k.Control.SimpleUpdate(dt, k.AI.Target)
}
