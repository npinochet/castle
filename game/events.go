package game

import (
	"game/comps/anim"
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

type Event struct {
	Saveable      bool
	PrepareEffect func(object *tiled.Object) func() (finish bool)
	Instances     []*EventInstance
}

type EventInstance struct {
	ID        uint
	Triggered bool
	effect    func() (finish bool)
	removed   bool
}

func (ei *EventInstance) Trigger() {
	if ei.removed {
		return
	}
	ei.Triggered = true
	if finish := ei.effect(); finish {
		vars.World.RemoveID(ei.ID)
		ei.removed = true
	}
}

type emptyEntity struct {
	core.BaseEntity
	init   func()
	update func()
}

func (e *emptyEntity) Init() {
	if e.init != nil {
		e.init()
	}
}
func (e *emptyEntity) Update(dt float64) {
	if e.update != nil {
		e.update()
	}
}
func (e *emptyEntity) Destroy() {}

var (
	Events = map[string]*Event{
		"ChestSpawn": {
			Saveable: true,
			PrepareEffect: func(object *tiled.Object) func() bool {
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
			}},
		"Kill": {
			Saveable: true,
			PrepareEffect: func(object *tiled.Object) func() bool {
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
			}},
		"Remove": {
			Saveable: true,
			PrepareEffect: func(object *tiled.Object) func() bool {
				id, _ := strconv.Atoi(object.Properties.GetString("target"))
				target := vars.World.Get(uint(id))
				if target == nil {
					return nil
				}

				return func() bool {
					vars.World.RemoveID(uint(id))

					return true
				}
			}},
		"TurnAround": {
			Saveable: false,
			PrepareEffect: func(object *tiled.Object) func() bool {
				id, _ := strconv.Atoi(object.Properties.GetString("entity"))
				entity := vars.World.Get(uint(id))
				if entity == nil {
					return nil
				}
				entityAnim := core.Get[*anim.Comp](entity)
				if entityAnim == nil {
					return nil
				}

				return func() bool {
					entityAnim.FlipX = !entityAnim.FlipX

					return true
				}
			}},
	}
)

func ReloadMapEvents(tileMap *core.Map) {
	for _, event := range Events {
		for _, instance := range event.Instances {
			vars.World.RemoveID(instance.ID)
		}
		event.Instances = nil
	}

	for _, object := range tileMap.GetObjects(eventsLayerName) {
		event, ok := Events[object.Name]
		if !ok {
			continue
		}
		effectFunc := event.PrepareEffect(object)
		if effectFunc == nil {
			continue
		}
		instance := &EventInstance{ID: uint(object.ID), effect: effectFunc}
		var entity core.Entity
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
			entity = hitboxEntity(rect, instance)
		case "Enter":
			entity = enterboxEntity(rect, instance)
		default:
			log.Printf("Warning: unknown event trigger '%s' for event '%s'\n", trigger, object.Name)
		}
		if entity == nil {
			log.Printf("Warning: could not create entity for event '%s'\n", object.Name)
			continue
		}
		vars.World.AddWithID(entity, instance.ID)
		event.Instances = append(event.Instances, instance)
	}
}

func enterboxEntity(rect bump.Rect, instance *EventInstance) core.Entity {
	entity := &emptyEntity{BaseEntity: core.BaseEntity{X: rect.X, Y: rect.Y, W: rect.W, H: rect.H}}
	entity.update = func() {
		if items := ext.QueryItems[core.Entity](nil, rect, "body"); len(items) > 0 {
			instance.Trigger()
		}
	}

	return entity
}

func hitboxEntity(rect bump.Rect, instance *EventInstance) core.Entity {
	entity := &emptyEntity{BaseEntity: core.BaseEntity{X: rect.X, Y: rect.Y, W: rect.W, H: rect.H}}
	comp := &hitbox.Comp{}
	comp.HitFunc = func(core.Entity, *bump.Collision, float64, hitbox.ContactType) { instance.Trigger() }
	entity.Add(comp)
	entity.init = func() { comp.PushHitbox(rect, hitbox.Hit, nil) }

	return entity
}
