package bump

import (
	"math"
)

// Collision detection and resolution lib based on bump.lua by kikito

const DELTA float64 = 1e-10 // floating-point margin of error

type Item interface{}
type Rect struct{ X, Y, W, H float64 }
type Vec2 struct{ X, Y float64 }

type Filter func(item, other Item) (response ColType, ok bool)
type SimpleFilter func(item Item) bool
type Response func(goal Vec2, col Colision, filter Filter) (newGoal Vec2, newCols []Colision)
type Colision struct {
	Overlaps            bool
	Intersection        float64
	Move                Vec2
	Touch               Vec2
	Normal              Vec2
	Item, Other         Item
	ItemRect, OtherRect Rect
	Type                ColType
	TypeData            Vec2
}

type ColType int

const (
	Touch ColType = iota
	Cross
	Slide
)

func DefaultFilter(item, other Item) (response ColType, collide bool) {
	return Slide, true
}

func DefaultSimpleFilter(item Item) bool {
	return true
}

type Space struct {
	rects     map[Item]Rect
	responses map[ColType]Response
}

func NewSpace() *Space {
	space := &Space{}
	space.rects = make(map[Item]Rect)
	space.responses = map[ColType]Response{
		Touch: func(goal Vec2, col Colision, filter Filter) (newGoal Vec2, cols []Colision) {
			return col.Touch, []Colision{}
		},
		Cross: func(goal Vec2, col Colision, filter Filter) (newGoal Vec2, cols []Colision) {
			return goal, space.Project(col.Item, col.ItemRect, goal, filter)
		},
		Slide: func(goal Vec2, col Colision, filter Filter) (newGoal Vec2, cols []Colision) {
			if col.Move.X != 0 || col.Move.Y != 0 {
				if col.Normal.X != 0 {
					goal.X = col.Touch.X
				} else {
					goal.Y = col.Touch.Y
				}
			}
			col.TypeData = goal
			return goal, space.Project(col.Item, Rect{col.Touch.X, col.Touch.Y, col.ItemRect.W, col.ItemRect.H}, goal, filter)
		},
	}
	return space
}

func (s *Space) Set(item Item, rect Rect) { s.rects[item] = rect }
func (s *Space) Remove(item Item)         { delete(s.rects, item) }
func (s *Space) Move(item Item, targetGoal Vec2, filter Filter) (goal Vec2, cols []Colision) {
	goal, cols = s.Check(item, targetGoal, filter)
	rect := s.rects[item]
	s.Set(item, Rect{goal.X, goal.Y, rect.W, rect.H})
	return
}

func (s *Space) Check(item Item, targetGoal Vec2, filter Filter) (goal Vec2, cols []Colision) {
	goal = targetGoal
	if filter == nil {
		filter = DefaultFilter
	}

	visited := map[Item]bool{}
	visited[item] = true

	visitedFilter := func(item, other Item) (response ColType, ok bool) {
		if visited[other] {
			return
		}
		return filter(item, other)
	}

	rect := s.rects[item]
	projectedCols := s.Project(item, rect, goal, visitedFilter)

	for len(projectedCols) > 0 {
		col := projectedCols[0]
		cols = append(cols, col)
		visited[col.Other] = true
		goal, projectedCols = s.responses[col.Type](goal, col, visitedFilter)
	}

	return
}

func (s *Space) Project(item Item, rect Rect, goal Vec2, filter Filter) (cols []Colision) {
	if filter == nil {
		filter = DefaultFilter
	}

	visited := map[Item]bool{}
	visited[item] = true

	for other, otherRect := range s.rects {
		if !visited[other] {
			visited[other] = true
			if responseName, ok := filter(item, other); ok {
				if col, ok := detectCollision(rect, otherRect, goal); ok {
					col.Item = item
					col.Other = other
					col.Type = responseName
					cols = append(cols, col)
				}
			}
		}
	}

	return
}

func (s *Space) Query(rect Rect, filter SimpleFilter) (cols []Colision) {
	if filter == nil {
		filter = DefaultSimpleFilter
	}

	for other, otherRect := range s.rects {
		if filter(other) {
			if col, ok := detectCollision(rect, otherRect, Vec2{rect.X, rect.Y}); ok {
				col.Item = other
				col.Other = other
				cols = append(cols, col)
			}
		}
	}

	return
}

func Overlaps(r1, r2 Rect) bool {
	return rectContainsPoint(rectDiff(r1, r2), Vec2{})
}

// Liang-Barsky algorithm
func lineSegmentIntersection(rect Rect, p1, p2 Vec2) (i1, i2 float64, normal Vec2, ok bool) {
	dx, dy := p2.X-p1.X, p2.Y-p1.Y
	p := [4]float64{-dx, dx, -dy, dy} // left, right, top, bottom
	q := [4]float64{p1.X - rect.X, rect.X + rect.W - p1.X, p1.Y - rect.Y, rect.Y + rect.H - p1.Y}
	nx := [4]float64{-1, 1, 0, 0}
	ny := [4]float64{0, 0, -1, 1}

	i1, i2 = math.Inf(-1), math.Inf(1)
	normal = Vec2{}

	for i := 0; i < 4; i++ {
		if p[i] == 0 {
			if q[i] <= 0 {
				return
			}
		} else {
			r := q[i] / p[i]
			if p[i] < 0 {
				if r > i2 {
					return
				} else if r > i1 {
					i1 = r
					normal = Vec2{nx[i], ny[i]}
				}
			} else {
				if r < i1 {
					return
				} else if r < i2 {
					i2 = r
				}
			}
		}
	}

	ok = true
	return
}

func detectCollision(rect1, rect2 Rect, goal Vec2) (col Colision, ok bool) {
	col.Move = Vec2{goal.X - rect1.X, goal.Y - rect1.Y}
	col.ItemRect = rect1
	col.OtherRect = rect2
	interRect := rectDiff(rect1, rect2)

	var colisioned bool
	if rectContainsPoint(interRect, Vec2{}) {
		p := rectNearestCorner(interRect, Vec2{})
		wi, hi := math.Min(rect1.W, math.Abs(p.X)), math.Min(rect1.H, math.Abs(p.Y))
		col.Intersection = -wi * hi
		col.Overlaps = true
		colisioned = true
	} else {
		i1, i2, normal, found := lineSegmentIntersection(interRect, Vec2{}, col.Move)
		if found && i1 < 1 && math.Abs(i1-i2) >= DELTA && (i1 > -DELTA || i1 == 0 && i2 > 0) {
			col.Normal = normal
			col.Intersection = i1
			col.Overlaps = false
			colisioned = true
			col.Touch = Vec2{rect1.X + col.Move.X*col.Intersection, rect1.Y + col.Move.Y*col.Intersection}
		}
	}

	if !colisioned {
		return
	}

	if col.Overlaps {
		if col.Move.X == 0 && col.Move.Y == 0 {
			p := rectNearestCorner(interRect, Vec2{})
			if math.Abs(p.X) < math.Abs(p.Y) {
				p.Y = 0
			} else {
				p.X = 0
			}
			col.Normal = Vec2{math.Copysign(1, p.X), math.Copysign(1, p.Y)}
			col.Touch = Vec2{rect1.X + p.X, rect1.Y + p.Y}
		} else {
			i1, _, normal, found := lineSegmentIntersection(interRect, Vec2{}, col.Move)
			if !found {
				return
			}
			col.Normal = normal
			col.Touch = Vec2{rect1.X + col.Move.X*i1, rect1.Y + col.Move.Y*i1}
		}
	}

	ok = true
	return
}

// Minkowsky Difference between 2 Rects
func rectDiff(r1, r2 Rect) Rect {
	return Rect{r2.X - r1.X - r1.W, r2.Y - r1.Y - r1.H, r1.W + r2.W, r1.H + r2.H}
}

func rectContainsPoint(r Rect, p Vec2) bool {
	return p.X-r.X > DELTA && p.Y-r.Y > DELTA && r.X+r.W-p.X > DELTA && r.Y+r.H-p.Y > DELTA
}

func rectNearestCorner(rect Rect, p Vec2) Vec2 {
	nearest := func(x, a, b float64) float64 {
		ret := b
		if math.Abs(a-x) < math.Abs(b-x) {
			ret = a
		}
		return ret
	}
	return Vec2{nearest(p.X, rect.X, rect.X+rect.W), nearest(p.Y, rect.Y, rect.Y+rect.H)}
}
