//nolint:revive, mnd
package vars

import (
	"game/core"
	"game/utils"
)

const (
	// Config.
	Scale                     = 4
	ScreenWidth, ScreenHeight = 160, 96 // 320, 240.
	Debug                     = debug

	// Pipeline Layers and Tags.
	PipelineUILayer      = 10
	PipelineScreenTag    = "screen"
	PipelineNormalMapTag = "normal"

	// Entity Flags.
	PlayerTeamFlag = iota
	EnemyTeamFlag

	// Anim.
	IdleTag       = "Idle"
	WalkTag       = "Walk"
	AttackTag     = "Attack"
	BlockTag      = "Block"
	ParryBlockTag = "ParryBlock"
	StaggerTag    = "Stagger"
	ClimbTag      = "Climb"
	ConsumeTag    = "Consume"

	HurtboxSliceName = "hurtbox"
	HitboxSliceName  = "hitbox"
	BlockSliceName   = "blockbox"
	HealSliceName    = "healbox"

	// Actor.
	DefaultAttackPushForce                    = 100
	DefaultReactForce                         = 50
	DefaultMaxXDiv, DefaultMaxXRecoverRateDiv = 2, 3

	// Stats.
	DefaultHealth         = 100
	DefaultStamina        = 80
	DefaultPoise          = 30
	DefaultHeal           = 2
	DefaultHealAmount     = 20
	AttackMultPerHeal     = 0.2
	DefaultRecoverRate    = 26
	DefaultRecoverSeconds = 3
	HeadHealthShowSeconds = 60

	// HUD consts.
	HudIconsX                = 7
	BarEndX1, BarEndX2, BarH = 8, 12, 7
	BarMiddleH               = BarH - 2
	MiddleBarX1, MiddleBarX2 = 7, 8
	InnerBarH                = 3
	EnemyBarW                = 10
	MaxTextWidth             = 50

	// Textbox.
	BoxX, DefaultBoxY            = 6.0, 30.0
	BoxMarginY, BoxMinY, BoxMaxY = 5, 25, 96 - BoxH - BoxMarginY
	BoxW, BoxH                   = ScreenWidth - BoxX*2, 3.0
	LineWidth, LineHeight        = (BoxW - 8), 6 + 1
	MaxLines                     = 4
)

var (
	// Global.
	World  *core.World
	Player core.Entity

	// Signaling.
	SaveGame, ResetGame bool

	// Player.
	Pad utils.ControlPack

	// Body.
	Gravity                     = 300.0
	DefaultMaxX, DefaultMaxY    = 20.0, 200.0
	GroundFriction, AirFriction = 8.0, 2.0 // TODO: Tune this variables. They might be too high.
	CollisionStiffness          = 1.0
	FrictionEpsilon             = 0.05
	CoyoteTimeSeconds           = 0.1
)
