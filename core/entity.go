package core

import (
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
)

type Component any

type Initializer interface{ Init(entity *Entity) }
type Updater interface{ Update(dt float64) }
type Drawer interface {
	Draw(screen *ebiten.Image, entityPos ebiten.GeoM)
}

type Entity struct {
	ID          uint64
	X, Y        float64
	H, W        float64
	World       *World
	flags       uint64 // TODO: Flags system for entites.
	components  map[string]Component
	orderedTags []string
}

func (e *Entity) Position() (float64, float64)               { return e.X, e.Y }
func (e *Entity) Rect() (float64, float64, float64, float64) { return e.X, e.Y, e.W, e.H }

func (e *Entity) Add(adding ...Component) {
	if e.components == nil {
		e.components = map[string]Component{}
	}
	for _, c := range adding {
		tag := reflect.TypeOf(c).String()
		if e.components[tag] == nil {
			e.orderedTags = append(e.orderedTags, tag)
		}
		e.components[tag] = c
	}
}

func (e *Entity) AddWithTag(component Component, tag string) (Component, bool) {
	if e.components[tag] != nil {
		return e.components[tag], false
	}
	e.components[tag] = component
	e.orderedTags = append(e.orderedTags, tag)

	return component, true
}

// func (e *Entity) GetComponent[T Component](tag string) Component { return e.Components[tag] }.
func Get[T Component](entity *Entity) T {
	var t T
	tag := reflect.TypeOf(t).String()
	comp, _ := entity.components[tag].(T)

	return comp
}

func GetWithTag[T Component](entity *Entity, tag string) T {
	comp, _ := entity.components[tag].(T)

	return comp
}
