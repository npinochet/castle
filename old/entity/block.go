package entity

import (
	"game/comps/basic/body"
	"game/comps/basic/hitbox"
	"game/comps/basic/render"
	"game/core"
	"game/libs/bump"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const blockSize = 20

type Block struct {
	body   *body.Comp
	hitbox *hitbox.Comp
}

func NewBlock(x, y float64, props map[string]any) *core.Entity {
	w, _ := props["w"].(float64)
	h, _ := props["h"].(float64)
	image := ebiten.NewImage(int(w), int(h))
	image.Fill(color.White)

	block := &core.Entity{X: x, Y: y, W: 8, H: 8}
	body := &body.Comp{}
	hitbox := &hitbox.Comp{}
	block.AddComponent(body, hitbox, &render.Comp{Image: image}, &Block{body: body, hitbox: hitbox})

	return block
}

func (b *Block) Init(entity *core.Entity) {
	b.hitbox.HurtFunc = b.BlockHurt
	b.hitbox.BlockFunc = b.BlockBlock
	b.hitbox.PushHitbox(bump.Rect{X: entity.X, Y: entity.Y, W: blockSize, H: blockSize}, false)
}

func (b *Block) BlockHurt(_ *core.Entity, col *bump.Collision, _ float64) {
	b.body.Vy -= 30

	force := 50.0
	if col.Normal.X > 0 {
		force *= -1
	}
	b.body.Vx += force
}

func (b *Block) BlockBlock(_ *core.Entity, col *bump.Collision, _ float64) {
	force := 50.0
	if col.Normal.X > 0 {
		force *= -1
	}
	b.body.Vx += force
}
