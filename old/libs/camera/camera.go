package camera

import (
	"game/libs/bump"
	"image"
	"math"
	"math/rand"

	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

const (
	defaultTransitionDuration = 0.8
	defaultStiffness          = 9
	heightJitterBuffer        = 8 // Prevent camera from snapping to celling on transition (not sure how it works).
)

type Recter interface {
	Rect() (float64, float64, float64, float64)
}

type Camera struct {
	x, y, w, h               float64
	following                Recter
	shakeTween               *gween.Tween
	shakeMagnitude           float64
	borders                  *bump.Rect
	rooms                    []bump.Rect
	transitionTween          *gween.Tween
	transitionX, transitionY float64
	stiffness                int
	transitionDuration       float32
	betweenRooms             bool
}

func New(w, h float64) *Camera {
	return &Camera{w: w, h: h, transitionDuration: defaultTransitionDuration, stiffness: defaultStiffness}
}

func (c *Camera) Position() (float64, float64) { return c.x, c.y }
func (c *Camera) SetPosition(x, y float64)     { c.x, c.y = x, y }
func (c *Camera) SetRooms(rooms []bump.Rect)   { c.rooms = rooms }
func (c *Camera) Follow(e Recter) {
	c.following = e
	c.SetRoomBorders(false)
}

func (c *Camera) Translate(x, y float64) {
	c.x += x
	c.y += y
}

func (c *Camera) Bounds() image.Rectangle {
	min := image.Point{int(c.x), int(c.y)}
	max := image.Point{int(math.Max(c.x, 0) + c.w), int(math.Max(c.y, 0) + c.h)}

	return image.Rectangle{min, max}
}

func (c *Camera) Update(dt float64) {
	if c.following == nil {
		return
	}

	ex, ey, w, h := c.following.Rect()
	x, y := ex+w/2-c.w/2, ey+h/2-c.h/2
	dx, dy := x-c.x, y-c.y

	c.Translate(damper(dt, dx, dy, c.stiffness))
	c.SetRoomBorders(true)
	if c.borders != nil {
		x := math.Max(math.Min(c.x, c.borders.X+c.borders.W-c.w), c.borders.X)
		y := math.Max(math.Min(c.y, c.borders.Y+c.borders.H-c.h), c.borders.Y)
		c.SetPosition(x, y)
	}
	if c.transitionTween != nil {
		prog, done := c.transitionTween.Update(float32(dt))
		if done {
			c.transitionTween = nil
		}
		c.Translate(float64(prog)*c.transitionX, float64(prog)*c.transitionY)
	}
	if c.shakeTween != nil {
		prog, done := c.shakeTween.Update(float32(dt))
		if done {
			c.shakeTween = nil
		}

		shakex := (rand.Float64()*2 - 1) * c.shakeMagnitude * float64(prog)
		shakey := (rand.Float64()*2 - 1) * c.shakeMagnitude * float64(prog)
		c.Translate(shakex, shakey)
	}
}

func (c *Camera) Shake(duration float32, magnitude float64) {
	if c.shakeTween != nil {
		return
	}
	c.shakeTween = gween.New(1, 0, duration, ease.Linear)
	c.shakeMagnitude = magnitude
}

func (c *Camera) SetRoomBorders(transition bool) {
	if c.following == nil || c.rooms == nil {
		return
	}

	x, y, w, h := c.following.Rect()
	follow := bump.Rect{X: x + w/4, Y: y, W: w / 2, H: h}

	prevRoom := c.borders
	roomCount := 0
	for i, room := range c.rooms {
		if bump.Overlaps(follow, room) {
			roomCount++
			if &c.rooms[i] != prevRoom {
				c.borders = &c.rooms[i]
			}
		}
	}
	if c.betweenRooms {
		c.borders = prevRoom
	}
	c.betweenRooms = roomCount > 1
	if roomCount == 0 {
		c.borders = nil
	}

	if transition && prevRoom != c.borders && c.borders != nil {
		targetX := math.Max(math.Min(c.x, c.borders.X+c.borders.W-c.w), c.borders.X)
		targetY := math.Max(math.Min(c.y, c.borders.Y+c.borders.H-c.h+heightJitterBuffer), c.borders.Y)
		c.transitionX, c.transitionY = c.x-targetX, c.y-targetY
		c.transitionTween = gween.New(1, 0, c.transitionDuration, ease.OutCubic)
	}
}

func damper(dt, dx, dy float64, stiffness int) (float64, float64) {
	dts := dt * float64(stiffness)

	return dx * dts, dy * dts
}
