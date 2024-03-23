package ext

import (
	"game/libs/bump"
	"game/vars"
)

type Recter interface {
	comparable
	Rect() (float64, float64, float64, float64)
}

func QueryItems[T comparable](item T, rect bump.Rect) []T {
	itemsFilter := func(item bump.Item) bool {
		if e, ok := item.(T); ok {
			return e != item
		}

		return false
	}

	cols := vars.World.Space.Query(rect, itemsFilter)
	var items []T
	for _, c := range cols {
		if e, ok := c.Other.(T); ok {
			items = append(items, e)
		}
	}

	return items
}

func QueryFront[T Recter](recter T, dist, height float64, onRight bool) []T {
	rect := bump.Rect{X: -dist, Y: -height / 2, W: dist, H: height}
	ex, ey, ew, _ := recter.Rect()
	if onRight {
		rect.X += dist + ew
	}
	rect.X += ex
	rect.Y += ey

	return QueryItems(recter, rect)
}
