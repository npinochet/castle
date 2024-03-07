package ext

import (
	"game/core"
	"game/libs/bump"
	"game/vars"
)

func QueryEntites(entity core.Entity, rect bump.Rect) []core.Entity {
	entityFilter := func(item bump.Item) bool {
		if e, ok := item.(core.Entity); ok {
			return e != entity
		}

		return false
	}

	cols := vars.World.Space.Query(rect, entityFilter)
	var ents []core.Entity
	for _, c := range cols {
		if e, ok := c.Other.(core.Entity); ok {
			ents = append(ents, e)
		}
	}

	return ents
}

func QueryFront(entity core.Entity, dist, height float64, onRight bool) []core.Entity {
	rect := bump.Rect{X: -dist, Y: -height / 2, W: dist, H: height}
	if onRight {
		rect.X += dist
	}
	ex, ey := entity.Position()
	rect.X += ex
	rect.Y += ey

	return QueryEntites(entity, rect)
}
