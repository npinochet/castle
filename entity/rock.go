package entity

import (
	"game/actor"
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
	actor.Actor
	actor.Render
	owner *actor.Actor
}

func init() {
	var err error
	rockImage, _, err = ebitenutil.NewImageFromFile(rockImageFile)
	if err != nil {
		panic(err)
	}
}

func NewRock(x, y float64, owner *actor.Actor) *Rock {
	vx, vy := rockMaxVel, 60.0
	if target := owner.AI.Target; target != nil {
		tx, ty := target.Position()
		vx = calculateVx(x, y, tx, ty, vy)
	}

	rock := &Rock{
		Actor:  actor.NewActor(x, y, rockSize, rockSize, nil),
		Render: actor.Render{Image: rockImage, RollingTime: rockRollingTime},
		owner:  owner,
	}
	rock.Actor.Body = actor.Body{
		Weight: rockWeight,
		Vx:     vx, Vy: -vy,
		MaxX:      rockMaxVel,
		FilterOut: []*actor.Actor{owner},
	}

	return rock
}

func (r *Rock) Init() {
	r.Hitbox.HurtFunc, r.Hitbox.BlockFunc = r.RockHurt, r.RockHurt
	r.Hitbox.PushHitbox(&r.Actor, bump.Rect{X: r.X, Y: r.X, W: rockSize, H: rockSize}, false)
}

func (r *Rock) Update(dt float64) {
	r.Actor.Update(dt)
	_, contacted := r.Hitbox.Hit(&r.Actor, bump.Rect{H: rockSize, W: rockSize}, rockDamage, []*actor.Actor{r.owner})
	if len(contacted) > 1 || r.Body.Ground {
		r.Remove()
	}
}

func (r *Rock) RockHurt(other *actor.Actor, _ *bump.Collision, _ float64) {
	if other != r.owner {
		r.Remove()
	}
}

func calculateVx(x, y, tx, ty, vy float64) float64 {
	widthBuffer := 10.0
	dx := math.Abs(x - tx - widthBuffer)
	dy := math.Max(ty-y, 0)
	a := actor.BGravity * (rockWeight + 1)
	t := vy / a
	t += dy / vy
	vx := math.Max(dx/(2*t), rockMinVel)
	if x > tx {
		vx *= -1
	}

	return vx
}
