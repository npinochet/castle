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
	anim := &anim.Comp{FilesName: knightAnimFile, OX: playerOffsetX, OY: playerOffsetY, OXFlip: playerOffsetFlip}
	body := &body.Comp{W: playerWidth, H: playerHeight, MaxX: 35}

	knight := &Knight{
		Actor: NewActor(x, y, body, anim, &stats.Comp{MaxPoise: 25}, 20, 20),
	}
	knight.speed = speed
	knight.SetDefaultAI(nil, nil)
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
	k.ManageAnim([]string{anim.AttackTag})
	if k.Anim.State == anim.WalkTag && k.speed == 0 {
		k.Anim.SetState(anim.IdleTag)
	}
	if k.AI.Target != nil {
		if k.Anim.State == anim.WalkTag || k.Anim.State == anim.IdleTag {
			k.Anim.FlipX = k.AI.Target.X < k.X
		}
	}
	if k.Anim.State != anim.AttackTag && k.Anim.State != anim.StaggerTag {
		if k.Anim.FlipX {
			k.Body.Vx -= k.speed * dt
		} else {
			k.Body.Vx += k.speed * dt
		}
	}

	if k.Stats.Health <= 0 {
		k.Remove()
	}
}
