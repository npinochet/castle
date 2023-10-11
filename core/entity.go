package core

import (
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
)

type Component any // Or use any with tag = reflect.TypeOf((var x T)).String().

type Initializer interface{ Init(*Entity) }
type Updater interface{ Update(dt float64) }
type Drawer interface {
	Draw(screen *ebiten.Image, entityPos ebiten.GeoM)
}

type Entity struct {
	ID          uint64
	World       *World
	X, Y, W, H  float64
	Components  map[string]Component
	orderedTags []string
}

func (e *Entity) Position() (float64, float64)               { return e.X, e.Y }
func (e *Entity) Rect() (float64, float64, float64, float64) { return e.X, e.Y, e.W, e.H }

func (e *Entity) AddComponent(components ...Component) Component {
	if e.Components == nil {
		e.Components = map[string]Component{}
	}
	for _, c := range components {
		tag := reflect.TypeOf(c).String()
		if e.Components[tag] == nil {
			e.orderedTags = append(e.orderedTags, tag)
		}
		e.Components[tag] = c
	}

	return components[0]
}

func (e *Entity) AddComponentWithTag(component Component, tag string) (Component, bool) {
	if e.Components[tag] != nil {
		return e.Components[tag], false
	}
	e.Components[tag] = component
	e.orderedTags = append(e.orderedTags, tag)

	return component, true
}

// func (e *Entity) GetComponent[T Component](tag string) Component { return e.Components[tag] }.
func GetComponent[T Component](entity *Entity) T {
	var t T
	tag := reflect.TypeOf(t).String()
	comp, _ := entity.Components[tag].(T)

	return comp
}
func GetComponentWithTag[T Component](entity *Entity, tag string) T {
	comp, _ := entity.Components[tag].(T)

	return comp
}

func (e *Entity) InitComponents() {
	for _, tag := range e.orderedTags {
		if initializer, ok := e.Components[tag].(Initializer); ok {
			initializer.Init(e)
		}
	}
}

func (e *Entity) UpdateComponents(dt float64) {
	for _, tag := range e.orderedTags {
		if updater, ok := e.Components[tag].(Updater); ok {
			updater.Update(dt)
		}
	}
}

func (e *Entity) Draw(screen *ebiten.Image) {
	entityPos := ebiten.GeoM{}
	entityPos.Translate(e.X, e.Y)
	x, y := e.World.Camera.Position()
	entityPos.Translate(-x, -y)

	for _, tag := range e.orderedTags {
		if drawer, ok := e.Components[tag].(Drawer); ok {
			drawer.Draw(screen, entityPos)
		}
	}
}
