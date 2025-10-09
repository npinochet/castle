package core

import "reflect"

type BaseEntity struct {
	X, Y, H, W     float64
	components     []Component
	tagToComponent map[string]int
}

func (e *BaseEntity) Components() []Component                    { return e.components }
func (e *BaseEntity) Position() (float64, float64)               { return e.X, e.Y }
func (e *BaseEntity) SetPosition(x, y float64)                   { e.X = x; e.Y = y }
func (e *BaseEntity) Rect() (float64, float64, float64, float64) { return e.X, e.Y, e.W, e.H }
func (e *BaseEntity) SetSize(w, h float64)                       { e.W = w; e.H = h }

func (e *BaseEntity) Add(adding ...Component) {
	if e.tagToComponent == nil {
		e.tagToComponent = map[string]int{}
	}
	for _, c := range adding {
		tag := reflect.TypeOf(c).String()
		e.tagToComponent[tag] = len(e.components)
		e.components = append(e.components, c)
	}
}

func (e *BaseEntity) AddWithTag(component Component, tag string) (Component, bool) {
	if e.components[e.tagToComponent[tag]] != nil {
		return e.components[e.tagToComponent[tag]], false
	}
	e.tagToComponent[tag] = len(e.components)
	e.components = append(e.components, component)

	return component, true
}

func (e *BaseEntity) Component(tag string) Component {
	if idx, ok := e.tagToComponent[tag]; ok {
		return e.components[idx]
	}

	return nil
}
