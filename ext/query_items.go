package ext

import (
	"game/libs/bump"
	"game/vars"
)

type Recter interface {
	comparable
	Rect() (float64, float64, float64, float64)
}

func QueryItems[T comparable](item T, rect bump.Rect, tags ...bump.Tag) []T {
	itemsFilter := func(other bump.Item) bool {
		if e, ok := other.(T); ok {
			return e != item
		}

		return false
	}

	cols := vars.World.Space.Query(rect, itemsFilter, tags...)
	var items []T
	for _, c := range cols {
		if e, ok := c.Other.(T); ok {
			items = append(items, e)
		}
	}

	return items
}

func QueryFront[T Recter](recter T, dist, height float64, onRight bool) []T {
	ex, ey, ew, _ := recter.Rect()
	rect := bump.Rect{X: ex - dist, Y: ey - height, W: dist, H: height * 2}
	if onRight {
		rect.X += dist + ew
	}

	return QueryItems(recter, rect)
}
