package core

import (
	"fmt"
	"game/libs/bump"

	"github.com/hajimehoshi/ebiten/v2"
)

type UID uint64

var idCount UID = 0

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
	Id         UID
	World      *World
	Active     bool
	X, Y       float64
	Components []Component
}

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
		if u, ok := c.(Updater); c.IsActive() && ok {
			u.Update(dt)
		}
	}
}

func (e *Entity) Draw(screen *ebiten.Image) {
	enitiyPos := ebiten.GeoM{}
	enitiyPos.Translate(e.X, e.Y)
	if e.World.Camera != nil {
		x, y := e.World.Camera.Position()
		enitiyPos.Translate(-x, -y)
	}

	for _, c := range e.Components {
		if c.IsActive() {
			if d, ok := c.(Drawer); ok {
				d.Draw(screen, enitiyPos)
			}
			if d, ok := c.(DebugDrawer); e.World.Debug && ok {
				d.DebugDraw(screen, enitiyPos)
			}
		}
	}
}

func (e *Entity) Destroy() {
	e.Active = false
	for _, c := range e.Components {
		if d, ok := c.(Destroyer); c.IsActive() && ok {
			c.SetActive(false)
			d.Destroy()
		}
	}
	e.World = nil
}

type World struct {
	Space    *bump.Space
	Camera   *Camera
	entities []*Entity
	TiledMap *TiledMap
	Debug    bool
}

func NewWorld() *World {
	world := &World{bump.NewSpace(), nil, nil, nil, true}
	return world
}

func (w *World) AddEntity(entity *Entity) *Entity {
	entity.Active = true
	entity.Id = idCount
	idCount += 1
	entity.World = w
	w.entities = append(w.entities, entity)
	entity.InitComponents()
	return entity
}

func (w *World) LoadTiledMap(tiledMap *TiledMap, camera *Camera, roomsLayer string) {
	w.Camera = camera
	w.TiledMap = tiledMap

	rooms, ok := tiledMap.GetObjectsRects(roomsLayer)
	if !ok {
		fmt.Println("Room layer not found")
	}
	camera.SetRooms(rooms)
}

func (w *World) Update(dt float64) {
	if w.TiledMap != nil {
		w.TiledMap.Update(dt)
	}
	for _, e := range w.entities {
		if e.Active {
			e.Update(dt)
		}
	}
}

func (w *World) Draw(screen *ebiten.Image) {
	if w.TiledMap != nil {
		background := w.TiledMap.backgroundImage.SubImage(w.Camera.Bounds()).(*ebiten.Image)
		screen.DrawImage(background, nil)
	}
	for _, e := range w.entities {
		if e.Active {
			e.Draw(screen)
		}
	}
	if w.TiledMap != nil {
		foreground := w.TiledMap.foregroundImage.SubImage(w.Camera.Bounds()).(*ebiten.Image)
		screen.DrawImage(foreground, nil)
	}
}

func (w *World) RemoveEntity(id UID) {
	var delete int = -1
	for i, e := range w.entities {
		if e.Id == id {
			delete = i
			break
		}
	}
	if delete >= 0 {
		w.entities[delete].Destroy()
		w.entities = append(w.entities[:delete], w.entities[delete+1:]...)
	}
}
