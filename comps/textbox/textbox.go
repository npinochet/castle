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

const flickerTime = 0.5

var (
	indicatorImage, _, _ = ebitenutil.NewImageFromFileSystem(assets.FS, "textboxindicator.png")
	advanceImage, _, _   = ebitenutil.NewImageFromFileSystem(assets.FS, "textboxadvance.png")
	backgroundColor      = color.RGBA{34, 32, 52, 255}
)

type Comp struct {
	Text                string
	Area                func() bump.Rect
	Indicator           bool
	active              bool
	entity              core.Entity
	camera              *camera.Camera
	lines               int
	boxH                int
	bgImage, textImage  *ebiten.Image
	advanceState        int
	advanceMax          int
	advanceFlicker      bool
	advanceFlickerTimer float64
}

func (c *Comp) Init(entity core.Entity) {
	c.entity = entity
	c.camera = vars.World.Camera
	text := c.Text
	text = strings.ReplaceAll(text, "\n\n", " \r ")
	text = strings.ReplaceAll(text, "\n", " \n ")

	var lines []string
	currentLine := ""
	advanceW := float64(advanceImage.Bounds().Size().X)
	for word := range strings.SplitSeq(text, " ") {
		newLines := 0
		if word == "\n" {
			newLines = 1
		}
		if word == "\r" {
			newLines = vars.MaxLines - (len(lines) % vars.MaxLines) - 1
		}
		if newLines > 0 {
			for range newLines {
				lines = append(lines, currentLine)
				currentLine = ""
			}

			continue
		}

		maxLineWidth := vars.LineWidth
		if len(lines)%vars.MaxLines == vars.MaxLines-1 {
			maxLineWidth -= advanceW
		}
		if w, _ := utils.TextSize(currentLine+" "+word, assets.NanoFont); w > maxLineWidth {
			lines = append(lines, currentLine)
			currentLine = ""
		}
		if len(currentLine) > 0 {
			currentLine += " "
		}
		currentLine += word
	}
	lines = append(lines, currentLine)
	c.Text = strings.Join(lines, "\n")
	c.lines = len(lines)
	c.advanceMax = int(math.Ceil(float64(c.lines)/vars.MaxLines)) - 1

	c.boxH = min(vars.MaxLines, c.lines)*vars.LineHeight + vars.BoxH
	c.bgImage = ebiten.NewImage(vars.BoxW, c.boxH)
	c.bgImage.Fill(backgroundColor)
	c.textImage = ebiten.NewImage(vars.BoxW, c.boxH-3)
}

func (c *Comp) Remove() {}

func (c *Comp) Update(dt float64) {
	active := false
	for _, e := range ext.QueryItems(c.entity, c.Area(), "body") {
		if core.GetFlag(e, vars.PlayerTeamFlag) {
			active = true

			break
		}
	}
	if active {
		if c.advanceFlickerTimer += dt; c.advanceFlickerTimer > flickerTime {
			c.advanceFlickerTimer = 0
			c.advanceFlicker = !c.advanceFlicker
		}
		if vars.Pad.KeyPressed(utils.KeyDown) && c.advanceState < c.advanceMax {
			c.advanceState++
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
		boxY = c.drawIndicator(pipeline)
	}

	op.GeoM.Translate(0, max(min(boxY, vars.BoxMaxY), vars.BoxMinY))
	normalOp := &ebiten.DrawImageOptions{GeoM: op.GeoM, Blend: ebiten.BlendDestinationOut}
	pipeline.Add(vars.PipelineNormalMapTag, vars.PipelineUILayer, func(normalMap *ebiten.Image) {
		normalMap.DrawImage(c.bgImage, normalOp)
	})
	c.textImage.Fill(color.Transparent)
	textOp := &ebiten.DrawImageOptions{}
	textOp.GeoM.Translate(4, -float64((c.boxH-3)*c.advanceState))
	utils.DrawText(c.textImage, c.Text, assets.NanoFont, textOp)

	textOnBGOp := &ebiten.DrawImageOptions{}
	textOnBGOp.GeoM.Translate(0, 2)
	pipeline.Add(vars.PipelineScreenTag, vars.PipelineUILayer, func(screen *ebiten.Image) {
		c.bgImage.Fill(backgroundColor)
		c.bgImage.DrawImage(c.textImage, textOnBGOp)
		screen.DrawImage(c.bgImage, op)
		if !c.advanceFlicker && c.advanceState < c.advanceMax {
			advanceOp := &ebiten.DrawImageOptions{GeoM: op.GeoM}
			advanceSize := advanceImage.Bounds().Size()
			advanceOp.GeoM.Translate(
				vars.BoxW-float64(advanceSize.X)-vars.BoxH,
				min(vars.MaxLines, float64(c.lines))*vars.LineHeight-float64(advanceSize.Y),
			)
			screen.DrawImage(advanceImage, advanceOp)
		}
	})
}

func (c *Comp) drawIndicator(pipeline *core.Pipeline) float64 {
	boxH := float64(vars.BoxH + min(vars.MaxLines, c.lines)*vars.LineHeight)
	cx, cy := c.camera.Position()
	x, y, w, _ := c.entity.Rect()
	boxY := y - cy - vars.BoxMarginY - boxH - 1
	iop := &ebiten.DrawImageOptions{}
	iw := float64(indicatorImage.Bounds().Size().X)
	px, py := x+w/2-iw/2, boxH
	ix := max(min(px-cx, vars.BoxW+vars.BoxX-iw), vars.BoxX)
	iop.GeoM.Translate(ix, boxY+py)
	pipeline.Add(vars.PipelineScreenTag, vars.PipelineUILayer, func(screen *ebiten.Image) { screen.DrawImage(indicatorImage, iop) })

	normalOp := &ebiten.DrawImageOptions{GeoM: iop.GeoM, Blend: ebiten.BlendDestinationOut}
	pipeline.Add(vars.PipelineNormalMapTag, vars.PipelineUILayer, func(normalMap *ebiten.Image) {
		normalMap.DrawImage(indicatorImage, normalOp)
	})

	return boxY
}

func (c *Comp) NewText(text string) {
	c.Text = text
	c.advanceState = 0
	c.Init(c.entity)
}
