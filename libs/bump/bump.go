package bump

import (
	"math"
	"sort"
)

// Collision detection and resolution library based on bump.lua by kikito.

const DELTA = 1e-10 // floating-point margin of error.

type Item interface{}
type Slope struct{ L, R float64 } // Rect left and right heights, ([0, 1]) 0 = full height, 1 = zero height.
type Rect struct {
	X, Y, W, H float64
	Priority   int   // Which rects should be evaluated for collision first (Slopes should have higher priority than solid blocks).
	Slope      Slope // Slope adjusted H accordingly.
}
type Vec2 struct{ X, Y float64 }

type Filter func(item, other Item) (response ColType, ok bool)
type SimpleFilter func(item Item) bool
type Response func(goal Vec2, col *Collision, filter Filter) (newGoal Vec2, newCols []*Collision)
type Collision struct {
	Overlaps            bool
	Intersection        float64
	Move, Touch, Normal Vec2
	Item, Other         Item
	ItemRect, OtherRect Rect
	Type                ColType
	TypeData            Vec2
}

type ColType int

const (
	Touch ColType = iota
	Cross
	PureSlide
	Slide
)

func NewRect(x, y, w, h float64) Rect                { return Rect{X: x, Y: y, W: w, H: h} }
func DefaultFilter(item, other Item) (ColType, bool) { return Slide, true }
func NilFilter(item, other Item) (ColType, bool)     { return 0, false }

func DefaultSimpleFilter(item Item) bool {
	return true
}

type Space struct {
	Rects     map[Item]Rect
	responses map[ColType]Response
}

func NewSpace() *Space {
	space := &Space{}
	space.Rects = map[Item]Rect{}
	space.responses = map[ColType]Response{
		Touch: func(goal Vec2, col *Collision, filter Filter) (Vec2, []*Collision) {
			return col.Touch, nil
		},
		Cross: func(goal Vec2, col *Collision, filter Filter) (Vec2, []*Collision) {
			return goal, space.Project(col.Item, col.ItemRect, goal, filter)
		},
		PureSlide: func(goal Vec2, col *Collision, filter Filter) (Vec2, []*Collision) {
			if col.Move.X != 0 || col.Move.Y != 0 {
				if col.Normal.X != 0 {
					goal.X = col.Touch.X
				} else {
					goal.Y = col.Touch.Y
				}
			}
			col.TypeData = goal
			rect := Rect{col.Touch.X, col.Touch.Y, col.ItemRect.W, col.ItemRect.H, col.ItemRect.Priority, col.ItemRect.Slope}

			return goal, space.Project(col.Item, rect, goal, filter)
		},
		Slide: func(goal Vec2, col *Collision, filter Filter) (Vec2, []*Collision) {
			if !col.OtherRect.IsSlope() {
				return space.responses[PureSlide](goal, col, filter)
			}
			col.Normal = Vec2{0, 0}
			if height := col.OtherRect.slopeY(goal.X + col.ItemRect.W/2); goal.Y > height-col.ItemRect.H {
				goal.Y = height - col.ItemRect.H
				col.Touch.Y = height - col.ItemRect.H
				col.Normal = Vec2{0, -1}
			}
			col.TypeData = goal

			return goal, nil
		},
	}

	return space
}

func (s *Space) Set(item Item, rect Rect) { s.Rects[item] = rect }
func (s *Space) Remove(item Item)         { delete(s.Rects, item) }
func (s *Space) Move(item Item, targetGoal Vec2, filter Filter) (goal Vec2, cols []*Collision) {
	goal, cols = s.Check(item, targetGoal, filter)
	rect := s.Rects[item]
	s.Set(item, Rect{goal.X, goal.Y, rect.W, rect.H, rect.Priority, rect.Slope})

	return
}

func (s *Space) Check(item Item, goal Vec2, filter Filter) (Vec2, []*Collision) {
	if filter == nil {
		filter = DefaultFilter
	}

	visited := map[Item]bool{item: true}
	visitedFilter := func(item, other Item) (ColType, bool) {
		if visited[other] {
			return 0, false
		}

		return filter(item, other)
	}

	rect := s.Rects[item]
	projectedCols := s.Project(item, rect, goal, visitedFilter)

	var cols []*Collision
	for len(projectedCols) > 0 {
		col := projectedCols[0]
		visited[col.Other] = true
		goal, projectedCols = s.responses[col.Type](goal, col, visitedFilter)
		cols = append(cols, col)
	}

	return goal, cols
}

func (s *Space) Project(item Item, rect Rect, goal Vec2, filter Filter) []*Collision {
	if filter == nil {
		filter = DefaultFilter
	}

	var cols []*Collision
	visited := map[Item]bool{item: true}
	// TODO: Optimize using cells.
	for _, other := range sortPriority(s.Rects) {
		if visited[other] {
			continue
		}
		visited[other] = true
		if responseName, ok := filter(item, other); ok {
			otherRect := s.Rects[other]
			if col, ok := detectCollision(rect, otherRect, goal); ok {
				col.Item, col.Other = item, other
				col.Type = responseName
				cols = append(cols, col)
			}
		}
	}

	return cols
}

func (s *Space) Query(rect Rect, filter SimpleFilter) []*Collision {
	if filter == nil {
		filter = DefaultSimpleFilter
	}
	projectFilter := func(item, other Item) (ColType, bool) { return 0, filter(other) }

	return s.Project(nil, rect, Vec2{rect.X, rect.Y}, projectFilter)
}

func Overlaps(r1, r2 Rect) bool {
	return rectContainsPoint(rectDiff(r1, r2), Vec2{})
}

func sortPriority(rects map[Item]Rect) []Item {
	i, keys := 0, make([]Item, len(rects))
	for k := range rects {
		keys[i] = k
		i++
	}
	sort.Slice(keys, func(i, j int) bool { return rects[keys[i]].Priority > rects[keys[j]].Priority })

	return keys
}

// Liang-Barsky algorithm.
func lineSegmentIntersection(rect Rect, p1, p2 Vec2) (i1, i2 float64, normal Vec2, ok bool) {
	dx, dy := p2.X-p1.X, p2.Y-p1.Y
	p := [4]float64{-dx, dx, -dy, dy} // left, right, top, bottom.
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

			continue
		}
		r := q[i] / p[i]
		if p[i] < 0 {
			if r > i2 {
				return
			} else if r > i1 {
				i1 = r
				normal = Vec2{nx[i], ny[i]}
			}

			continue
		}
		if r < i1 {
			return
		} else if r < i2 {
			i2 = r
		}
	}

	ok = true

	return i1, i2, normal, ok
}

func detectCollision(rect1, rect2 Rect, goal Vec2) (*Collision, bool) {
	col := &Collision{}
	col.Move = Vec2{goal.X - rect1.X, goal.Y - rect1.Y}
	col.ItemRect, col.OtherRect = rect1, rect2
	interRect := rectDiff(rect1, rect2)

	if !detectCollisionPhase1(interRect, rect1, col) {
		return col, false
	}

	if !col.Overlaps {
		return col, true
	}

	if (col.Move == Vec2{}) {
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
			return col, false
		}
		col.Normal = normal
		col.Touch = Vec2{rect1.X + col.Move.X*i1, rect1.Y + col.Move.Y*i1}
	}

	return col, true
}

func detectCollisionPhase1(interRect, rect1 Rect, col *Collision) bool {
	collisioned := false
	if rectContainsPoint(interRect, Vec2{}) {
		p := rectNearestCorner(interRect, Vec2{})
		wi, hi := math.Min(rect1.W, math.Abs(p.X)), math.Min(rect1.H, math.Abs(p.Y))
		col.Intersection = -wi * hi
		col.Overlaps = true
		collisioned = true
	} else {
		i1, i2, normal, found := lineSegmentIntersection(interRect, Vec2{}, col.Move)
		if found && i1 < 1 && math.Abs(i1-i2) >= DELTA && (i1 > -DELTA || i1 == 0 && i2 > 0) {
			col.Normal = normal
			col.Intersection = i1
			col.Overlaps = false
			collisioned = true
			col.Touch = Vec2{rect1.X + col.Move.X*col.Intersection, rect1.Y + col.Move.Y*col.Intersection}
		}
	}

	return collisioned
}

// Minkowsky Difference between 2 Rects.
func rectDiff(r1, r2 Rect) Rect {
	return Rect{r2.X - r1.X - r1.W, r2.Y - r1.Y - r1.H, r1.W + r2.W, r1.H + r2.H, 0, Slope{}}
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

func (r Rect) IsSlope() bool {
	return r.Slope != Slope{}
}

func (r Rect) slopeY(x float64) float64 {
	prog := (x - r.X) / r.W
	clamp := math.Min(math.Max(prog, 0), 1)
	lerp := r.Slope.L + clamp*(r.Slope.R-r.Slope.L)

	return r.Y + lerp*r.H
}
