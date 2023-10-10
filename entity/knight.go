package entity

import (
	"game/comps/anim"
	"game/comps/stats"
	"game/core"
)

const knightAnimFile = "assets/knight"

type Knight struct {
	*core.Entity
	ActorControl
}

func (k *Knight) Tag() string { return "Knight" }

func NewKnight(x, y, w, h float64, props *core.Property) *core.Entity {
	speed := 100.0
	animc := &anim.Comp{FilesName: knightAnimFile, OX: playerOffsetX, OY: playerOffsetY, OXFlip: playerOffsetFlip}
	animc.FlipX = props.FlipX

	knight := &Knight{
		Entity: NewActorControl(x, y, playerWidth, playerHeight, []string{anim.AttackTag}, animc, nil, &stats.Comp{MaxPoise: 25}),
	}
	knight.AddComponent(knight)
	knight.BindControl(knight.Entity)
	knight.Control.Speed = speed

	knight.SetDefaultAI(nil)

	return knight.Entity
}

func (k *Knight) Init(entity *core.Entity) {
	hurtbox, err := k.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	k.Hitbox.PushHitbox(hurtbox, false)
}

func (k *Knight) Update(dt float64) {
	k.Control.SimpleUpdate(dt, k.AI.Target)
}
