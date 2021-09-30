package entity

import (
	"game/comp"
	"game/core"
	"game/libs/bump"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

func (b *BlockComponent) IsActive() bool        { return b.active }
func (b *BlockComponent) SetActive(active bool) { b.active = active }

type BlockComponent struct {
	active bool
	body   *comp.BodyComponent
	hitbox *comp.HitboxComponent
}

func NewBlock(x, y float64, props map[string]interface{}) *core.Entity {
	w, h := props["w"].(float64), props["h"].(float64)
	image := ebiten.NewImage(int(w), int(h))
	image.Fill(color.White)

	block := &core.Entity{X: x, Y: y}
	body := &comp.BodyComponent{W: 8, H: 8}
	hitbox := &comp.HitboxComponent{}
	block.AddComponent(body)
	block.AddComponent(hitbox)
	block.AddComponent(&comp.RenderComponent{Image: image})
	block.AddComponent(&BlockComponent{body: body, hitbox: hitbox})

	return block
}

func (b *BlockComponent) Init(entity *core.Entity) {
	b.hitbox.HurtFunc = b.BlockHurt
	b.hitbox.BlockFunc = b.BlockBlock
	b.hitbox.PushHitbox(b.body.X, b.body.X, 20, 20, false)
}

func (b *BlockComponent) BlockHurt(otherHc *comp.HitboxComponent, col bump.Colision, damage float64) {
	b.body.Vy -= 30

	force := 50.0
	if col.Normal.X > 0 {
		force *= -1
	}
	b.body.Vx += force
}

func (b *BlockComponent) BlockBlock(otherHc *comp.HitboxComponent, col bump.Colision, damage float64) {
	force := 50.0
	if col.Normal.X > 0 {
		force *= -1
	}
	b.body.Vx += force
}
