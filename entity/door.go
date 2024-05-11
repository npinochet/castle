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

const doorW, doorH = 3, 8 * 3

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
	door := &Door{
		BaseEntity:     &core.BaseEntity{X: x, Y: y, W: doorW, H: doorH},
		render:         &render.Comp{X: imageOffset, Image: doorCloseImage, FlipX: props.FlipX},
		body:           &body.Comp{Solid: true, Tags: []bump.Tag{"solid"}},
		hitbox:         &hitbox.Comp{},
		opensFromRight: props.FlipX,
		open:           props.Custom["open"] == "true",
	}
	door.Add(door.render, door.body, door.hitbox)

	return door
}
func (r *Door) Priority() int { return -1 }

func (r *Door) Init() {
	r.hitbox.HitFunc = r.doorHurt
	r.hitbox.PushHitbox(bump.Rect{W: doorW, H: doorH}, hitbox.Hit, nil)
	if r.open {
		r.Open()
	}
}

func (r *Door) Update(_ float64) {}

func (r *Door) Open() {
	r.open = true
	r.body.Remove()
	r.hitbox.Remove()
	r.render.Image = doorOpenImage
	// Open subsequent doors.
	for _, door := range ext.QueryFront(r, tileSize, doorH, !r.opensFromRight) {
		door.Open()
	}
}

func (r *Door) doorHurt(other core.Entity, _ *bump.Collision, _ float64, _ hitbox.ContactType) {
	if r.open || !core.GetFlag(other, vars.PlayerTeamFlag) {
		return
	}
	ox, _ := other.Position()
	if r.opensFromRight && ox > r.X || !r.opensFromRight && ox < r.X {
		r.Open()
	}
}
