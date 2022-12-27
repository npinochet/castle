package entity

import (
	"game/comps/ai"
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

func NewGhoul(x, y, w, h float64, props map[string]string) *core.Entity {
	speed := 100.0
	animc := &anim.Comp{FilesName: ghoulAnimFile, OX: ghoulOffsetX, OY: ghoulOffsetY, OXFlip: ghoulOffsetFlip}
	body := &body.Comp{W: ghoulWidth, H: ghoulHeight, MaxX: ghoulMaxSpeed}

	ghoul := &ghoul{
		Actor: NewActor(x, y, body, animc, &stats.Comp{MaxPoise: ghoulDamage}, ghoulDamage, ghoulDamage),
	}
	ghoul.speed = speed
	combatOptions := ghoul.SetDefaultAI(nil)
	ghoul.AddComponent(ghoul)

	combatOptions = append(combatOptions, ai.WeightedState{"Attack1", 0.125})
	combatOptions = append(combatOptions, ai.WeightedState{"Attack2", 0.125})
	// Remove Attack from combatOptions
	ghoul.AI.SetCombatOptions(combatOptions)
	ghoul.AI.Fsm.SetAction("Attack", nil)
	ghoul.AI.Fsm.SetAction("Attack1", ghoul.AI.AnimBuilder("Attack1", nil).
		SetCooldown(ai.Cooldown{2, 3}).
		AddCondition(ghoul.AI.EnoughStamina(0.2)).
		SetEntry(func() { ghoul.Attack("Attack1") }).
		Build())
	ghoul.AI.Fsm.SetAction("Attack2", ghoul.AI.AnimBuilder("Attack2", nil).
		SetCooldown(ai.Cooldown{2, 3}).
		AddCondition(ghoul.AI.EnoughStamina(0.2)).
		SetEntry(func() { ghoul.Attack("Attack2") }).
		Build())
	ghoul.Anim.Fsm.Transitions["Attack1"] = anim.IdleTag
	ghoul.Anim.Fsm.Transitions["Attack2"] = anim.IdleTag

	return &ghoul.Entity
}

func (g *ghoul) Init(entity *core.Entity) {
	hurtbox, err := g.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	g.Hitbox.PushHitbox(hurtbox, false)
}

func (g *ghoul) Update(dt float64) {
	g.ManageAnim([]string{"Attack1", "Attack2"})
	if g.Anim.State == anim.WalkTag && g.speed == 0 {
		g.Anim.SetState(anim.IdleTag)
	}
	if g.AI.Target != nil {
		if g.Anim.State == anim.WalkTag || g.Anim.State == anim.IdleTag {
			g.Anim.FlipX = g.AI.Target.X > g.X
		}
	}
	if g.Anim.State != "Attack1" && g.Anim.State != anim.StaggerTag {
		if g.Anim.FlipX {
			g.Body.Vx += g.speed * dt
		} else {
			g.Body.Vx -= g.speed * dt
		}
	}

	if g.Stats.Health <= 0 {
		g.World.RemoveEntity(g.ID) // TODO: creates infinite/recursive loop sometimes I think.
	}
}
