package entity

import (
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/render"
	"game/core"
	"game/entity/actor"
	"game/libs/bump"
	"game/vars"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	rockSize               = 5
	rockDamage             = 5
	rockWeight             = 0.6
	rockMinVel, rockMaxVel = 50.0, 100.0
	rockRollingTime        = 200 * time.Millisecond
)

var (
	rockImageFile = "assets/rock.png"
	rockImage     *ebiten.Image
)

type Rock struct {
	*core.BaseEntity
	render      *render.Comp
	body        *body.Comp
	hitbox      *hitbox.Comp
	owner       actor.Actor
	ownerHitbox *hitbox.Comp
}

func init() {
	var err error
	rockImage, _, err = ebitenutil.NewImageFromFile(rockImageFile)
	if err != nil {
		panic(err)
	}
}

func NewRock(x, y float64, owner actor.Actor) *Rock {
	_, _, ownerHitbox, _, ownerAI := owner.Comps()
	vx, vy := rockMaxVel, 60.0
	if target := ownerAI.Target; target != nil {
		tx, ty := target.Position()
		vx = calculateVx(x, y, tx, ty, vy)
	}

	rock := &Rock{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: rockSize, H: rockSize},
		render:     &render.Comp{Image: rockImage, RollingTime: rockRollingTime},
		body: &body.Comp{
			Weight: rockWeight,
			Vx:     vx, Vy: -vy,
			MaxX:      rockMaxVel,
			FilterOut: []core.Entity{owner},
		},
		hitbox:      &hitbox.Comp{},
		owner:       owner,
		ownerHitbox: ownerHitbox,
	}
	rock.Add(rock.render, rock.body, rock.hitbox)

	return rock
}

func (r *Rock) Init() {
	r.body.Friction = false
	r.hitbox.HitFunc = r.RockHurt
	r.hitbox.PushHitbox(bump.Rect{X: r.X, Y: r.X, W: rockSize, H: rockSize}, hitbox.Hit, nil)
}

func (r *Rock) Update(_ float64) {
	_, contacted := r.hitbox.HitFromHitBox(bump.Rect{H: rockSize, W: rockSize}, rockDamage, []*hitbox.Comp{r.ownerHitbox})
	if len(contacted) > 1 || r.body.Ground {
		vars.World.Remove(r)
	}
}

func (r *Rock) RockHurt(other core.Entity, _ *bump.Collision, _ float64, _ hitbox.ContactType) {
	if other != r.owner {
		vars.World.Remove(r)
	}
}

func calculateVx(x, y, tx, ty, vy float64) float64 {
	widthBuffer := 10.0
	dx := math.Abs(x - tx - widthBuffer)
	dy := math.Max(ty-y, 0)
	a := vars.Gravity * rockWeight
	t := vy / a
	t += dy / vy
	vx := math.Max(dx/(2*t), rockMinVel)
	if x > tx {
		vx *= -1
	}

	return vx
}
