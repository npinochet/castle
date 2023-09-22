package entity

import (
	"game/actor"
	"game/core"
	"game/libs/bump"
	"time"
)

const (
	skelemanAnimFile                                     = "assets/skeleman"
	skelemanWidth, skelemanHeight                        = 8, 12
	skelemanOffsetX, skelemanOffsetY, skelemanOffsetFlip = -12, -5, 20
	skelemanSpeed, skelemanMaxSpeed                      = 100, 35
	skelemanDamage                                       = 20
	skelemanPoise                                        = 30
)

type Skeleman struct {
	actor.Actor
}

func NewSkeleman(x, y, _, _ float64, props *core.Property) core.Entity {
	skeleman := &Skeleman{
		Actor: actor.NewActor(x, y, skelemanWidth, skelemanHeight, []string{"AttackShort", "AttackLong"}),
	}
	skeleman.Speed = skelemanSpeed
	skeleman.Stats.MaxPoise = skelemanPoise
	skeleman.Anim.FilesName = skelemanAnimFile
	skeleman.Anim.OX, skeleman.Anim.OY = skelemanOffsetX, skelemanOffsetY
	skeleman.Anim.OXFlip = skelemanOffsetFlip
	skeleman.Anim.FlipX = props.FlipX

	var view bump.Rect
	if props.View != nil {
		view = bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
	}

	skeleman.setupAI(view)

	return skeleman
}

func (s *Skeleman) Update(dt float64) {
	s.Actor.Update(dt)
	s.BasicUpdate(dt)
}

func (s *Skeleman) AttackJump(damage, stamina float64) {
	s.Speed, s.Body.MaxX = skelemanSpeed, skelemanMaxSpeed*2
	go func() {
		time.Sleep(1 * time.Millisecond)
		s.Attack("AttackShort", damage, stamina)
	}()
	s.Body.Vy = -s.Speed
	if s.Anim.FlipX {
		s.Body.Vx += skelemanMaxSpeed * 2
	} else {
		s.Body.Vx -= skelemanMaxSpeed * 2
	}
}

func (s *Skeleman) setupAI(view bump.Rect) {
	aiConfig := actor.DefaultAIConfig()
	aiConfig.ViewRect = view
	aiConfig.PaceReact = []actor.AIWeightedState{{"AttackShort", 1}, {"Wait", 0}}
	jumpAttack := actor.Attack{"AttackJump", skelemanDamage, 30}
	aiConfig.Attacks = []actor.Attack{
		{"AttackShort", skelemanDamage, 20},
		{"AttackLong", skelemanDamage / 2, 40},
		jumpAttack,
	}
	aiConfig.CombatOptions = []actor.AIWeightedState{
		{"Pursuit", 100},
		{"Pace", 2},
		{"Wait", 1},
		{"RunAttack", 1},
		{"AttackLong", 1},
		{"AttackShort", 1},
		{"AttackJump", 1},
	}

	s.Speed, s.Body.MaxX = skelemanSpeed, skelemanMaxSpeed
	s.SetDefaultAI(aiConfig)

	s.AI.Fsm.SetAction(actor.AIState("AttackJump"), s.AI.AnimBuilder("AttackShort", nil).
		SetCooldown(actor.AICooldown{1.5, 2.5}).
		AddCondition(s.AI.EnoughStamina(jumpAttack.StaminaDamage)).
		SetEntry(func() { s.AttackJump(jumpAttack.Damage, jumpAttack.StaminaDamage) }).
		SetExit(func() { s.Body.MaxX = skelemanMaxSpeed }).
		Build())
}
