package textbox

import (
	"game/assets"
	"game/comps/body"
	"game/core"
	"game/libs/bump"
	"game/libs/camera"
	"game/utils"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	boxX, defaultBoxY            = 6.0, 30.0
	boxMarginY, boxMinY, boxMaxY = 5, 25, 96 - boxH - boxMarginY
	boxW, boxH                   = 160 - boxX*2, 15.0
)

var (
	indicatorImage, _, _ = ebitenutil.NewImageFromFile("assets/textboxindicator.png")
	backgroundColor      = color.RGBA{34, 32, 52, 255}
)

type Comp struct {
	Text   string
	Body   *body.Comp
	Area   func() bump.Rect
	active bool
	camera *camera.Camera
}

func (c *Comp) Init(entity *core.Entity) {
	c.camera = entity.World.Camera
}

func (c *Comp) Update(dt float64) {
	c.active = len(c.Body.QueryRect(c.Area())) > 0
}

func (c *Comp) Draw(screen *ebiten.Image, _ ebiten.GeoM) {
	if !c.active {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(boxX, 0)
	boxY := defaultBoxY
	if c.Body != nil {
		cx, cy := c.camera.Position()
		rect := c.Body.Rect()
		boxY = rect.Y - cy - boxH - boxMarginY

		iop := &ebiten.DrawImageOptions{}
		w, _ := indicatorImage.Size()
		px, py := rect.X+rect.W/2-float64(w)/2, boxH
		ix := math.Max(math.Min(px-cx, boxW+boxX-float64(w)), boxX)
		iop.GeoM.Translate(ix, boxY+py)
		screen.DrawImage(indicatorImage, iop)
	}

	op.GeoM.Translate(0, math.Max(math.Min(boxY, boxMaxY), boxMinY))
	screen.DrawImage(c.drawBackground(), op)

	op.GeoM.Translate(2, 2)
	// TODO: wrap text if falls out side text box
	utils.DrawText(screen, c.Text, assets.TinyFont, op)
}

func (c *Comp) drawBackground() *ebiten.Image {
	bg := ebiten.NewImage(boxW, boxH)
	bg.Fill(backgroundColor)

	return bg
}
