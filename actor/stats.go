package actor

import (
	"fmt"
	"game/assets"
	"game/utils"
	"image"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

const (
	stDefaultHealth         = 100
	stDefaultStamina        = 80
	stDefaultPoise          = 30
	stDefaultHeal           = 3
	stDefaultRecoverRate    = 20
	stDefaultRecoverSeconds = 3

	// HUD consts.
	stHudIconsX                    = 7
	stBarEndX1, stBarEndX2, stBarH = 8, 12, 7
	stBarMiddleH                   = stBarH - 2
	stMiddleBarX1, stMiddleBarX2   = 7, 8
	stInnerBarH                    = 3
	stEnemyBarW                    = 10
	stMaxTextWidth                 = 50
)

var (
	stHudImage, _, _    = ebitenutil.NewImageFromFile("assets/hud.png")
	stBarEndImage, _    = stHudImage.SubImage(image.Rect(stBarEndX1, 0, stBarEndX2, stBarH)).(*ebiten.Image)
	stMiddleBarImage, _ = stHudImage.SubImage(image.Rect(stMiddleBarX1, 0, stMiddleBarX2, stBarH)).(*ebiten.Image)
	stHealthColor       = color.RGBA{172, 50, 50, 255}
	stStaminaColor      = color.RGBA{55, 148, 110, 255}
	stPoiseColor        = color.RGBA{91, 110, 225, 255}
	stBorderColor       = color.RGBA{34, 32, 52, 255}
	stEmptyColor        = color.RGBA{89, 86, 82, 255}
	stLagColor          = color.RGBA{251, 242, 54, 255}
)

var StDebugDraw = false

type Stats struct {
	Hud, Pause, NoDebug                     bool
	MaxHealth, Health                       float64
	MaxStamina, Stamina                     float64
	MaxPoise, Poise                         float64
	MaxHeal, Heal                           int
	Exp                                     int
	StaminaRecoverRate, PoiseRecoverSeconds float64
	healthTween, staminaTween, poiseTween   *gween.Tween
	healthLag, staminaLag, poiseLag         float64
	poiseTimer                              *time.Timer
	StaminaRecoverRateMultiplier            float64
}

func (c *Stats) Init() {
	if c.MaxHealth == 0 {
		c.MaxHealth = stDefaultHealth
	}
	if c.MaxStamina == 0 {
		c.MaxStamina = stDefaultStamina
	}
	if c.MaxPoise == 0 {
		c.MaxPoise = stDefaultPoise
	}
	if c.MaxHeal == 0 {
		c.MaxHeal = stDefaultHeal
	}
	if c.Health < c.MaxHealth {
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
		c.StaminaRecoverRate = stDefaultRecoverRate
	}
	if c.PoiseRecoverSeconds == 0 {
		c.PoiseRecoverSeconds = stDefaultRecoverSeconds
	}
	c.healthLag = c.Health
	c.staminaLag = c.Stamina
	c.poiseLag = c.Poise
}

func (c *Stats) Update(dt float64) {
	if c.healthTween != nil {
		if lag, done := c.healthTween.Update(float32(dt)); done {
			c.healthTween = nil
			c.healthLag = c.Health
		} else {
			c.healthLag = float64(lag)
		}
	}

	if !c.Pause {
		c.Stamina += c.StaminaRecoverRate * (c.StaminaRecoverRateMultiplier + 1) * dt
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
}

func (c *Stats) Draw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	if StDebugDraw {
		c.debugDraw(screen, entityPos)
	}
	if c.Hud {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(1, 1)
		screen.DrawImage(c.drawHud(), op)

		return
	}
	if c.Health >= c.MaxHealth {
		return
	}

	healthBar := c.headBarImage(c.Health, c.MaxHealth, c.healthLag, stHealthColor)
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-2, -10)
	screen.DrawImage(healthBar, op)
}

func (c *Stats) HealthPercent() float64  { return c.Health / c.MaxHealth }
func (c *Stats) StaminaPercent() float64 { return c.Stamina / c.MaxStamina }
func (c *Stats) PoisePercent() float64   { return c.Poise / c.MaxPoise }

func (c *Stats) AddHealth(amount float64) {
	c.healthLag = math.Max(c.Health, c.healthLag)
	c.Health = math.Min(c.Health+amount, c.MaxHealth)
	c.healthTween = gween.New(float32(c.healthLag), float32(c.Health), 1, ease.Linear)
}
func (c *Stats) AddStamina(amount float64) {
	c.staminaLag = c.Stamina
	// c.staminaLag = math.Max(c.Stamina, c.staminaLag)
	c.Stamina = math.Min(c.Stamina+amount, c.MaxStamina)
	c.staminaTween = gween.New(float32(c.staminaLag), float32(c.Stamina), 1, ease.Linear)
}
func (c *Stats) AddPoise(amount float64) {
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
func (c *Stats) AddHeal(amount int) {
	if c.Heal += amount; c.Heal > c.MaxHeal {
		c.Heal = c.MaxHeal
	}
}
func (c *Stats) AddExp(amount int) {
	c.Exp += amount
}

func (c *Stats) drawHud() *ebiten.Image {
	_, h := stHudImage.Size()
	w, _ := ebiten.WindowSize()
	hud := ebiten.NewImage(w, h)
	icons, _ := stHudImage.SubImage(image.Rect(0, 0, stHudIconsX, h)).(*ebiten.Image)
	hud.DrawImage(icons, nil)

	c.drawSegment(hud, 0, c.Health, c.MaxHealth, c.healthLag, stHealthColor)
	c.drawSegment(hud, 1, c.Stamina, c.MaxStamina, c.staminaLag, stStaminaColor)
	// c.drawSegment(hud, 2, c.Poise, c.MaxPoise, c.poiseLag, stPoiseColor)
	c.drawCount(hud, 2, 128, 0) // c.Exp, 0) //nolint: gomnd
	c.drawCount(hud, 3, c.Heal, 2)

	return hud
}

func (c *Stats) drawSegment(hud *ebiten.Image, y, current, max, lag float64, barColor color.Color) {
	fullBar := ebiten.NewImage(int(max), stInnerBarH)
	fullBar.Fill(stEmptyColor)
	if lag > 0 {
		bar := ebiten.NewImage(int(lag+1), stInnerBarH)
		bar.Fill(stLagColor)
		fullBar.DrawImage(bar, nil)
	}
	if current > 0 {
		bar := ebiten.NewImage(int(current+1), stInnerBarH)
		bar.Fill(barColor)
		fullBar.DrawImage(bar, nil)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(max, 1)
	op.GeoM.Translate(stHudIconsX, stBarMiddleH*y)
	hud.DrawImage(stMiddleBarImage, op)

	op.GeoM.Reset()
	op.GeoM.Translate(stHudIconsX, stBarMiddleH*y+2)
	hud.DrawImage(fullBar, op)

	op.GeoM.Reset()
	op.GeoM.Translate(stBarMiddleH+max, stBarMiddleH*y)
	hud.DrawImage(stBarEndImage, op)
}

func (c *Stats) drawCount(hud *ebiten.Image, y float64, count int, offset float64) {
	fullBar := ebiten.NewImage(stMaxTextWidth, stBarH+2)
	fullBar.Fill(stBorderColor)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 2)
	w, _ := utils.DrawText(fullBar, fmt.Sprint(count), assets.TinyFont, op)
	fullBar, _ = fullBar.SubImage(image.Rect(0, 0, w+2, stBarH+2)).(*ebiten.Image)

	op.GeoM.Reset()
	op.GeoM.Translate(0, -2)
	op.GeoM.Translate(stHudIconsX, stBarMiddleH*y+2+offset)
	hud.DrawImage(fullBar, op)
}

func (c *Stats) headBarImage(current, max, lag float64, barColor color.Color) *ebiten.Image {
	image := ebiten.NewImage(stEnemyBarW+2, stInnerBarH)
	image.Fill(stBorderColor)
	fullBar := ebiten.NewImage(stEnemyBarW, 1)
	fullBar.Fill(stEmptyColor)

	round := 0.5
	if width := int((lag/max)*stEnemyBarW + round); width > 0 {
		bar := ebiten.NewImage(width, 1)
		bar.Fill(stLagColor)
		fullBar.DrawImage(bar, nil)
	}

	if width := int((current/max)*stEnemyBarW + round); width > 0 {
		bar := ebiten.NewImage(width, 1)
		bar.Fill(barColor)
		fullBar.DrawImage(bar, nil)
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(1, 1)
	image.DrawImage(fullBar, op)

	return image
}

func (c *Stats) debugDraw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	if c.NoDebug {
		return
	}
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-5, -16)
	utils.DrawText(screen, fmt.Sprintf("%0.2f/%0.2f/%0.2f", c.Health, c.Stamina, c.Poise), assets.TinyFont, op)
}
