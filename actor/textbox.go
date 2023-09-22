package actor

import (
	"game/libs/bump"
	"game/libs/camera"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	tbBoxX, tbDefaultBoxY              = 6.0, 30.0
	rbBoxMarginY, tbBoxMinY, tbBoxMaxY = 5, 25, 96 - tbBoxH - rbBoxMarginY
	tbBoxInnerW                        = 160
	tbBoxW, tbBoxH                     = tbBoxInnerW - tbBoxX*2, 15.0
	tbLineSize                         = (tbBoxW - 4) / 4
)

var (
	tbIndicatorImage, _, _ = ebitenutil.NewImageFromFile("assets/textboxindicator.png")
	tbBackgroundColor      = color.RGBA{34, 32, 52, 255}
)

type Textbox struct {
	Text   string
	Area   func() bump.Rect
	active bool
	camera *camera.Camera
}

func (c *Textbox) Init(a *Actor) {
	c.camera = a.World.Camera
	words := strings.Split(c.Text, " ")
	c.Text = ""
	i := 1
	for _, word := range words {
		if len(c.Text+word)+1 > (i*tbLineSize)-(i-1) {
			c.Text += "\n"
			i++
		}
		c.Text += " " + word
	}
}

func (c *Textbox) Update(a *Actor, dt float64) {
	c.active = len(a.Body.QueryActors(a, c.Area(), true)) > 0
}

/*func (c *Textbox) Draw(screen *ebiten.Image, _ ebiten.GeoM) {
	if !c.active {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(tbBoxX, 0)
	boxY := tbDefaultBoxY
	if c.Body != nil {
		cx, cy := c.camera.Position()
		x, y, w, _ := c.entity.Rect()
		boxY = y - cy - tbBoxH - rbBoxMarginY

		iop := &ebiten.DrawImageOptions{}
		iw, _ := tbIndicatorImage.Size()
		px, py := x+w/2-float64(iw)/2, tbBoxH
		ix := math.Max(math.Min(px-cx, tbBoxW+tbBoxX-float64(iw)), tbBoxX)
		iop.GeoM.Translate(ix, boxY+py)
		screen.DrawImage(tbIndicatorImage, iop)
	}

	op.GeoM.Translate(0, math.Max(math.Min(boxY, tbBoxMaxY), tbBoxMinY))
	screen.DrawImage(c.drawBackground(), op)

	op.GeoM.Translate(2, 2)
	// TODO: wrap text if falls out side text box
	utils.DrawText(screen, c.Text, assets.TinyFont, op)
}*/

func (c *Textbox) drawBackground() *ebiten.Image {
	bg := ebiten.NewImage(tbBoxW, tbBoxH)
	bg.Fill(tbBackgroundColor)

	return bg
}
