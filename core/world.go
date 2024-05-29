package core

import (
	"game/libs/bump"
	"game/libs/camera"
	"log"
	"reflect"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

type Entity interface {
	Init()
	Update(dt float64)
	Components() []Component
	Component(name string) Component
	Position() (float64, float64)
	SetPosition(x, y float64)
	Rect() (float64, float64, float64, float64)
	SetSize(w, h float64)
}

type Component interface {
	Init(entity Entity)
	Update(dt float64)
	Draw(screen *ebiten.Image, entityPos ebiten.GeoM)
	Remove()
}

type Sortable interface{ Priority() int }

type World struct {
	Space      *bump.Space
	Camera     *camera.Camera
	Speed      float64
	Map        *Map
	entities   []Entity
	idToEntity map[uint]Entity
	entityToID map[Entity]uint
	toInit     []Entity

	freezeTimer float64
}

func NewWorld(width, height float64) *World {
	return &World{
		Space:      bump.NewSpace(),
		Camera:     camera.New(width, height),
		Speed:      1,
		idToEntity: map[uint]Entity{},
		entityToID: map[Entity]uint{},
	}
}

func (w *World) Add(entity Entity) Entity {
	w.toInit = append(w.toInit, entity)

	return entity
}

func (w *World) AddWithID(entity Entity, id uint) Entity {
	w.idToEntity[id] = entity
	w.entityToID[entity] = id

	return w.Add(entity)
}

func (w *World) Update(dt float64) {
	for _, entity := range w.toInit {
		w.insertEntity(entity)
	}
	w.toInit = nil
	dt *= w.Speed
	w.Camera.Update(dt)
	if w.freezeTimer -= dt; w.freezeTimer >= 0 {
		return
	}
	w.Map.Update(dt)
	// Copy entities to prevent inconsistencies when adding/removing entities during update.
	entities := make([]Entity, len(w.entities))
	copy(entities, w.entities)
	for _, e := range entities {
		for _, c := range e.Components() {
			c.Update(dt)
		}
		e.Update(dt)
	}
}

func (w *World) Draw(screen *ebiten.Image) {
	w.Map.Draw(screen, w.Camera, func() {
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
	})
}

func (w *World) SetMap(tiledMap *Map, roomsLayer string) {
	w.Map = tiledMap

	rooms, ok := tiledMap.GetObjectsRects(roomsLayer)
	if !ok {
		log.Println("world: room layer not found")
	}
	w.Camera.SetRooms(rooms)
}

func (w *World) Get(id uint) Entity       { return w.idToEntity[id] }
func (w *World) GetID(entity Entity) uint { return w.entityToID[entity] }
func (w *World) GetAll() []Entity         { return w.entities }

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
			for _, c := range e.Components() {
				c.Remove()
			}
			w.entities = append(w.entities[:i], w.entities[i+1:]...)
			if id, ok := w.entityToID[e]; ok {
				delete(w.idToEntity, id)
			}
			delete(w.entityToID, e)

			break
		}
	}
}

func (w *World) RemoveAll() {
	for _, e := range w.entities {
		for _, c := range e.Components() {
			c.Remove()
		}
	}
	w.entities = nil
	w.idToEntity = map[uint]Entity{}
	w.entityToID = map[Entity]uint{}
}

func (w *World) Freeze(time float64) { w.freezeTimer = time }

func (w *World) insertEntity(entity Entity) {
	for _, c := range entity.Components() {
		c.Init(entity)
	}
	entity.Init()
	i, _ := slices.BinarySearchFunc(w.entities, entity, func(o, e Entity) int {
		oz, ez := 0, 0
		if s, ok := o.(Sortable); ok {
			oz = s.Priority()
		}
		if s, ok := e.(Sortable); ok {
			ez = s.Priority()
		}

		return oz - ez
	})
	w.entities = slices.Insert(w.entities, i, entity)
}
