package core

import (
	"game/libs/bump"
	"game/libs/camera"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

var IDCount uint64 = 100

type World struct {
	Space         *bump.Space
	Camera        *camera.Camera
	entities      []*Entity
	entitiesCache map[uint64]uint64
	Map           *Map
	freezeTimer   float64
}

func NewWorld(width, height float64) *World {
	return &World{bump.NewSpace(), camera.New(width, height), nil, map[uint64]uint64{}, nil, 0}
}

func (w *World) Add(entity *Entity) *Entity {
	entity.ID = IDCount
	w.entitiesCache[IDCount] = IDCount
	IDCount++
	entity.World = w
	w.entities = append(w.entities, entity)
	for _, comp := range entity.components {
		if initializer, ok := comp.(Initializer); ok {
			initializer.Init(entity)
		}
	}

	return entity
}

func (w *World) GetByID(id uint64) *Entity {
	if entID, ok := w.entitiesCache[id]; ok {
		return w.entities[entID]
	}

	return nil
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
		for _, comp := range e.components {
			if updater, ok := comp.(Updater); ok {
				updater.Update(dt)
			}
		}
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
		entityPos.Reset()
		entityPos.Translate(e.X, e.Y)
		entityPos.Translate(-cx, -cy)
		for _, comp := range e.components {
			if drawer, ok := comp.(Drawer); ok {
				drawer.Draw(screen, entityPos)
			}
		}
	}
	if w.Map != nil {
		foreground, _ := w.Map.foregroundImage.SubImage(w.Camera.Bounds()).(*ebiten.Image)
		screen.DrawImage(foreground, nil)
	}
}

func (w *World) RemoveEntity(id uint64) {
	for i, e := range w.entities {
		if e.ID == id {
			w.entities = append(w.entities[:i], w.entities[i+1:]...)

			break
		}
	}
}

func (w *World) Freeze(time float64) {
	w.freezeTimer = time
}
