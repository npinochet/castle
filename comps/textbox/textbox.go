package textbox

import (
	"game/assets"
	"game/core"
	"game/ext"
	"game/libs/bump"
	"game/libs/camera"
	"game/utils"
	"game/vars"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	indicatorImage, _, _ = ebitenutil.NewImageFromFileSystem(assets.FS, "textboxindicator.png")
	backgroundColor      = color.RGBA{34, 32, 52, 255}
)

type Comp struct {
	Text      string
	Area      func() bump.Rect
	Indicator bool
	active    bool
	entity    core.Entity
	camera    *camera.Camera
	lines     int
}

func (c *Comp) Init(entity core.Entity) {
	c.entity = entity
	c.camera = vars.World.Camera
	words := strings.Split(c.Text, " ")
	c.Text = ""
	c.lines = 1
	lineWidth := 0
	for _, word := range words {
		if lineWidth+len(word)+1 > vars.LineWidth || strings.Contains(word, "\n") {
			c.Text += "\n"
			c.lines++
			lineWidth = 0
		}
		normalized := strings.ReplaceAll(word, "\n", "")
		c.Text += " " + normalized
		lineWidth += len(normalized) + 1
	}
}

func (c *Comp) Remove() {}

func (c *Comp) Update(_ float64) {
	active := false
	for _, e := range ext.QueryItems(c.entity, c.Area(), "body") {
		if core.GetFlag(e, vars.PlayerTeamFlag) {
			active = true

			break
		}
	}
	c.active = active
}

func (c *Comp) Draw(pipeline *core.Pipeline, _ ebiten.GeoM) {
	if !c.active {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(vars.BoxX, 0)
	boxY := vars.DefaultBoxY
	if c.Indicator {
		boxH := vars.BoxH + float64(c.lines)*vars.LineHeight
		cx, cy := c.camera.Position()
		x, y, w, _ := c.entity.Rect()
		boxY = y - cy - vars.BoxMarginY - boxH
		iop := &ebiten.DrawImageOptions{}
		iw := indicatorImage.Bounds().Size().X
		px, py := x+w/2-float64(iw)/2, boxH
		ix := math.Max(math.Min(px-cx, vars.BoxW+vars.BoxX-float64(iw)), vars.BoxX)
		iop.GeoM.Translate(ix, boxY+py)
		pipeline.Add(vars.PipelineScreenTag, vars.PipelineUILayer, func(screen *ebiten.Image) { screen.DrawImage(indicatorImage, iop) })

		normalOp := &ebiten.DrawImageOptions{GeoM: iop.GeoM, Blend: ebiten.BlendDestinationOut}
		pipeline.Add(vars.PipelineNormalMapTag, vars.PipelineUILayer, func(normalMap *ebiten.Image) {
			normalMap.DrawImage(indicatorImage, normalOp)
		})
	}

	op.GeoM.Translate(0, math.Max(math.Min(boxY, vars.BoxMaxY), vars.BoxMinY))
	normalOp := &ebiten.DrawImageOptions{GeoM: op.GeoM, Blend: ebiten.BlendDestinationOut}
	pipeline.Add(vars.PipelineNormalMapTag, vars.PipelineUILayer, func(normalMap *ebiten.Image) {
		normalMap.DrawImage(c.drawBackground(), normalOp)
	})
	pipeline.Add(vars.PipelineScreenTag, vars.PipelineUILayer, func(screen *ebiten.Image) {
		screen.DrawImage(c.drawBackground(), op)
		op.GeoM.Translate(0, 2)
		utils.DrawText(screen, c.Text, assets.TinyFont, op)
	})
}

func (c *Comp) drawBackground() *ebiten.Image {
	bg := ebiten.NewImage(vars.BoxW, vars.BoxH+c.lines*vars.LineHeight)
	bg.Fill(backgroundColor)

	return bg
}
