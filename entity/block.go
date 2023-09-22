package entity

import (
	"game/actor"
	"game/core"
	"game/libs/bump"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const blockSize = 20

type Block struct {
	core.CoreEntity
	actor.Body
	actor.Hitbox
	actor.Render
}

func NewBlock(x, y float64, props map[string]interface{}) *Block {
	w, _ := props["w"].(float64)
	h, _ := props["h"].(float64)
	image := ebiten.NewImage(int(w), int(h))
	image.Fill(color.White)

	return &Block{
		CoreEntity: core.CoreEntity{X: x, Y: y, W: 8, H: 8},
		Render:     actor.Render{Image: image},
	}
}

func (b *Block) Init() {
	b.Hitbox.HurtFunc = b.BlockHurt
	b.Hitbox.BlockFunc = b.BlockBlock
	b.Hitbox.PushHitbox(nil, bump.Rect{X: b.CoreEntity.X, Y: b.CoreEntity.Y, W: blockSize, H: blockSize}, false)
}

func (b *Block) BlockHurt(_ *actor.Actor, col *bump.Collision, _ float64) {
	b.Body.Vy -= 30

	force := 50.0
	if col.Normal.X > 0 {
		force *= -1
	}
	b.Body.Vx += force
}

func (b *Block) BlockBlock(_ *actor.Actor, col *bump.Collision, _ float64) {
	force := 50.0
	if col.Normal.X > 0 {
		force *= -1
	}
	b.Body.Vx += force
}
