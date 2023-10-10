package entity

import (
	"game/comps/ai"
	"game/comps/anim"
	"game/comps/stats"
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

type skeleman struct {
	*core.Entity
	ActorControl
}

func (s *skeleman) Tag() string { return "skeleman" }

func NewSkeleman(x, y, w, h float64, props *core.Property) *core.Entity {
	animc := &anim.Comp{FilesName: skelemanAnimFile, OX: skelemanOffsetX, OY: skelemanOffsetY, OXFlip: skelemanOffsetFlip}
	animc.FlipX = props.FlipX

	attackTags := []string{"AttackShort", "AttackLong"}
	skeleman := &skeleman{
		Entity: NewActorControl(x, y, skelemanWidth, skelemanHeight, attackTags, animc, nil, &stats.Comp{MaxPoise: skelemanPoise}),
	}
	skeleman.AddComponent(skeleman)
	skeleman.BindControl(skeleman.Entity)
	skeleman.Control.Speed = skelemanSpeed

	var view bump.Rect
	if props.View != nil {
		view = bump.NewRect(props.View.X, props.View.Y, props.View.Width, props.View.Height)
	}

	skeleman.setupAI(view)

	return skeleman.Entity
}

func (g *skeleman) Init(entity *core.Entity) {
	hurtbox, err := g.Anim.GetFrameHitbox(anim.HurtboxSliceName)
	if err != nil {
		panic("no hurtbox found")
	}
	g.Hitbox.PushHitbox(hurtbox, false)
}

func (g *skeleman) Update(dt float64) {
	g.Control.SimpleUpdate(dt, g.AI.Target)
}

func (g *skeleman) AttackJump(damage, stamina float64) {
	g.Control.Speed, g.Body.MaxX = skelemanSpeed, skelemanMaxSpeed*2
	go func() {
		time.Sleep(1 * time.Millisecond)
		g.Control.Attack("AttackShort", damage, stamina)
	}()
	g.Body.Vy = -g.Control.Speed
	if g.Anim.FlipX {
		g.Body.Vx += skelemanMaxSpeed * 2
	} else {
		g.Body.Vx -= skelemanMaxSpeed * 2
	}
}

func (g *skeleman) setupAI(view bump.Rect) {
	aiConfig := DefaultAIConfig()
	aiConfig.viewRect = view
	aiConfig.PaceReact = []ai.WeightedState{{"AttackShort", 1}, {"Wait", 0}}
	jumpAttack := Attack{"AttackJump", skelemanDamage, 30}
	aiConfig.Attacks = []Attack{
		{"AttackShort", skelemanDamage, 20},
		{"AttackLong", skelemanDamage / 2, 40},
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

	g.Control.Speed, g.Body.MaxX = skelemanSpeed, skelemanMaxSpeed
	g.SetDefaultAI(aiConfig)

	g.AI.Fsm.SetAction(ai.State("AttackJump"), g.AI.AnimBuilder("AttackShort", nil).
		SetCooldown(ai.Cooldown{1.5, 2.5}).
		AddCondition(g.AI.EnoughStamina(jumpAttack.StaminaDamage)).
		SetEntry(func() { g.AttackJump(jumpAttack.Damage, jumpAttack.StaminaDamage) }).
		SetExit(func() { g.Body.MaxX = skelemanMaxSpeed }).
		Build())
}
