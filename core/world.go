package core

import (
	"game/libs/bump"
	"game/libs/camera"
	"log"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
)

type Entity interface {
	Components() []Component
	Component(name string) Component
	Init()
	Update(dt float64)
	Position() (float64, float64)
	SetPosition(x, y float64)
	Rect() (float64, float64, float64, float64)
}

type Component interface {
	Init(entity Entity)
	Update(dt float64)
	Draw(screen *ebiten.Image, entityPos ebiten.GeoM)
}

type World struct {
	Space       *bump.Space
	Camera      *camera.Camera
	entities    []Entity
	Map         *Map
	freezeTimer float64
}

func NewWorld(width, height float64) *World {
	return &World{bump.NewSpace(), camera.New(width, height), nil, nil, 0}
}

func (w *World) Add(entity Entity) Entity {
	w.entities = append(w.entities, entity)
	for _, c := range entity.Components() {
		c.Init(entity)
	}
	entity.Init()

	return entity
}

func (w *World) SetMap(tiledMap *Map, roomsLayer string) {
	w.Map = tiledMap

	rooms, ok := tiledMap.GetObjectsRects(roomsLayer)
	if !ok {
		log.Println("world: room layer not found")
	}
	w.Camera.SetRooms(rooms)
}

func (w *World) Update(dt float64) {
	w.Camera.Update(dt)
	if w.freezeTimer -= dt; w.freezeTimer >= 0 {
		return
	}
	w.Map.Update(dt)
	for _, e := range w.entities {
		for _, c := range e.Components() {
			c.Update(dt)
		}
		e.Update(dt)
	}
}

func (w *World) Draw(screen *ebiten.Image) {
	// TODO: Hide BackgroundImage and ForegroundImage draw code on Map package.
	if w.Map != nil {
		background, _ := w.Map.backgroundImage.SubImage(w.Camera.Bounds()).(*ebiten.Image)
		screen.DrawImage(background, nil)
	}
	cx, cy := w.Camera.Position()
	entityPos := ebiten.GeoM{}
	for _, e := range w.entities {
		x, y := e.Position()
		entityPos.Reset()
		entityPos.Translate(x, y)
		entityPos.Translate(-cx, -cy)
		for _, c := range e.Components() {
			c.Draw(screen, entityPos)
		}
	}
	if w.Map != nil {
		foreground, _ := w.Map.foregroundImage.SubImage(w.Camera.Bounds()).(*ebiten.Image)
		screen.DrawImage(foreground, nil)
	}
}

func Get[T Component](entity Entity) T {
	var t T
	tag := reflect.TypeOf(t).String()
	t, _ = entity.Component(tag).(T)

	return t
}

func GetWithTag[T Component](entity Entity, tag string) T {
	t, _ := entity.Component(tag).(T)

	return t
}

func (w *World) Remove(entity Entity) {
	for i, e := range w.entities {
		if entity == e {
			w.entities = append(w.entities[:i], w.entities[i+1:]...)

			break
		}
	}
}

func (w *World) Freeze(time float64) {
	w.freezeTimer = time
}
