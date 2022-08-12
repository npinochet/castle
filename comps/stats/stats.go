package stats

import (
	"game/core"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	defaultHealth      = 100
	defaultStamina     = 100
	defaultPoise       = 100
	defaultRecoverRate = 20
)

func (c *Comp) IsActive() bool        { return c.active }
func (c *Comp) SetActive(active bool) { c.active = active }

type Comp struct {
	active                               bool
	MaxHealth, MaxStamina, MaxPoise      float64
	Health, Stamina, Poise               float64
	StaminaRecoverRate, PoiseRecoverRate float64
}

func (c *Comp) Init(entity *core.Entity) {
	if c.MaxHealth == 0 {
		c.MaxHealth, c.Health = defaultHealth, defaultHealth
	}
	if c.MaxStamina == 0 {
		c.MaxStamina, c.Stamina = defaultStamina, defaultStamina
	}
	if c.MaxPoise == 0 {
		c.MaxPoise, c.Poise = defaultPoise, defaultPoise
	}
	if c.StaminaRecoverRate == 0 {
		c.StaminaRecoverRate = defaultRecoverRate
	}
	if c.PoiseRecoverRate == 0 {
		c.PoiseRecoverRate = defaultRecoverRate
	}
}

func (c *Comp) Update(dt float64) {
	c.Stamina += c.StaminaRecoverRate * dt
	c.Poise += c.PoiseRecoverRate * dt
	c.Stamina = math.Min(c.Stamina, c.MaxStamina)
	c.Poise = math.Min(c.Poise, c.MaxPoise)
}

func (c *Comp) DebugDraw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {

}

func (c *Comp) HealthPercent() float64  { return c.Health / c.MaxHealth }
func (c *Comp) StaminaPercent() float64 { return c.Stamina / c.MaxStamina }
func (c *Comp) PoisePercent() float64   { return c.Poise / c.MaxPoise }

func (c *Comp) AddHealth(amount float64) {
	c.Health = math.Min(c.Health+amount, c.MaxHealth)
}
func (c *Comp) AddStamina(amount float64) {
	c.Stamina = math.Min(c.Stamina+amount, c.MaxStamina)
}
func (c *Comp) AddPoise(amount float64) {
	c.Poise = math.Min(c.Poise+amount, c.MaxPoise)
}
