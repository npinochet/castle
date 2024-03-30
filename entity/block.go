package entity

import (
	"game/comps/body"
	"game/comps/hitbox"
	"game/core"
	"game/libs/bump"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const blockSize = 20

type Block struct {
	*core.BaseEntity
	body   *body.Comp
	hitbox *hitbox.Comp
}

func NewBlock(x, y, w, h float64, props *core.Properties) *Block {
	image := ebiten.NewImage(int(w), int(h))
	image.Fill(color.White)

	block := &Block{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: 8, H: 8},
		body:       &body.Comp{},
		hitbox:     &hitbox.Comp{},
	}
	block.Add(block.body, block.hitbox)

	return block
}

func (b *Block) Init() {
	b.hitbox.HitFunc = func(other core.Entity, col *bump.Collision, damage float64, contactType hitbox.ContactType) {
		if contactType == hitbox.Hit {
			b.body.Vy -= 30
		}
		force := 50.0
		if col.Normal.X > 0 {
			force *= -1
		}
		b.body.Vx += force
	}
	b.hitbox.PushHitbox(bump.Rect{X: b.X, Y: b.Y, W: blockSize, H: blockSize}, hitbox.Hit, nil)
}
