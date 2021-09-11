package core

import (
	"game/libs/bump"
	"image"
	"math"
	"math/rand"

	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

func damper(dx, dy float64, stiffness int) (float64, float64) {
	dts := (1.0 / 60) * float64(stiffness)
	return dx * dts, dy * dts
}

type Camera struct {
	x, y, w, h               float64
	following                *Entity
	fw, fh                   float64
	shakeTween               *gween.Tween
	shakeMagnitude           float64
	borders                  *bump.Rect
	rooms                    []bump.Rect
	transitionTween          *gween.Tween
	transitionX, transitionY float64
	transitionDuration       float32
}

func NewCamera(w, h float64, stiffness int) *Camera {
	return &Camera{w: w, h: h, transitionDuration: 0.8}
}

func (c *Camera) Position() (float64, float64) { return c.x, c.y }
func (c *Camera) SetPosition(x, y float64)     { c.x, c.y = x, y }
func (c *Camera) SetRooms(rooms []bump.Rect)   { c.rooms = rooms }
func (c *Camera) Follow(e *Entity, w, h float64) {
	c.following = e
	c.fw, c.fh = w, h
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

func (c *Camera) Update(dt float64, stiffness int) {
	if c.following == nil {
		return
	}

	x, y := c.following.X+c.fw/2-c.w/2, c.following.Y+c.fh/2-c.h/2
	dx, dy := x-c.x, y-c.y

	c.Translate(damper(dx, dy, stiffness))
	c.SetRoomBorders()
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
		c.SetPosition(c.x+float64(prog)*c.transitionX, c.y+float64(prog)*c.transitionY)
	}

	if c.shakeTween != nil {
		prog, done := c.shakeTween.Update(float32(dt))
		if done {
			c.shakeTween = nil
		}

		shakex := (rand.Float64()*2 - 1) * c.shakeMagnitude * float64(prog)
		shakey := (rand.Float64()*2 - 1) * c.shakeMagnitude * float64(prog)
		c.SetPosition(c.x+shakex, c.y+shakey)
	}
}

func (c *Camera) Shake(duration float32, magnitude float64) {
	if c.shakeTween != nil {
		return
	}
	c.shakeTween = gween.New(1, 0, duration, ease.Linear)
	c.shakeMagnitude = magnitude
}

func (c *Camera) SetRoomBorders() {
	if c.following == nil || c.rooms == nil {
		return
	}

	x, y := c.following.X+c.fw/2, c.following.Y+c.fh/2
	follow := bump.Rect{X: x, Y: y, W: c.fw, H: c.fh}

	currentRoom, found := c.borders, false
	for i, room := range c.rooms {
		if bump.Overlaps(follow, room) {
			c.borders = &c.rooms[i]
			found = true
			break
		}
	}
	if !found {
		c.borders = nil
		return
	}

	if currentRoom != c.borders && currentRoom != nil && c.borders != nil {
		c.transitionX, c.transitionY = currentRoom.X-c.borders.X, currentRoom.Y-c.borders.Y
		c.transitionTween = gween.New(1, 0, c.transitionDuration, ease.OutCubic)
	}
}
