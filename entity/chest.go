package entity

import (
	"game/assets"
	"game/comps/hitbox"
	"game/comps/render"
	"game/core"
	"game/libs/bump"
	"game/vars"
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const chestW, chestH = 14, 9

var (
	chestImage, _, _                                    = ebitenutil.NewImageFromFileSystem(assets.FS, "chest.png")
	chestCloseImage, chestSemiOpenImage, chestOpenImage *ebiten.Image
)

func init() {
	imgSize := chestImage.Bounds().Size()
	chestCloseImage = chestImage.SubImage(image.Rect(0, imgSize.Y-chestH, imgSize.X, imgSize.Y)).(*ebiten.Image)
	chestSemiOpenImage = chestImage.SubImage(image.Rect(0, chestH, imgSize.X, 2*chestH+2)).(*ebiten.Image)
	chestOpenImage = chestImage.SubImage(image.Rect(0, 0, imgSize.X, chestH)).(*ebiten.Image)
}

type Chest struct {
	*core.BaseEntity
	render *render.Comp
	hitbox *hitbox.Comp
	reward int
	open   bool
}

func NewChest(x, y, _, _ float64, props *core.Properties) *Chest {
	y += tileSize - chestH
	imageOffset := 0.0
	if props.FlipX {
		imageOffset = chestW - tileSize*2
		x -= chestW - tileSize
	}
	chest := &Chest{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: chestW, H: chestH},
		render:     &render.Comp{X: imageOffset, Image: chestCloseImage, FlipX: props.FlipX, Layer: -1},
		hitbox:     &hitbox.Comp{},
		reward:     100,
		open:       props.Custom["open"] == "true",
	}
	chest.Add(chest.render, chest.hitbox)

	return chest
}

func (c *Chest) Init() {
	c.hitbox.HitFunc = c.chestHurt
	c.hitbox.PushHitbox(bump.Rect{W: chestW, H: chestH}, hitbox.Hit, nil)
	if c.open {
		c.Open(false)
	}
}

func (c *Chest) Update(_ float64) {}

func (c *Chest) Opened() bool { return c.open }

func (c *Chest) Open(reward bool) {
	c.open = true
	c.hitbox.Remove()
	c.render.Image = chestSemiOpenImage
	c.render.Y = chestH - float64(chestSemiOpenImage.Bounds().Dy())
	if !reward {
		return
	}
	time.AfterFunc(500*time.Millisecond, func() {
		c.render.Image = chestOpenImage
		c.render.Y = 0
		for i := 0; i < c.reward; i++ {
			vars.World.Add(NewFlake(c))
		}
	})
}

func (c *Chest) chestHurt(other core.Entity, _ *bump.Collision, _ float64, _ hitbox.ContactType) {
	if c.open || !core.GetFlag(other, vars.PlayerTeamFlag) {
		return
	}

	c.Open(true)
}
