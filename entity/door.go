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

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const doorW = 3

var doorImage, _, _ = ebitenutil.NewImageFromFileSystem(assets.FS, "door.png")

type Door struct {
	*core.BaseEntity
	render         *render.Comp
	body           *body.Comp
	hitbox         *hitbox.Comp
	open           bool
	opensFromRight bool
}

func NewDoor(x, y, _, h float64, props *core.Properties) *Door {
	imageOffset := 0.0
	if props.FlipX {
		imageOffset = -tileSize + doorW
		x -= imageOffset
	}
	door := &Door{
		BaseEntity:     &core.BaseEntity{X: x, Y: y, W: doorW, H: h},
		render:         &render.Comp{X: imageOffset, FlipX: props.FlipX, Layer: -1},
		body:           &body.Comp{Solid: true, Tags: []bump.Tag{"solid"}},
		hitbox:         &hitbox.Comp{},
		opensFromRight: props.FlipX,
		open:           props.Custom["open"] == "true",
	}
	door.render.Image = door.image()
	door.Add(door.render, door.body, door.hitbox)

	return door
}

func (d *Door) Init() {
	d.hitbox.HitFunc = d.doorHurt
	d.hitbox.PushHitbox(bump.Rect{W: doorW, H: d.H}, hitbox.Hit, nil)
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
	d.render.Image = d.image()
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
	d.render.Image = d.image()
	// Close subsequent doors.
	for _, door := range ext.QueryFront(d, tileSize, d.H, !d.opensFromRight) {
		door.Close()
	}
}

func (d *Door) image() *ebiten.Image {
	img := ebiten.NewImage(tileSize, int(d.H))
	w := 0
	if d.open {
		w = tileSize
	}
	doorTop := doorImage.SubImage(image.Rect(w, 0, w+tileSize, tileSize)).(*ebiten.Image)
	doorFill := doorImage.SubImage(image.Rect(w, 2*tileSize, w+tileSize, 3*tileSize)).(*ebiten.Image)
	doorBase := doorImage.SubImage(image.Rect(w, tileSize, w+tileSize, 3*tileSize)).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	img.DrawImage(doorTop, op)
	op.GeoM.Translate(0, tileSize)

	for size := d.H - tileSize; size > 2*tileSize; size -= tileSize {
		img.DrawImage(doorFill, op)
		op.GeoM.Translate(0, tileSize)
	}
	img.DrawImage(doorBase, op)

	return img
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
