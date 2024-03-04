package entity

import (
	"game/comps/ai"
	"game/comps/basic/anim"
	"game/comps/basic/stats"
	"game/core"
	"game/entity/defaults"
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

type skeleman struct{ *defaults.Actor }

func NewSkeleman(x, y, _, _ float64, props *core.Property) *core.Entity {
	attackTags := []string{"AttackShort", "AttackLong"}
	skeleman := &skeleman{
		Actor: defaults.NewActor(x, y, skelemanWidth, skelemanHeight, attackTags),
	}
	skeleman.Anim = &anim.Comp{
		FilesName: skelemanAnimFile,
		OX:        skelemanOffsetX,
		OY:        skelemanOffsetY,
		OXFlip:    skelemanOffsetFlip,
		FlipX:     props.FlipX,
	}
	skeleman.Stats = &stats.Comp{MaxPoise: skelemanPoise}
	skeleman.Control.Speed = skelemanSpeed

	var view bump.Rect
	if props.View != nil {
		view = bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
	}
	skeleman.setupAI(view)

	skeleman.SetupComponents()
	skeleman.AddComponent(skeleman)

	return skeleman.Entity
}

func (s *skeleman) Init(_ *core.Entity) {
	hurtbox, err := s.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	s.Hitbox.PushHitbox(hurtbox, false)
}

func (s *skeleman) Update(dt float64) {
	s.Control.SimpleUpdate(dt, s.AI.Target)
}

func (s *skeleman) AttackJump(damage, stamina float64) {
	s.Control.Speed, s.Body.MaxX = skelemanSpeed, skelemanMaxSpeed*2
	go func() {
		time.Sleep(1 * time.Millisecond)
		s.Control.Attack("AttackShort", damage, stamina)
	}()
	s.Body.Vy = -s.Control.Speed
	if s.Anim.FlipX {
		s.Body.Vx += skelemanMaxSpeed * 2
	} else {
		s.Body.Vx -= skelemanMaxSpeed * 2
	}
}

func (s *skeleman) setupAI(view bump.Rect) {
	aiConfig := defaults.DefaultAIConfig()
	aiConfig.ViewRect = view
	aiConfig.PaceReact = []ai.WeightedState{{"AttackShort", 1}, {"Wait", 0}}
	jumpAttack := defaults.Attack{"AttackJump", skelemanDamage}
	aiConfig.Attacks = []defaults.Attack{
		{"AttackShort", skelemanDamage},
		{"AttackLong", skelemanDamage / 2},
		jumpAttack,
	}
	aiConfig.CombatOptions = []ai.WeightedState{
		{"Pursuit", 100},
		{"Pace", 2},
		{"Wait", 1},
		{"RunAttack", 1},
		{"AttackLong", 1},
		{"AttackShort", 1},
		{"AttackJump", 1},
	}

	s.Control.Speed, s.Body.MaxX = skelemanSpeed, skelemanMaxSpeed
	s.SetDefaultAI(aiConfig)

	s.AI.Fsm.SetAction(ai.State("AttackJump"), s.AnimBuilder("AttackShort", nil).
		SetCooldown(ai.Cooldown{1.5, 2.5}).
		SetEntry(func() { s.AttackJump(jumpAttack.Damage, 0) }).
		SetExit(func() { s.Body.MaxX = skelemanMaxSpeed }).
		Build())
}
