package core

import "github.com/hajimehoshi/ebiten/v2"

type Component interface {
	SetActive(active bool)
	IsActive() bool
}

type Initializer interface{ Init(*Entity) }
type Updater interface{ Update(dt float64) }
type Destroyer interface{ Destroy() }
type Drawer interface {
	Draw(screen *ebiten.Image, enitiyPos ebiten.GeoM)
}
type DebugDrawer interface {
	DebugDraw(screen *ebiten.Image, enitiyPos ebiten.GeoM)
}

type Entity struct {
	ID         uint64
	World      *World
	Active     bool
	X, Y       float64
	Components []Component
}

func (e *Entity) Position() (float64, float64) { return e.X, e.Y }

func (e *Entity) AddComponent(components ...Component) Component {
	for _, component := range components {
		component.SetActive(true)
		e.Components = append(e.Components, component)
	}

	return components[0]
}

func (e *Entity) InitComponents() {
	for _, c := range e.Components {
		if initializer, ok := c.(Initializer); ok {
			initializer.Init(e)
		}
	}
}

func (e *Entity) Update(dt float64) {
	for _, c := range e.Components {
		if updater, ok := c.(Updater); c.IsActive() && ok {
			updater.Update(dt)
		}
	}
}

func (e *Entity) Draw(screen *ebiten.Image) {
	enitiyPos := ebiten.GeoM{}
	enitiyPos.Translate(e.X, e.Y)
	x, y := e.World.Camera.Position()
	enitiyPos.Translate(-x, -y)

	for _, c := range e.Components {
		if c.IsActive() {
			if drawer, ok := c.(Drawer); ok {
				drawer.Draw(screen, enitiyPos)
			}
			if drawer, ok := c.(DebugDrawer); e.World.Debug && ok {
				drawer.DebugDraw(screen, enitiyPos)
			}
		}
	}
}

// TODO: Is it necessary? Isn't enough with the GC?
func (e *Entity) Destroy() {
	e.Active = false
	for _, c := range e.Components {
		c.SetActive(false)
		// if destroyer, ok := c.(Destroyer); ok {
		//   destroyer.Destroy()
		// }
	}
}
