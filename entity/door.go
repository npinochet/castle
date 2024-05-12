package entity

import (
	"game/assets"
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/render"
	"game/core"
	"game/ext"
	"game/libs/bump"
	"game/vars"
	"image"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const doorW, doorH = 3, tileSize * 3

var (
	doorImage, _, _               = ebitenutil.NewImageFromFileSystem(assets.FS, "door.png")
	doorCloseImage, doorOpenImage *ebiten.Image
)

func init() {
	doorCloseImage = doorImage.SubImage(image.Rect(0, 0, tileSize, doorH)).(*ebiten.Image)
	doorOpenImage = doorImage.SubImage(image.Rect(tileSize, 0, tileSize*2, doorH)).(*ebiten.Image)
}

type Door struct {
	*core.BaseEntity
	render         *render.Comp
	body           *body.Comp
	hitbox         *hitbox.Comp
	open           bool
	opensFromRight bool
}

func NewDoor(x, y, _, _ float64, props *core.Properties) *Door {
	imageOffset := 0.0
	if props.FlipX {
		imageOffset = -tileSize + doorW
		x -= imageOffset
	}
	h := doorH
	if size, _ := strconv.Atoi(props.Custom["size"]); size != 0 {
		h = tileSize * size
	}
	image := doorCloseImage.SubImage(image.Rect(0, 0, tileSize, h)).(*ebiten.Image)
	door := &Door{
		BaseEntity:     &core.BaseEntity{X: x, Y: y, W: doorW, H: float64(h)},
		render:         &render.Comp{X: imageOffset, Image: image, FlipX: props.FlipX},
		body:           &body.Comp{Solid: true, Tags: []bump.Tag{"solid"}},
		hitbox:         &hitbox.Comp{},
		opensFromRight: props.FlipX,
		open:           props.Custom["open"] == "true",
	}
	door.Add(door.render, door.body, door.hitbox)

	return door
}
func (d *Door) Priority() int { return -1 }

func (d *Door) Init() {
	d.hitbox.HitFunc = d.doorHurt
	d.hitbox.PushHitbox(bump.Rect{W: doorW, H: doorH}, hitbox.Hit, nil)
	if d.open {
		d.open = false
		d.Open()
	}
}

func (d *Door) Update(_ float64) {}

func (d *Door) Opened() bool { return d.open }

func (d *Door) Open() {
	if d.open {
		return
	}
	d.open = true
	d.body.Remove()
	d.hitbox.Remove()
	d.render.Image = doorOpenImage.SubImage(image.Rect(tileSize, 0, tileSize*2, int(d.H))).(*ebiten.Image)
	// Open subsequent doors.
	for _, door := range ext.QueryFront(d, tileSize, d.H, !d.opensFromRight) {
		door.Open()
	}
}

func (d *Door) Close() {
	if !d.open {
		return
	}
	d.open = false
	d.body.Init(d)
	d.hitbox.Init(d)
	d.render.Image = doorCloseImage.SubImage(image.Rect(0, 0, tileSize, int(d.H))).(*ebiten.Image)
	for _, door := range ext.QueryFront(d, tileSize, d.H, !d.opensFromRight) {
		door.Close()
	}
}

func (d *Door) doorHurt(other core.Entity, _ *bump.Collision, _ float64, _ hitbox.ContactType) {
	if d.open || !core.GetFlag(other, vars.PlayerTeamFlag) {
		return
	}
	ox, _ := other.Position()
	if d.opensFromRight && ox > d.X || !d.opensFromRight && ox < d.X {
		d.Open()
	}
}
