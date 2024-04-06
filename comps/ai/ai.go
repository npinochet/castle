package ai

import (
	"game/assets"
	"game/core"
	"game/libs/bump"
	"game/utils"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

var DebugDraw = false

type Action struct {
	Name        string
	Next        func(dt float64) bool
	Entry, Exit func()
}

type actionItem struct {
	action *Action
	timer  float64
}

type Comp struct {
	core.Entity
	Target      core.Entity
	act         func()
	actionQueue []actionItem
	DebugRect   *bump.Rect
}

func (c *Comp) Init(entity core.Entity) {
	c.Entity = entity
}

func (c *Comp) SetAct(act func()) { c.act = act }

func (c *Comp) Add(timeout float64, action *Action) {
	if timeout <= 0 {
		timeout = math.MaxFloat64
	}
	c.actionQueue = append(c.actionQueue, actionItem{action, timeout})
	if len(c.actionQueue) == 1 && action.Entry != nil {
		action.Entry()
	}
}

func (c *Comp) Update(dt float64) {
	if len(c.actionQueue) == 0 {
		if c.act != nil {
			c.act()
		}
		if len(c.actionQueue) == 0 {
			return
		}
	}
	item := &c.actionQueue[0]
	item.timer -= dt
	if item.timer <= 0 || item.action.Next(dt) {
		if item.action.Exit != nil {
			item.action.Exit()
		}
		if c.actionQueue = c.actionQueue[1:]; len(c.actionQueue) > 0 {
			if nextItem := c.actionQueue[0]; nextItem.action.Entry != nil {
				nextItem.action.Entry()
			}
		}
		c.Update(dt)
	}
}

func (c *Comp) Draw(screen *ebiten.Image, entityPos ebiten.GeoM) {
	if !DebugDraw || len(c.actionQueue) == 0 {
		return
	}
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-5, -10)
	utils.DrawText(screen, "AI:"+c.actionQueue[0].action.Name, assets.TinyFont, op)
	if c.DebugRect != nil {
		image := ebiten.NewImage(int(c.DebugRect.W), int(c.DebugRect.H))
		image.Fill(color.NRGBA{255, 255, 0, 75})
		op := &ebiten.DrawImageOptions{GeoM: entityPos}
		op.GeoM.Translate(-c.DebugRect.X, -c.DebugRect.Y)
		screen.DrawImage(image, op)
	}
}

func (c *Comp) InTargetRange(minDist, maxDist float64) bool {
	if c.Target == nil {
		return false
	}

	x, y := c.Position()
	tx, ty := c.Target.Position()
	dist := utils.Distante(x, y, tx, ty)

	in := dist >= minDist
	out := true
	if maxDist > 0 {
		out = dist <= maxDist
	}

	return in && out
}
