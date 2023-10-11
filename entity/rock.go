package entity

import (
	"game/comps/basic/body"
	"game/comps/basic/hitbox"
	"game/comps/basic/render"
	"game/core"
	"game/entity/defaults"
	"game/libs/bump"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	rockSize               = 5
	rockDamage             = 5
	rockWeight             = -0.4
	rockMinVel, rockMaxVel = 50.0, 100.0
	rockRollingTime        = 200 * time.Millisecond
)

var (
	rockImageFile = "assets/rock.png"
	rockImage     *ebiten.Image
)

type Rock struct {
	*core.Entity
	render *render.Comp
	body   *body.Comp
	hitbox *hitbox.Comp
	owner  *defaults.Actor
}

func init() {
	var err error
	rockImage, _, err = ebitenutil.NewImageFromFile(rockImageFile)
	if err != nil {
		panic(err)
	}
}

func NewRock(x, y float64, owner *defaults.Actor) *core.Entity {
	vx, vy := rockMaxVel, 60.0
	if target := owner.AI.Target; target != nil {
		tx, ty := target.Position()
		vx = calculateVx(x, y, tx, ty, vy)
	}

	body := &body.Comp{
		Weight: rockWeight,
		Vx:     vx, Vy: -vy,
		MaxX:      rockMaxVel,
		FilterOut: []*body.Comp{owner.Body},
	}
	rock := &Rock{
		Entity: &core.Entity{X: x, Y: y, W: rockSize, H: rockSize},
		render: &render.Comp{Image: rockImage, RollingTime: rockRollingTime},
		body:   body,
		hitbox: &hitbox.Comp{},
		owner:  owner,
	}
	rock.AddComponent(rock.render, body, rock.hitbox, rock)

	return rock.Entity
}

func (r *Rock) Init(entity *core.Entity) {
	r.body.Friction = false
	r.hitbox.HurtFunc, r.hitbox.BlockFunc = r.RockHurt, r.RockHurt
	r.hitbox.PushHitbox(bump.Rect{X: r.X, Y: r.X, W: rockSize, H: rockSize}, false)
}

func (r *Rock) Update(dt float64) {
	_, contacted := r.hitbox.HitFromHitBox(bump.Rect{H: rockSize, W: rockSize}, rockDamage, []*hitbox.Comp{r.owner.Hitbox})
	if len(contacted) > 1 || r.body.Ground {
		r.Remove()
	}
}

func (r *Rock) RockHurt(other *core.Entity, _ *bump.Collision, _ float64) {
	if other != r.owner.Control.Entity {
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

func calculateVx(x, y, tx, ty, vy float64) float64 {
	widthBuffer := 10.0
	dx := math.Abs(x - tx - widthBuffer)
	dy := math.Max(ty-y, 0)
	a := body.Gravity * (rockWeight + 1)
	t := vy / a
	t += dy / vy
	vx := math.Max(dx/(2*t), rockMinVel)
	if x > tx {
		vx *= -1
	}

	return vx
}
