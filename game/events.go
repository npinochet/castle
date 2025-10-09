package game

import (
	"game/comps/hitbox"
	"game/comps/stats"
	"game/core"
	"game/ext"
	"game/libs/bump"
	"game/vars"
	"log"
	"strconv"

	"github.com/lafriks/go-tiled"
)

const eventsLayerName = "events"

type Event func(object *tiled.Object) func() (finish bool)

type emptyEntity struct {
	core.BaseEntity
	update func()
}

func (e *emptyEntity) Init() {}
func (e *emptyEntity) Update(dt float64) {
	if e.update != nil {
		e.update()
	}
}

var (
	hitboxEntity = &emptyEntity{}
	events       = map[string]Event{
		"ChestSpawn": func(object *tiled.Object) func() bool {
			id1, _ := strconv.Atoi(object.Properties.GetString("enemy1"))
			id2, _ := strconv.Atoi(object.Properties.GetString("enemy2"))
			id3, _ := strconv.Atoi(object.Properties.GetString("enemy3"))
			enemy1 := vars.World.RemoveID(uint(id1))
			enemy2 := vars.World.RemoveID(uint(id2))
			enemy3 := vars.World.RemoveID(uint(id3))

			return func() bool {
				vars.World.AddWithID(enemy1, uint(id1))
				vars.World.AddWithID(enemy2, uint(id2))
				vars.World.AddWithID(enemy3, uint(id3))

				return true
			}
		},
		"Kill": func(object *tiled.Object) func() bool {
			id, _ := strconv.Atoi(object.Properties.GetString("target"))
			target := vars.World.Get(uint(id))
			if target == nil {
				return nil
			}
			targetStats := core.Get[*stats.Comp](target)
			if targetStats == nil {
				return nil
			}

			return func() bool {
				targetStats.Health = 0

				return true
			}
		},
	}
)

func LoadMapEvents(tileMap *core.Map) {
	for _, object := range tileMap.GetObjects(eventsLayerName) {
		event := events[object.Name]
		if event == nil {
			continue
		}
		updateFunc := event(object)
		if updateFunc == nil {
			continue
		}
		rect := bump.Rect{X: object.X, Y: object.Y, W: object.Width, H: object.Height}
		trigger := object.Type
		if trigger == "" {
			trigger = object.Class
		}
		if trigger == "" {
			trigger = object.Properties.GetString("trigger")
		}
		switch trigger {
		case "Hit":
			addHitbox(rect, updateFunc)
		case "Enter":
			addEnterbox(rect, updateFunc)
		default:
			log.Printf("Warning: unknown event trigger '%s' for event '%s'\n", trigger, object.Name)
		}
	}
}

func addEnterbox(rect bump.Rect, enterFunc func() bool) {
	entity := &emptyEntity{BaseEntity: core.BaseEntity{X: rect.X, Y: rect.Y, W: rect.W, H: rect.H}}
	entity.update = func() {
		if items := ext.QueryItems[core.Entity](nil, rect, "body"); len(items) > 0 {
			if finish := enterFunc(); finish {
				vars.World.Remove(entity)
			}
		}
	}
	vars.World.Add(entity)
}

func addHitbox(rect bump.Rect, hitFunc func() bool) {
	comp := &hitbox.Comp{}
	comp.HitFunc = func(core.Entity, *bump.Collision, float64, hitbox.ContactType) {
		if finish := hitFunc(); finish {
			comp.Remove()
		}
	}
	comp.Init(hitboxEntity)
	comp.PushHitbox(rect, hitbox.Hit, nil)
}
