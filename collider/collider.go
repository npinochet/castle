package collider

import (
	"game/core"
	"game/libs/bump"
)

var Space *bump.Space

func Query(rect bump.Rect, filter func(item bump.Item) bool) []*bump.Collision {
	c.debugQueryRect = rect

	return c.space.Query(rect, filter)
}

func QueryEntites(rect bump.Rect) []*core.Entity {
	entityFilter := func(item bump.Item) bool {
		if comp, ok := item.(*Comp); ok {
			return comp != c
		}

		return false
	}

	cols := c.Query(rect, entityFilter)
	var ents []*core.Entity
	for _, c := range cols {
		if comp, ok := c.Other.(*Comp); ok {
			ents = append(ents, comp.entity)
		}
	}

	return ents
}

func QueryFront(dist, height float64, lookingRight bool) []*core.Entity {
	rect := bump.Rect{X: -dist, Y: -height / 2, W: dist, H: height}
	if lookingRight {
		rect.X += dist
	}
	rect.X += c.entity.X
	rect.Y += c.entity.Y

	return c.QueryEntites(rect)
}
