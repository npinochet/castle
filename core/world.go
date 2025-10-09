package core

import (
	"game/libs/bump"
	"game/libs/camera"
	"log"
	"math"
	"reflect"
	"sync"

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
	Draw(pipeline *Pipeline, entityPos ebiten.GeoM)
	Remove()
}

type World struct {
	Space      *bump.Space
	Camera     *camera.Camera
	Speed      float64
	Map        *Map
	entities   []Entity
	idToEntity map[uint]Entity
	entityToID map[Entity]uint
	toInit     []Entity
	toRemove   []Entity
	removed    []Entity
	mutex      sync.Mutex

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
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.toInit = append(w.toInit, entity)

	return entity
}

func (w *World) AddWithID(entity Entity, id uint) Entity {
	w.idToEntity[id] = entity
	w.entityToID[entity] = id

	return w.Add(entity)
}

func (w *World) Update(dt float64) {
	w.mutex.Lock()
	for _, entity := range w.toInit {
		for _, c := range entity.Components() {
			c.Init(entity)
		}
		entity.Init()
		w.entities = append(w.entities, entity)
	}
	w.toInit = nil
	w.mutex.Unlock()

	dt *= w.Speed
	w.Camera.Update(dt)
	if w.freezeTimer -= dt; w.freezeTimer >= 0 {
		return
	}
	w.Map.Update(dt)
	for _, e := range w.entities {
		if !w.Camera.InFrame(e, 1, 1) {
			continue
		}
		for _, c := range e.Components() {
			c.Update(dt)
		}
		e.Update(dt)
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()
	for _, entity := range w.toRemove {
		for i, e := range w.entities {
			if entity != e {
				continue
			}
			for _, c := range e.Components() {
				c.Remove()
			}
			w.entities[i] = w.entities[len(w.entities)-1]
			w.entities = w.entities[:len(w.entities)-1]
			w.removed = append(w.removed, e)

			break
		}
	}
	w.toRemove = nil
}

func (w *World) Draw(pipeline *Pipeline) {
	w.Map.Draw(pipeline, w.Camera)
	cx, cy := w.Camera.Position()
	entityPos := ebiten.GeoM{}
	for _, e := range w.entities {
		if !w.Camera.InFrame(e, 0.1, 0.1) {
			continue
		}
		x, y := e.Position()
		entityPos.Reset()
		entityPos.Translate(math.Ceil(x-cx), math.Ceil(y-cy))
		for _, c := range e.Components() {
			c.Draw(pipeline, entityPos)
		}
	}
}

func (w *World) SetMap(tiledMap *Map, roomsLayer string) {
	w.Map = tiledMap
	rooms := tiledMap.GetObjectsRects(roomsLayer)
	if rooms == nil {
		log.Println("world: room layer not found")
	} else {
		w.Camera.SetRooms(rooms)
	}
	tiledMap.LoadTilesetCollisionObjects(w.Space)
}

func (w *World) Get(id uint) Entity       { return w.idToEntity[id] }
func (w *World) GetID(entity Entity) uint { return w.entityToID[entity] }
func (w *World) GetAll() []Entity         { return w.entities }
func (w *World) GetRemoved() []Entity     { return w.removed }

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
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.toRemove = append(w.toRemove, entity)
}

func (w *World) RemoveID(id uint) Entity {
	entity := w.idToEntity[id]
	if entity == nil {
		return nil
	}
	delete(w.idToEntity, id)
	delete(w.entityToID, entity)
	w.Remove(entity)

	return entity
}

func (w *World) RemoveAll() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

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
