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
	indicatorImage, _, _ = ebitenutil.NewImageFromFile("assets/textboxindicator.png")
	backgroundColor      = color.RGBA{34, 32, 52, 255}
)

type Comp struct {
	Text      string
	Area      func() bump.Rect
	Indicator bool
	active    bool
	entity    core.Entity
	camera    *camera.Camera
}

func (c *Comp) Init(entity core.Entity) {
	c.entity = entity
	c.camera = vars.World.Camera
	words := strings.Split(c.Text, " ")
	c.Text = ""
	i := 1
	for _, word := range words {
		if len(c.Text+word)+1 > (i*vars.LineSize)-(i-1) {
			c.Text += "\n"
			i++
		}
		c.Text += " " + word
	}
}

func (c *Comp) Update(_ float64) {
	c.active = len(ext.QueryItems(c.entity, c.Area(), "body")) > 0
}

func (c *Comp) Draw(screen *ebiten.Image, _ ebiten.GeoM) {
	if !c.active {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(vars.BoxX, 0)
	boxY := vars.DefaultBoxY
	if c.Indicator {
		cx, cy := c.camera.Position()
		x, y, w, _ := c.entity.Rect()
		boxY = y - cy - vars.BoxH - vars.BoxMarginY

		iop := &ebiten.DrawImageOptions{}
		iw := indicatorImage.Bounds().Dx()
		px, py := x+w/2-float64(iw)/2, vars.BoxH
		ix := math.Max(math.Min(px-cx, vars.BoxW+vars.BoxX-float64(iw)), vars.BoxX)
		iop.GeoM.Translate(ix, boxY+py)
		screen.DrawImage(indicatorImage, iop)
	}

	op.GeoM.Translate(0, math.Max(math.Min(boxY, vars.BoxMaxY), vars.BoxMinY))
	screen.DrawImage(c.drawBackground(), op)

	op.GeoM.Translate(2, 2)
	// TODO: wrap text if falls out side text box
	utils.DrawText(screen, c.Text, assets.TinyFont, op)
}

func (c *Comp) drawBackground() *ebiten.Image {
	bg := ebiten.NewImage(vars.BoxW, vars.BoxH)
	bg.Fill(backgroundColor)

	return bg
}
