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
	entitiesCache map[uint64]*Entity
	Map           *Map
	Debug         bool
}

func NewWorld(width, height float64) *World {
	return &World{bump.NewSpace(), camera.New(width, height), nil, map[uint64]*Entity{}, nil, true}
}

func (w *World) AddEntity(entity *Entity) *Entity {
	entity.Active = true
	entity.ID = IDCount
	IDCount++
	entity.World = w
	w.entities = append(w.entities, entity)
	entity.InitComponents()

	return entity
}

func (w *World) GetEntityByID(id uint64) *Entity {
	if ent, ok := w.entitiesCache[id]; ok {
		return ent
	}

	for _, ent := range w.entities {
		if ent.ID == id {
			w.entitiesCache[id] = ent

			return ent
		}
	}

	return nil
}

func (w *World) SetMap(tiledMap *Map, roomsLayer string) {
	w.Map = tiledMap

	rooms, ok := tiledMap.GetObjectsRects(roomsLayer)
	if !ok {
		log.Println("Room layer not found")
	}
	w.Camera.SetRooms(rooms)
}

func (w *World) Update(dt float64) {
	w.Camera.Update(dt)
	if w.Map != nil {
		w.Map.Update(dt)
	}
	for _, e := range w.entities {
		if e.Active {
			e.Update(dt)
		}
	}
}

func (w *World) Draw(screen *ebiten.Image) {
	// TODO: Hide BackgroundImage and ForegroundImage on Tiled package.
	if w.Map != nil {
		background, _ := w.Map.backgroundImage.SubImage(w.Camera.Bounds()).(*ebiten.Image)
		screen.DrawImage(background, nil)
	}
	for _, e := range w.entities {
		if e.Active {
			e.Draw(screen)
		}
	}
	if w.Map != nil {
		foreground, _ := w.Map.foregroundImage.SubImage(w.Camera.Bounds()).(*ebiten.Image)
		screen.DrawImage(foreground, nil)
	}
}

func (w *World) RemoveEntity(id uint64) {
	deleteCount := -1
	for i, e := range w.entities {
		if e.ID == id {
			deleteCount = i

			break
		}
	}
	if deleteCount >= 0 {
		w.entities[deleteCount].Destroy()
		w.entities = append(w.entities[:deleteCount], w.entities[deleteCount+1:]...)
	}
}
