package entity

import (
	"game/comps/anim"
	"game/comps/body"
	"game/comps/stats"
	"game/core"
)

const (
	ghoulAnimFile                               = "assets/ghoul"
	ghoulWidth, ghoulHeight                     = 9, 13
	ghoulOffsetX, ghoulOffsetY, ghoulOffsetFlip = -6.5, -4, 14
	ghoulMaxSpeed                               = 20
	ghoulDamage                                 = 20
)

type ghoul struct {
	*Actor
}

func Newghoul(x, y, w, h float64, props map[string]string) *core.Entity {
	speed := 100.0
	anim := &anim.Comp{FilesName: ghoulAnimFile, OX: ghoulOffsetX, OY: ghoulOffsetY, OXFlip: ghoulOffsetFlip}
	body := &body.Comp{W: ghoulWidth, H: ghoulHeight, MaxX: ghoulMaxSpeed}

	ghoul := &ghoul{
		Actor: NewActor(x, y, body, anim, &stats.Comp{MaxPoise: ghoulDamage}, ghoulDamage, ghoulDamage),
	}
	ghoul.speed = speed
	ghoul.AI = ghoul.NewDefaultAI(nil)
	ghoul.AddComponent(ghoul.AI)
	ghoul.AddComponent(ghoul)

	return &ghoul.Entity
}

func (k *ghoul) Init(entity *core.Entity) {
	hurtbox, err := k.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	k.Hitbox.PushHitbox(hurtbox, false)
}

func (k *ghoul) Update(dt float64) {
	k.ManageAnim()
	if k.Anim.State == anim.WalkTag && k.speed == 0 {
		k.Anim.SetState(anim.IdleTag)
	}
	if k.AI.Target != nil {
		if k.Anim.State == anim.WalkTag || k.Anim.State == anim.IdleTag {
			k.Anim.FlipX = k.AI.Target.X > k.X
		}
	}
	if k.Anim.State != anim.AttackTag && k.Anim.State != anim.StaggerTag {
		if k.Anim.FlipX {
			k.Body.Vx += k.speed * dt
		} else {
			k.Body.Vx -= k.speed * dt
		}
	}

	if k.Stats.Health <= 0 {
		k.World.RemoveEntity(k.ID) // TODO: creates infinite/recursive loop sometimes I think.
	}
}
