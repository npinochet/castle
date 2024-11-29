package entity

import (
	"game/assets"
	"game/comps/hitbox"
	"game/comps/render"
	"game/core"
	"game/ext"
	"game/libs/bump"
	"slices"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	spikeMargin = 1
	spikeDamage = 20
	spikeTimer  = 2
)

var spikeImage, _, _ = ebitenutil.NewImageFromFileSystem(assets.FS, "spike.png")

type Spike struct {
	*core.BaseEntity
	render       *render.Comp
	hitbox       *hitbox.Comp
	contacted    []*hitbox.Comp
	contactTimer map[*hitbox.Comp]float64
}

func NewSpike(x, y, _, _ float64, props *core.Properties) *Spike {
	spike := &Spike{
		BaseEntity: &core.BaseEntity{X: x + spikeMargin, Y: y, W: tileSize - spikeMargin, H: tileSize},
		render:     &render.Comp{Image: spikeImage, FlipX: props.FlipX, Layer: 1},
		hitbox:     &hitbox.Comp{},
	}
	spike.Add(spike.render, spike.hitbox)

	return spike
}

func (s *Spike) Init() {
	s.contactTimer = map[*hitbox.Comp]float64{}
}

func (s *Spike) Update(dt float64) {
	for c, t := range s.contactTimer {
		if s.contactTimer[c] = t - dt; t <= 0 {
			delete(s.contactTimer, c)
			i := slices.Index(s.contacted, c)
			l := len(s.contacted) - 1
			s.contacted[i], s.contacted = s.contacted[l], s.contacted[:l]
		}
	}

	area := bump.NewRect(s.Rect())
	if len(ext.QueryItems[core.Entity](nil, area, "body")) == 0 {
		return
	}

	area.X, area.Y = 0, 0
	_, contacted := s.hitbox.HitFromHitBox(area, spikeDamage, s.contacted)
	for _, c := range contacted {
		if _, ok := s.contactTimer[c]; ok {
			continue
		}
		s.contactTimer[c] = spikeTimer
		s.contacted = append(s.contacted, c)
	}
}
