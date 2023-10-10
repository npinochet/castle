package entity

import (
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/render"
	"game/core"
	"game/libs/bump"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const blockSize = 20

type BlockComponent struct {
	body   *body.Comp
	hitbox *hitbox.Comp
}

func (b *BlockComponent) Tag() string { return "BlockComponent" }

func NewBlock(x, y float64, props map[string]any) *core.Entity {
	w, _ := props["w"].(float64)
	h, _ := props["h"].(float64)
	image := ebiten.NewImage(int(w), int(h))
	image.Fill(color.White)

	block := &core.Entity{X: x, Y: y, W: 8, H: 8}
	body := &body.Comp{}
	hitbox := &hitbox.Comp{}
	block.AddComponent(body, hitbox, &render.Comp{Image: image}, &BlockComponent{body: body, hitbox: hitbox})

	return block
}

func (b *BlockComponent) Init(entity *core.Entity) {
	b.hitbox.HurtFunc = b.BlockHurt
	b.hitbox.BlockFunc = b.BlockBlock
	b.hitbox.PushHitbox(bump.Rect{X: entity.X, Y: entity.Y, W: blockSize, H: blockSize}, false)
}

func (b *BlockComponent) BlockHurt(other *core.Entity, col *bump.Collision, damage float64) {
	b.body.Vy -= 30

	force := 50.0
	if col.Normal.X > 0 {
		force *= -1
	}
	b.body.Vx += force
}

func (b *BlockComponent) BlockBlock(other *core.Entity, col *bump.Collision, damage float64) {
	force := 50.0
	if col.Normal.X > 0 {
		force *= -1
	}
	b.body.Vx += force
}
