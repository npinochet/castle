package stats

import (
	"fmt"
	"game/assets"
	"game/core"
	"game/utils"
	"game/vars"
	"image"
	"image/color"
	"math"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

const minAttackMultToShow = 0.1

var (
	hudImage, _, _    = ebitenutil.NewImageFromFileSystem(assets.FS, "hud.png")
	barEndImage, _    = hudImage.SubImage(image.Rect(vars.BarEndX1, 0, vars.BarEndX2, vars.BarH)).(*ebiten.Image)
	middleBarImage, _ = hudImage.SubImage(image.Rect(vars.MiddleBarX1, 0, vars.MiddleBarX2, vars.BarH)).(*ebiten.Image)
	healthColor       = color.RGBA{172, 50, 50, 255}
	staminaColor      = color.RGBA{55, 148, 110, 255}
	borderColor       = color.RGBA{34, 32, 52, 255}
	emptyColor        = color.RGBA{89, 86, 82, 255}
	lagColor          = color.RGBA{251, 242, 54, 255}

	fullEmptyBarImage  = ebiten.NewImage(1, vars.InnerBarH)
	fullBarImage       = ebiten.NewImage(1, vars.InnerBarH)
	fullLagBarImage    = ebiten.NewImage(1, vars.InnerBarH)
	fullAttackBarImage = ebiten.NewImage(1, vars.BarH)
	fullCountBar       = ebiten.NewImage(1, vars.BarH+2)
	headBar            = ebiten.NewImage(vars.EnemyBarW+2, vars.InnerBarH)
	headInnerBar       = ebiten.NewImage(vars.EnemyBarW, 1)
	headLagBar         = ebiten.NewImage(1, 1)
	headFillerBar      = ebiten.NewImage(1, 1)
	iconsImage         *ebiten.Image
)

var DebugDraw = false

func init() {
	fullEmptyBarImage.Fill(emptyColor)
	fullLagBarImage.Fill(lagColor)
	fullAttackBarImage.Fill(borderColor)
	fullCountBar.Fill(borderColor)
	iconsImage, _ = hudImage.SubImage(image.Rect(0, 0, vars.HudIconsX, hudImage.Bounds().Dy())).(*ebiten.Image)

	headBar.Fill(borderColor)
	headInnerBar.Fill(emptyColor)
	headLagBar.Fill(lagColor)
	headFillerBar.Fill(healthColor)
}

type Comp struct {
	Hud, Pause, NoDebug                                    bool
	MaxHealth, Health                                      float64
	MaxStamina, Stamina                                    float64
	MaxPoise, Poise                                        float64
	MaxHeal, Heal                                          int
	HealAmount                                             float64
	AttackMultPerHeal, AttackMult                          float64
	Exp                                                    int
	StaminaRecoverRate, PoiseRecoverSeconds                float64
	healthTween, staminaTween, poiseTween, attackMultTween *gween.Tween
	healthLag, staminaLag, poiseLag                        float64
	poiseTimer                                             *time.Timer
	entityW                                                float64
	headHealthTimer                                        float64
}

func (c *Comp) Init(entity core.Entity) {
	if c.MaxHealth == 0 {
		c.MaxHealth = vars.DefaultHealth
	}
	if c.MaxStamina == 0 {
		c.MaxStamina = vars.DefaultStamina
	}
	if c.MaxPoise == 0 {
		c.MaxPoise = vars.DefaultPoise
	}
	if c.MaxHeal == 0 {
		c.MaxHeal = vars.DefaultHeal
	}
	if c.HealAmount == 0 {
		c.HealAmount = vars.DefaultHealAmount
	}
	if c.AttackMultPerHeal == 0 {
		c.AttackMultPerHeal = vars.AttackMultPerHeal
	}
	if c.Health == 0 {
		c.Health = c.MaxHealth
	}
	if c.Stamina < c.MaxStamina {
		c.Stamina = c.MaxStamina
	}
	if c.Poise < c.MaxPoise {
		c.Poise = c.MaxPoise
	}
	if c.Heal < c.MaxHeal {
		c.Heal = c.MaxHeal
	}
	if c.StaminaRecoverRate == 0 {
		c.StaminaRecoverRate = vars.DefaultRecoverRate
	}
	if c.PoiseRecoverSeconds == 0 {
		c.PoiseRecoverSeconds = vars.DefaultRecoverSeconds
	}
	c.healthLag = c.Health
	c.staminaLag = c.Stamina
	c.poiseLag = c.Poise
	_, _, c.entityW, _ = entity.Rect()
}

func (c *Comp) Remove() {
	if c.poiseTimer != nil {
		c.poiseTimer.Stop()
	}
}

func (c *Comp) Update(dt float64) {
	c.headHealthTimer -= dt
	if c.healthTween != nil {
		if lag, done := c.healthTween.Update(float32(dt)); done {
			c.healthTween = nil
			c.healthLag = c.Health
		} else {
			c.healthLag = float64(lag)
			c.headHealthTimer = vars.HeadHealthShowSeconds
		}
	}

	if !c.Pause {
		c.Stamina += c.StaminaRecoverRate * dt
		c.Stamina = math.Min(c.Stamina, c.MaxStamina)
		c.Poise = math.Min(c.Poise, c.MaxPoise)
	} else if c.Hud {
		return
	}

	if c.staminaTween != nil {
		if lag, done := c.staminaTween.Update(float32(dt)); done {
			c.staminaTween = nil
			c.staminaLag = c.Stamina
		} else {
			c.staminaLag = float64(lag)
		}
	}
	if c.poiseTween != nil {
		if lag, done := c.poiseTween.Update(float32(dt)); done {
			c.poiseTween = nil
			c.poiseLag = c.Poise
		} else {
			c.poiseLag = float64(lag)
		}
	}
	if c.attackMultTween != nil {
		if attackMult, done := c.attackMultTween.Update(float32(dt)); done {
			c.attackMultTween = nil
		} else {
			c.AttackMult = float64(attackMult)
		}
	}
}

func (c *Comp) Draw(pipeline *core.Pipeline, entityPos ebiten.GeoM) {
	c.debugDraw(pipeline, entityPos)
	if c.Hud {
		c.drawHud(pipeline)

		return
	}
	if c.Health >= c.MaxHealth || c.Health <= 0 {
		return
	}
	if c.headHealthTimer <= 0 {
		return
	}
	c.drawHeadHealthBar(pipeline, entityPos, c.Health, c.MaxHealth, c.healthLag)
}

func (c *Comp) HealthPercent() float64  { return c.Health / c.MaxHealth }
func (c *Comp) StaminaPercent() float64 { return c.Stamina / c.MaxStamina }
func (c *Comp) PoisePercent() float64   { return c.Poise / c.MaxPoise }

func (c *Comp) AddHealth(amount float64) {
	if overHealth := c.Health - c.MaxHealth + amount; overHealth > 0 {
		attackMult := (overHealth / c.HealAmount) * c.AttackMultPerHeal
		if c.attackMultTween != nil {
			newAttackMult, _ := c.attackMultTween.Set(math.MaxFloat32)
			c.AttackMult = float64(newAttackMult)
		}
		c.attackMultTween = gween.New(float32(c.AttackMult), float32(c.AttackMult+attackMult), 1, ease.OutCubic)
	}
	c.healthLag = math.Max(c.Health, c.healthLag)
	c.Health = math.Min(c.Health+amount, c.MaxHealth)
	c.healthTween = gween.New(float32(c.healthLag), float32(c.Health), 1, ease.Linear)
}
func (c *Comp) AddStamina(amount float64) {
	c.staminaLag = c.Stamina
	c.Stamina = math.Min(c.Stamina+amount, c.MaxStamina)
	c.staminaTween = gween.New(float32(c.staminaLag), float32(c.Stamina), 1, ease.Linear)
}
func (c *Comp) AddPoise(amount float64) {
	c.poiseLag = math.Max(c.Poise, c.poiseLag)
	c.Poise = math.Min(c.Poise+amount, c.MaxPoise)
	c.poiseTween = gween.New(float32(c.poiseLag), float32(c.Poise), 1, ease.Linear)

	if amount < 0 {
		if c.poiseTimer != nil {
			c.poiseTimer.Stop()
		}
		timer := c.PoiseRecoverSeconds
		if c.Poise <= 0 {
			timer = c.PoiseRecoverSeconds / 10
		}
		c.poiseTimer = time.AfterFunc(time.Duration(float64(time.Second)*timer), func() {
			c.Poise = c.MaxPoise
			c.poiseLag = c.Poise
		})
	}
}
func (c *Comp) AddHeal(amount int) {
	if c.Heal += amount; c.Heal > c.MaxHeal {
		c.Heal = c.MaxHeal
	}
	if amount < 0 {
		healthAmount := float64(-amount) * c.HealAmount
		c.AddHealth(healthAmount)
	}
}
func (c *Comp) AddExp(amount int) {
	c.Exp += amount
}

func (c *Comp) drawHud(pipeline *core.Pipeline) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(1, 1)
	pipeline.Add(vars.PipelineScreenTag, vars.PipelineUILayer, func(screen *ebiten.Image) { screen.DrawImage(iconsImage, op) })
	pipeline.Add(vars.PipelineNormalMapTag, vars.PipelineUILayer, func(normalMap *ebiten.Image) {
		normalMap.DrawImage(iconsImage, &ebiten.DrawImageOptions{GeoM: op.GeoM, Blend: ebiten.BlendDestinationOut})
	})

	const staminaVisualScale = 0.8
	c.drawSegment(pipeline, op.GeoM, 0, c.Health, c.MaxHealth, c.healthLag, healthColor)
	c.drawSegment(
		pipeline, op.GeoM, 1, c.Stamina*staminaVisualScale, c.MaxStamina*staminaVisualScale, c.staminaLag*staminaVisualScale, staminaColor,
	)
	c.drawAttackMult(pipeline, op.GeoM)
	c.drawCount(pipeline, op.GeoM, 2, c.Heal, 0)
	c.drawCount(pipeline, op.GeoM, 3, c.Exp, 2)
}

func (c *Comp) drawSegment(pipeline *core.Pipeline, geoM ebiten.GeoM, y, current, max, lag float64, barColor color.Color) {
	normalOp := &ebiten.DrawImageOptions{Blend: ebiten.BlendDestinationOut}
	normalOp.GeoM.Scale(max+2, 1)
	normalOp.GeoM.Concat(geoM)
	normalOp.GeoM.Translate(vars.HudIconsX, vars.BarMiddleH*y)
	pipeline.Add(vars.PipelineNormalMapTag, vars.PipelineUILayer, func(normalMap *ebiten.Image) {
		normalMap.DrawImage(fullAttackBarImage, normalOp)
	})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(max, 1)
	op.GeoM.Concat(geoM)
	op.GeoM.Translate(vars.HudIconsX, vars.BarMiddleH*y)

	fillerGeoM := geoM
	fillerGeoM.Translate(vars.HudIconsX, vars.BarMiddleH*y+2)
	pipeline.Add(vars.PipelineScreenTag, vars.PipelineUILayer, func(screen *ebiten.Image) {
		screen.DrawImage(middleBarImage, op)

		op.GeoM.Reset()
		op.GeoM.Scale(max, 1)
		op.GeoM.Concat(fillerGeoM)
		screen.DrawImage(fullEmptyBarImage, op)
		if lag > 0 {
			op.GeoM.Reset()
			op.GeoM.Scale(lag+1, 1)
			op.GeoM.Concat(fillerGeoM)
			screen.DrawImage(fullLagBarImage, op)
		}
		if current > 0 {
			op.GeoM.Reset()
			op.GeoM.Scale(current+1, 1)
			op.GeoM.Concat(fillerGeoM)
			fullBarImage.Fill(barColor)
			screen.DrawImage(fullBarImage, op)
		}

		op.GeoM.Reset()
		op.GeoM.Concat(geoM)
		op.GeoM.Translate(vars.BarMiddleH+max, vars.BarMiddleH*y)
		screen.DrawImage(barEndImage, op)
	})
}

func (c *Comp) drawCount(pipeline *core.Pipeline, geoM ebiten.GeoM, y float64, count int, offset float64) {
	text := strconv.Itoa(count)
	w, _ := utils.TextSize(text, assets.NanoFont)
	op := &ebiten.DrawImageOptions{GeoM: geoM}
	op.GeoM.Translate(vars.HudIconsX, vars.BarMiddleH*y+offset)

	opBackground := &ebiten.DrawImageOptions{}
	opBackground.GeoM.Scale(float64(w)+2, 1)
	opBackground.GeoM.Concat(op.GeoM)
	pipeline.Add(vars.PipelineScreenTag, vars.PipelineUILayer, func(screen *ebiten.Image) {
		screen.DrawImage(fullCountBar, opBackground)
		op.GeoM.Translate(0, 2)
		utils.DrawText(screen, text, assets.NanoFont, op)
	})
	pipeline.Add(vars.PipelineNormalMapTag, vars.PipelineUILayer, func(normalMap *ebiten.Image) {
		normalMap.DrawImage(fullCountBar, &ebiten.DrawImageOptions{GeoM: opBackground.GeoM, Blend: ebiten.BlendDestinationOut})
	})
}

func (c *Comp) drawAttackMult(pipeline *core.Pipeline, geoM ebiten.GeoM) {
	if c.AttackMult < minAttackMultToShow {
		return
	}
	op := &ebiten.DrawImageOptions{GeoM: geoM}
	endImgW := float64(barEndImage.Bounds().Dx())
	op.GeoM.Translate(c.MaxHealth+vars.BarMiddleH+endImgW, 0)

	text := fmt.Sprintf("x%.1fATK", 1+c.AttackMult)
	w, _ := utils.TextSize(text, assets.NanoFont)
	opBackground := &ebiten.DrawImageOptions{}
	opBackground.GeoM.Scale(float64(w)+2, 1)
	opBackground.GeoM.Concat(op.GeoM)
	pipeline.Add(vars.PipelineScreenTag, vars.PipelineUILayer, func(screen *ebiten.Image) {
		screen.DrawImage(fullAttackBarImage, opBackground)
		op.GeoM.Translate(0, 1)
		utils.DrawText(screen, text, assets.NanoFont, op)
	})
	pipeline.Add(vars.PipelineNormalMapTag, vars.PipelineUILayer, func(normalMap *ebiten.Image) {
		normalMap.DrawImage(fullAttackBarImage, &ebiten.DrawImageOptions{GeoM: opBackground.GeoM, Blend: ebiten.BlendDestinationOut})
	})
}

func (c *Comp) drawHeadHealthBar(pipeline *core.Pipeline, entityPos ebiten.GeoM, current, max, lag float64) {
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	barW := float64(headBar.Bounds().Dx())
	op.GeoM.Translate((c.entityW-barW)/2, -7)
	normalOp := &ebiten.DrawImageOptions{GeoM: op.GeoM, Blend: ebiten.BlendDestinationOut}
	pipeline.Add(vars.PipelineNormalMapTag, vars.PipelineUILayer, func(normalMap *ebiten.Image) {
		normalMap.DrawImage(headBar, normalOp)
	})

	pipeline.Add(vars.PipelineScreenTag, vars.PipelineUILayer, func(screen *ebiten.Image) {
		screen.DrawImage(headBar, op)
		op.GeoM.Translate(1, 1)
		screen.DrawImage(headInnerBar, op)

		if width := math.Floor((lag / max) * vars.EnemyBarW); width > 0 {
			opScaler := &ebiten.DrawImageOptions{}
			opScaler.GeoM.Scale(width, 1)
			opScaler.GeoM.Concat(op.GeoM)
			screen.DrawImage(headLagBar, opScaler)
		}

		if width := math.Round((current / max) * vars.EnemyBarW); width > 0 {
			opScaler := &ebiten.DrawImageOptions{}
			opScaler.GeoM.Scale(width, 1)
			opScaler.GeoM.Concat(op.GeoM)
			screen.DrawImage(headFillerBar, opScaler)
		}
	})
}

func (c *Comp) debugDraw(pipeline *core.Pipeline, entityPos ebiten.GeoM) {
	if !DebugDraw || c.NoDebug {
		return
	}
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-5, -16)
	pipeline.Add(vars.PipelineScreenTag, vars.PipelineUILayer, func(screen *ebiten.Image) {
		utils.DrawText(screen, fmt.Sprintf("%0.2f/%0.2f/%0.2f", c.Health, c.Stamina, c.Poise), assets.NanoFont, op)
	})
}
