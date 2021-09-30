package comp

import (
	"game/core"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	defaultHealth      float64 = 100
	defaultStamina     float64 = 100
	defaultPoise       float64 = 100
	defaultRecoverRate float64 = 10
)

func (sc *StatsComponent) IsActive() bool        { return sc.active }
func (sc *StatsComponent) SetActive(active bool) { sc.active = active }

type StatsComponent struct {
	active                          bool
	MaxHealth, MaxStamina, MaxPoise float64
	Health, Stamina, Poise          float64
	StaminaRecoverRate              float64
	PoiseRecoverRate                float64
}

func (sc *StatsComponent) Init(entity *core.Entity) {
	if sc.MaxHealth == 0 {
		sc.MaxHealth, sc.Health = defaultHealth, defaultHealth
	}
	if sc.MaxStamina == 0 {
		sc.MaxStamina, sc.Stamina = defaultStamina, defaultStamina
	}
	if sc.MaxPoise == 0 {
		sc.MaxPoise, sc.Poise = defaultPoise, defaultPoise
	}
	if sc.StaminaRecoverRate == 0 {
		sc.StaminaRecoverRate = defaultRecoverRate
	}
	if sc.PoiseRecoverRate == 0 {
		sc.PoiseRecoverRate = defaultRecoverRate
	}
}

func (sc *StatsComponent) Update(dt float64) {
	sc.Stamina += sc.StaminaRecoverRate * dt
	sc.Poise += sc.PoiseRecoverRate * dt
	sc.Stamina = math.Min(sc.Stamina, sc.MaxStamina)
	sc.Poise = math.Min(sc.Poise, sc.MaxPoise)
}

func (sc *StatsComponent) DebugDraw(screen *ebiten.Image, enitiyPos ebiten.GeoM) {

}

func (sc *StatsComponent) AddHealth(amount float64) {
	sc.Health = math.Min(sc.Health+amount, sc.MaxHealth)
}
func (sc *StatsComponent) AddStamina(amount float64) {
	sc.Stamina = math.Min(sc.Stamina+amount, sc.MaxStamina)
}
func (sc *StatsComponent) AddPoise(amount float64) {
	sc.Poise = math.Min(sc.Poise+amount, sc.MaxPoise)
}
