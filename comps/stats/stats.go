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

var (
	hudImage, _, _    = ebitenutil.NewImageFromFile("assets/hud.png")
	barEndImage, _    = hudImage.SubImage(image.Rect(vars.BarEndX1, 0, vars.BarEndX2, vars.BarH)).(*ebiten.Image)
	middleBarImage, _ = hudImage.SubImage(image.Rect(vars.MiddleBarX1, 0, vars.MiddleBarX2, vars.BarH)).(*ebiten.Image)
	healthColor       = color.RGBA{172, 50, 50, 255}
	staminaColor      = color.RGBA{55, 148, 110, 255}
	//poiseColor        = color.RGBA{91, 110, 225, 255}
	borderColor = color.RGBA{34, 32, 52, 255}
	emptyColor  = color.RGBA{89, 86, 82, 255}
	lagColor    = color.RGBA{251, 242, 54, 255}
)

var DebugDraw = false

type Comp struct {
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
}

func (c *Comp) Init(_ core.Entity) {
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
		c.StaminaRecoverRate = vars.DefaultRecoverRate
	}
	if c.PoiseRecoverSeconds == 0 {
		c.PoiseRecoverSeconds = vars.DefaultRecoverSeconds
	}
	c.healthLag = c.Health
	c.staminaLag = c.Stamina
	c.poiseLag = c.Poise
}

func (c *Comp) Update(dt float64) {
	if c.healthTween != nil {
		if lag, done := c.healthTween.Update(float32(dt)); done {
			c.healthTween = nil
			c.healthLag = c.Health
		} else {
			c.healthLag = float64(lag)
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
}

func (c *Comp) Draw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	c.debugDraw(screen, entityPos)
	if c.Hud {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(1, 1)
		screen.DrawImage(c.drawHud(), op)

		return
	}
	if c.Health >= c.MaxHealth {
		return
	}

	healthBar := c.headBarImage(c.Health, c.MaxHealth, c.healthLag, healthColor)
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-2, -10)
	screen.DrawImage(healthBar, op)
}

func (c *Comp) HealthPercent() float64  { return c.Health / c.MaxHealth }
func (c *Comp) StaminaPercent() float64 { return c.Stamina / c.MaxStamina }
func (c *Comp) PoisePercent() float64   { return c.Poise / c.MaxPoise }

func (c *Comp) AddHealth(amount float64) {
	c.healthLag = math.Max(c.Health, c.healthLag)
	c.Health = math.Min(c.Health+amount, c.MaxHealth)
	c.healthTween = gween.New(float32(c.healthLag), float32(c.Health), 1, ease.Linear)
}
func (c *Comp) AddStamina(amount float64) {
	c.staminaLag = c.Stamina
	//c.staminaLag = math.Max(c.Stamina, c.staminaLag)
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
}
func (c *Comp) AddExp(amount int) {
	c.Exp += amount
}

func (c *Comp) drawHud() *ebiten.Image {
	h := hudImage.Bounds().Dy()
	w, _ := ebiten.WindowSize()
	hud := ebiten.NewImage(w, h)
	icons, _ := hudImage.SubImage(image.Rect(0, 0, vars.HudIconsX, h)).(*ebiten.Image)
	hud.DrawImage(icons, nil)

	c.drawSegment(hud, 0, c.Health, c.MaxHealth, c.healthLag, healthColor)
	c.drawSegment(hud, 1, c.Stamina, c.MaxStamina, c.staminaLag, staminaColor)
	// c.drawSegment(hud, 2, c.Poise, c.MaxPoise, c.poiseLag, poiseColor)
	c.drawCount(hud, 2, 128, 0) // c.Exp, 0) //nolint: gomnd
	c.drawCount(hud, 3, c.Heal, 2)

	return hud
}

func (c *Comp) drawSegment(hud *ebiten.Image, y, current, max, lag float64, barColor color.Color) {
	fullBar := ebiten.NewImage(int(max), vars.InnerBarH)
	fullBar.Fill(emptyColor)
	if lag > 0 {
		bar := ebiten.NewImage(int(lag+1), vars.InnerBarH)
		bar.Fill(lagColor)
		fullBar.DrawImage(bar, nil)
	}
	if current > 0 {
		bar := ebiten.NewImage(int(current+1), vars.InnerBarH)
		bar.Fill(barColor)
		fullBar.DrawImage(bar, nil)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(max, 1)
	op.GeoM.Translate(vars.HudIconsX, vars.BarMiddleH*y)
	hud.DrawImage(middleBarImage, op)

	op.GeoM.Reset()
	op.GeoM.Translate(vars.HudIconsX, vars.BarMiddleH*y+2)
	hud.DrawImage(fullBar, op)

	op.GeoM.Reset()
	op.GeoM.Translate(vars.BarMiddleH+max, vars.BarMiddleH*y)
	hud.DrawImage(barEndImage, op)
}

func (c *Comp) drawCount(hud *ebiten.Image, y float64, count int, offset float64) {
	fullBar := ebiten.NewImage(vars.MaxTextWidth, vars.BarH+2)
	fullBar.Fill(borderColor)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 2)
	w, _ := utils.DrawText(fullBar, strconv.Itoa(count), assets.TinyFont, op)
	fullBar, _ = fullBar.SubImage(image.Rect(0, 0, w+2, vars.BarH+2)).(*ebiten.Image)

	op.GeoM.Reset()
	op.GeoM.Translate(0, -2)
	op.GeoM.Translate(vars.HudIconsX, vars.BarMiddleH*y+2+offset)
	hud.DrawImage(fullBar, op)
}

func (c *Comp) headBarImage(current, max, lag float64, barColor color.Color) *ebiten.Image {
	image := ebiten.NewImage(vars.EnemyBarW+2, vars.InnerBarH)
	image.Fill(borderColor)
	fullBar := ebiten.NewImage(vars.EnemyBarW, 1)
	fullBar.Fill(emptyColor)

	round := 0.5
	if width := int((lag/max)*vars.EnemyBarW + round); width > 0 {
		bar := ebiten.NewImage(width, 1)
		bar.Fill(lagColor)
		fullBar.DrawImage(bar, nil)
	}

	if width := int((current/max)*vars.EnemyBarW + round); width > 0 {
		bar := ebiten.NewImage(width, 1)
		bar.Fill(barColor)
		fullBar.DrawImage(bar, nil)
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(1, 1)
	image.DrawImage(fullBar, op)

	return image
}

func (c *Comp) debugDraw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	if !DebugDraw || c.NoDebug {
		return
	}
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-5, -16)
	utils.DrawText(screen, fmt.Sprintf("%0.2f/%0.2f/%0.2f", c.Health, c.Stamina, c.Poise), assets.TinyFont, op)
}
