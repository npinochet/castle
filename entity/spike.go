package entity

import (
	"game/assets"
	"game/comps/hitbox"
	"game/comps/render"
	"game/core"
	"game/ext"
	"game/libs/bump"
	"slices"
	"time"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	spikeDamage = 20
	spikeTimer  = 2 * time.Second
)

var (
	spikeImage, _, _  = ebitenutil.NewImageFromFileSystem(assets.FS, "spike.png")
	spikeContactTimer = map[*hitbox.Comp]time.Time{}
	spikeContacted    []*hitbox.Comp
)

type Spike struct {
	*core.BaseEntity
	render *render.Comp
	hitbox *hitbox.Comp
}

func NewSpike(x, y, _, _ float64, props *core.Properties) *Spike {
	spike := &Spike{
		BaseEntity: &core.BaseEntity{X: x + 1, Y: y, W: tileSize - 3, H: tileSize},
		render:     &render.Comp{X: -2, Image: spikeImage, FlipX: props.FlipX, FlipY: props.FlipY, Layer: 1},
		hitbox:     &hitbox.Comp{},
	}
	spike.Add(spike.render, spike.hitbox)

	return spike
}

func (s *Spike) Init() {}

func (s *Spike) Update(dt float64) {
	for c, t := range spikeContactTimer {
		if elapsed := time.Since(t); elapsed > spikeTimer {
			delete(spikeContactTimer, c)
			i := slices.Index(spikeContacted, c)
			l := len(spikeContacted) - 1
			spikeContacted[i], spikeContacted = spikeContacted[l], spikeContacted[:l]
		}
	}

	area := bump.NewRect(s.Rect())
	if len(ext.QueryItems[core.Entity](nil, area, "body")) == 0 {
		return
	}

	area.X, area.Y = 0, 0
	_, contacted := s.hitbox.HitFromHitBox(area, spikeDamage, spikeContacted)
	for _, c := range contacted {
		if _, ok := spikeContactTimer[c]; ok {
			continue
		}
		spikeContactTimer[c] = time.Now()
		spikeContacted = append(spikeContacted, c)
	}
}
