package bump

import (
	"math"
	"slices"
	"sync"
)

// Collision detection and resolution library based on bump.lua by kikito.

var (
	CellSize   = 32.0
	SlopePivot = 0.5 // goes from 0 to 0.5, 0.5 being the center of the rect.
)

const DELTA = 1e-10 // floating-point margin of error.

type Item any
type Tag string

type Rect struct {
	X, Y, W, H float64
	Type       RectType
}
type Vec2 struct{ X, Y float64 }

type Filter func(item, other Item) (response ColType, selected bool)
type SelectFilter func(item Item) bool
type Response func(goal Vec2, col *Collision, filter Filter, tags ...Tag) (newGoal Vec2, newCols []*Collision)
type Collision struct {
	Overlaps            bool
	Intersection        float64
	Move, Touch, Normal Vec2
	Item, Other         Item
	ItemRect, OtherRect Rect
	Type                ColType
	PreviousGoal        Vec2
}

type RectType uint

const (
	Full             RectType = iota
	TopRightSlope             // A triangle slope where right angle is at the top right.
	TopLeftSlope              // A triangle slope where right angle is at the top left.
	BottomRightSlope          // A triangle slope where right angle is at the bottom right.
	BottomLeftSlope           // A triangle slope where right angle is at the bottom left.
)

type ColType uint

const (
	Touch ColType = iota
	Cross
	RectSlide
	Slide
)

type cell [2]int
type location struct {
	tag  Tag
	cell cell
}

func NewRect(x, y, w, h float64) Rect                 { return Rect{X: x, Y: y, W: w, H: h} }
func DefaultResponseFilter(_, _ Item) (ColType, bool) { return Slide, true }
func NilFilter(_, _ Item) (ColType, bool)             { return 0, false }

type Space struct {
	Responses   map[ColType]Response
	rects       map[Item]Rect
	tags        map[Item][]Tag
	searchSpace map[location]map[Item]bool
	cellSize    float64
	mutex       sync.RWMutex
}

func NewSpace() *Space {
	space := &Space{
		rects:       map[Item]Rect{},
		tags:        map[Item][]Tag{},
		searchSpace: map[location]map[Item]bool{},
		cellSize:    CellSize,
	}
	space.Responses = map[ColType]Response{
		Touch: func(_ Vec2, col *Collision, _ Filter, _ ...Tag) (Vec2, []*Collision) { return col.Touch, nil },
		Cross: func(goal Vec2, col *Collision, filter Filter, tags ...Tag) (Vec2, []*Collision) {
			return goal, space.Project(col.Item, col.ItemRect, goal, filter, tags...)
		},
		RectSlide: func(goal Vec2, col *Collision, filter Filter, tags ...Tag) (Vec2, []*Collision) {
			col.PreviousGoal = goal
			if (col.Move != Vec2{}) {
				if col.Normal.X != 0 {
					goal.X = col.Touch.X
				} else {
					goal.Y = col.Touch.Y
				}
			}
			rect := Rect{col.Touch.X, col.Touch.Y, col.ItemRect.W, col.ItemRect.H, col.ItemRect.Type}

			return goal, space.Project(col.Item, rect, goal, filter, tags...)
		},
		Slide: func(goal Vec2, col *Collision, filter Filter, tags ...Tag) (Vec2, []*Collision) {
			if col.OtherRect.Type == Full {
				return space.Responses[RectSlide](goal, col, filter, tags...)
			}
			col.PreviousGoal = goal
			col.Normal = Vec2{0, 0}
			col.Touch.Y = goal.Y

			pivotLeft := goal.X + col.ItemRect.W*SlopePivot
			pivotRight := goal.X + col.ItemRect.W*(1-SlopePivot)
			switch col.OtherRect.Type {
			case TopRightSlope, TopLeftSlope:
				height := max(col.OtherRect.slopeHeight(pivotLeft), col.OtherRect.slopeHeight(pivotRight))
				if goal.Y < height {
					goal.Y = height
					col.Normal = Vec2{0, 1}
				}
			case BottomRightSlope, BottomLeftSlope:
				height := min(col.OtherRect.slopeHeight(pivotLeft), col.OtherRect.slopeHeight(pivotRight))
				if goal.Y > height-col.ItemRect.H {
					goal.Y = height - col.ItemRect.H
					col.Normal = Vec2{0, -1}
				}
			case Full:
				break
			}

			return goal, nil
		},
	}

	return space
}

func (s *Space) Set(item Item, rect Rect, tags ...Tag) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	oldRect, ok := s.rects[item]
	s.rects[item] = rect

	cells, oldCells := s.cellCoords(rect), s.cellCoords(oldRect)
	if slices.Equal(cells, oldCells) && (len(tags) == 0 || slices.Equal(tags, s.tags[item])) {
		return
	}
	if len(tags) > 0 {
		s.tags[item] = tags
	}
	for _, tag := range append(s.tags[item], "") {
		if ok {
			for _, cell := range oldCells {
				loc := location{tag, cell}
				if delete(s.searchSpace[loc], item); len(s.searchSpace[loc]) == 0 {
					delete(s.searchSpace, loc)
				}
			}
		}
		for _, cell := range cells {
			loc := location{tag, cell}
			if s.searchSpace[loc] == nil {
				s.searchSpace[loc] = map[Item]bool{}
			}
			s.searchSpace[loc][item] = true
		}
	}
}

func (s *Space) Rect(item Item) Rect {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.rects[item]
}

func (s *Space) Has(item Item, tags ...Tag) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, cell := range s.cellCoords(s.rects[item]) {
		for _, tag := range tags {
			if !s.searchSpace[location{tag, cell}][item] {
				return false
			}
		}
	}

	return true
}

func (s *Space) Remove(item Item) {
	if item == nil {
		return
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if rect, ok := s.rects[item]; ok {
		for _, cell := range s.cellCoords(rect) {
			for _, tag := range append(s.tags[item], "") {
				loc := location{tag, cell}
				if delete(s.searchSpace[loc], item); len(s.searchSpace[loc]) == 0 {
					delete(s.searchSpace, loc)
				}
			}
		}
	}
	delete(s.tags, item)
	delete(s.rects, item)
}

func (s *Space) Move(item Item, targetGoal Vec2, filter Filter, tags ...Tag) (Vec2, []*Collision) {
	goal, cols := s.Check(item, targetGoal, filter, tags...)
	rect := s.Rect(item)
	rect.X, rect.Y = goal.X, goal.Y
	s.Set(item, rect)

	return goal, cols
}

func (s *Space) Check(item Item, goal Vec2, filter Filter, tags ...Tag) (Vec2, []*Collision) {
	if filter == nil {
		filter = DefaultResponseFilter
	}

	visited := map[Item]bool{item: true}
	visitedFilter := func(item, other Item) (ColType, bool) {
		if visited[other] {
			return 0, false
		}

		return filter(item, other)
	}

	projectedCols := s.Project(item, s.Rect(item), goal, visitedFilter, tags...)
	var cols []*Collision
	for len(projectedCols) > 0 {
		col := projectedCols[0]
		visited[col.Other] = true
		goal, projectedCols = s.Responses[col.Type](goal, col, visitedFilter, tags...)
		cols = append(cols, col)
	}

	return goal, cols
}

func (s *Space) Project(item Item, rect Rect, goal Vec2, filter Filter, tags ...Tag) []*Collision {
	if filter == nil {
		filter = DefaultResponseFilter
	}
	if len(tags) == 0 {
		tags = []Tag{""}
	}
	s.mutex.RLock()
	var items []Item
	for _, cell := range s.cellCoords(rect) {
		for _, tag := range tags {
			for other := range s.searchSpace[location{tag, cell}] {
				if item == other {
					continue
				}
				items = append(items, other)
			}
		}
	}
	s.mutex.RUnlock()

	var cols []*Collision
	for _, other := range items {
		if responseName, ok := filter(item, other); ok {
			if col, ok := detectCollision(rect, s.Rect(other), goal); ok {
				col.Item, col.Other = item, other
				col.Type = responseName
				cols = append(cols, col)
			}
		}
	}
	slices.SortFunc(cols, func(a, b *Collision) int {
		if a.Intersection == b.Intersection {
			ir := a.ItemRect
			if rectSquareDistance(ir, a.OtherRect) < rectSquareDistance(ir, b.OtherRect) {
				return -1
			}

			return 1
		}
		if a.Intersection < b.Intersection {
			return -1
		}

		return 1
	})

	return cols
}

func (s *Space) Query(rect Rect, filter SelectFilter, tags ...Tag) []*Collision {
	if filter == nil {
		filter = func(_ Item) bool { return true }
	}
	projectFilter := func(_, other Item) (ColType, bool) { return 0, filter(other) }

	return s.Project(nil, rect, Vec2{rect.X, rect.Y}, projectFilter, tags...)
}

func Overlaps(r1, r2 Rect) bool {
	return rectContainsPoint(rectDiff(r1, r2), Vec2{})
}

func (s *Space) cellCoords(rect Rect) []cell {
	cx, cy := int(rect.X/s.cellSize), int(rect.Y/s.cellSize)
	cr, cb := int((rect.X+rect.W)/s.cellSize), int((rect.Y+rect.H)/s.cellSize)

	coords := make([]cell, 0, (cr+1-cx)*(cb+1-cy))
	for y := cy; y <= cb; y++ {
		for x := cx; x <= cr; x++ {
			coords = append(coords, cell{x, y})
		}
	}

	return coords
}

func (r Rect) slopeHeight(x float64) float64 {
	if r.Type == Full {
		return r.Y
	}
	lerp := math.Min(math.Max((x-r.X)/r.W, 0), 1)
	if r.Type == TopRightSlope || r.Type == BottomLeftSlope {
		return r.Y + lerp*r.H
	}

	return r.Y + (1-lerp)*r.H
}

// Liang-Barsky algorithm.
func lineSegmentIntersection(rect Rect, p1, p2 Vec2) (float64, float64, Vec2, bool) {
	dx, dy := p2.X-p1.X, p2.Y-p1.Y
	p := [4]float64{-dx, dx, -dy, dy} // left, right, top, bottom.
	q := [4]float64{p1.X - rect.X, rect.X + rect.W - p1.X, p1.Y - rect.Y, rect.Y + rect.H - p1.Y}
	n := [4]Vec2{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	i1, i2 := math.Inf(-1), math.Inf(1)
	normal := Vec2{}

	for i := range 4 {
		if p[i] == 0 {
			if q[i] <= 0 {
				return i1, i2, normal, false
			}

			continue
		}
		r := q[i] / p[i]
		if p[i] < 0 {
			if r > i2 {
				return i1, i2, normal, false
			} else if r > i1 {
				i1 = r
				normal = n[i]
			}
		} else {
			if r < i1 {
				return i1, i2, normal, false
			} else if r < i2 {
				i2 = r
			}
		}
	}

	return i1, i2, normal, true
}

func detectCollision(rect1, rect2 Rect, goal Vec2) (*Collision, bool) {
	col := &Collision{}
	col.Move = Vec2{goal.X - rect1.X, goal.Y - rect1.Y}
	col.ItemRect, col.OtherRect = rect1, rect2
	interRect := rectDiff(rect1, rect2)

	if !detectCollisionFirstPhase(interRect, rect1, col) {
		return col, false
	}
	if !col.Overlaps {
		return col, true
	}

	if (col.Move == Vec2{}) {
		p := rectNearestCorner(interRect, Vec2{})
		col.Normal = Vec2{math.Copysign(1, p.X), math.Copysign(1, p.Y)}
		if math.Abs(p.X) < math.Abs(p.Y) {
			p.Y = 0
			col.Normal.Y = 0
		} else {
			p.X = 0
			col.Normal.X = 0
		}
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

func detectCollisionFirstPhase(interRect, rect1 Rect, col *Collision) bool {
	collisioned := false
	if rectContainsPoint(interRect, Vec2{}) {
		collisioned = true
		p := rectNearestCorner(interRect, Vec2{})
		wi, hi := math.Min(rect1.W, math.Abs(p.X)), math.Min(rect1.H, math.Abs(p.Y))
		col.Intersection = -wi * hi
		col.Overlaps = true
	} else {
		i1, i2, normal, ok := lineSegmentIntersection(interRect, Vec2{}, col.Move)
		if ok && i1 < 1 && math.Abs(i1-i2) >= DELTA && (i1 > -DELTA || i1 == 0 && i2 > 0) {
			collisioned = true
			col.Normal = normal
			col.Intersection = i1
			col.Overlaps = false
			col.Touch = Vec2{rect1.X + col.Move.X*i1, rect1.Y + col.Move.Y*i1}
		}
	}

	return collisioned
}

// Minkowsky Difference between 2 Rects.
func rectDiff(r1, r2 Rect) Rect {
	return Rect{r2.X - r1.X - r1.W, r2.Y - r1.Y - r1.H, r1.W + r2.W, r1.H + r2.H, Full}
}

func rectContainsPoint(r Rect, p Vec2) bool {
	return p.X-r.X > DELTA && p.Y-r.Y > DELTA && r.X+r.W-p.X > DELTA && r.Y+r.H-p.Y > DELTA
}

func rectSquareDistance(r1, r2 Rect) float64 {
	dx := r1.X - r2.X + (r1.W-r2.W)/2
	dy := r1.Y - r2.Y + (r1.H-r2.H)/2

	return dx*dx + dy*dy
}

func rectNearestCorner(rect Rect, p Vec2) Vec2 {
	nearest := func(x, a, b float64) float64 {
		if math.Abs(a-x) < math.Abs(b-x) {
			return a
		}

		return b
	}

	return Vec2{nearest(p.X, rect.X, rect.X+rect.W), nearest(p.Y, rect.Y, rect.Y+rect.H)}
}
