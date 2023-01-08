package entity

import (
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/render"
	"game/core"
	"game/libs/bump"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	rockSize               = 5
	rockDamage             = 5
	rockWeight             = -0.4
	rockMinVel, rockMaxVel = 50.0, 100.0
)

var (
	rockImageFile = "assets/rock.png"
	rockImage     *ebiten.Image
)

type Rock struct {
	core.Entity
	render *render.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	owner  *Actor
}

func init() {
	var err error
	rockImage, _, err = ebitenutil.NewImageFromFile(rockImageFile)
	if err != nil {
		panic(err)
	}
}

func NewRock(x, y float64, owner *Actor) *core.Entity {
	vx, vy := rockMaxVel, 60.0
	if target := owner.AI.Target; target != nil {
		tx, _ := target.Position()
		vx = calculateVx(x, tx, vy)
	}

	body := &body.Comp{
		W: rockSize, H: rockSize,
		Weight: rockWeight,
		Vx:     vx, Vy: -vy,
		MaxX:      rockMaxVel,
		FilterOut: []*body.Comp{owner.Body},
	}
	rock := &Rock{
		Entity: core.Entity{X: x, Y: y},
		render: &render.Comp{Image: rockImage},
		body:   body,
		hitbox: &hitbox.Comp{},
		owner:  owner,
	}
	rock.AddComponent(rock.render, body, rock.hitbox, rock)

	return &rock.Entity
}

func (r *Rock) Init(entity *core.Entity) {
	r.body.Friction = false
	r.hitbox.HurtFunc, r.hitbox.BlockFunc = r.RockHurt, r.RockHurt
	r.hitbox.PushHitbox(bump.Rect{X: r.body.X, Y: r.body.X, W: rockSize, H: rockSize}, false)
}

func (r *Rock) Update(dt float64) {
	hb := bump.Rect{X: r.body.X, Y: r.body.Y, H: rockSize, W: rockSize}
	_, contacted := r.hitbox.HitFromHitBox(hb, rockDamage, []*hitbox.Comp{r.owner.Hitbox})
	if len(contacted) > 1 || r.body.Ground {
		r.Remove()
	}
}

func (r *Rock) RockHurt(otherHc *hitbox.Comp, _ *bump.Collision, _ float64) {
	if otherHc != r.owner.Hitbox {
		r.Remove()
	}
}

func (r *Rock) Remove() {
	if r.body != nil {
		r.World.Space.Remove(r.body)
	}
	if r.hitbox != nil {
		for r.hitbox.PopHitbox() != nil {
		}
	}
	r.World.RemoveEntity(r.ID)
}

func calculateVx(x, tx, vy float64) float64 {
	widthBuffer := 10.0
	dx := math.Abs(x - tx - widthBuffer)
	a := body.Gravity * (rockWeight + 1)
	t := vy / a
	vx := math.Max(dx/(2*t), rockMinVel)
	if x > tx {
		vx *= -1
	}

	return vx
}
