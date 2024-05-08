package entity

import (
	"game/assets"
	"game/comps/render"
	"game/comps/textbox"
	"game/core"
	"game/ext"
	"game/libs/bump"
	"game/vars"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const graveW, graveH = 24, 19

var graveImage, _, _ = ebitenutil.NewImageFromFileSystem(assets.FS, "grave.png")

type Grave struct {
	*core.BaseEntity
	render  *render.Comp
	active  bool
	textbox *textbox.Comp
}

func NewGrave(x, y, _, _ float64, _ *core.Properties) *Grave {
	y += tileSize - graveH
	entity := &core.BaseEntity{X: x, Y: y, W: graveW, H: graveH}
	grave := &Grave{
		BaseEntity: entity,
		render:     &render.Comp{Image: graveImage},
		textbox: &textbox.Comp{
			Text:      "Press [E] to interact",
			Area:      func() bump.Rect { return bump.NewRect(entity.Rect()) },
			Indicator: true,
		},
	}
	grave.Add(grave.render, grave.textbox)

	return grave
}
func (g *Grave) Priority() int { return -1 }

func (g *Grave) Init() {}

func (g *Grave) Update(_ float64) {
	active := false
	for _, e := range ext.QueryItems[core.Entity](g, bump.NewRect(g.Rect()), "body") {
		if core.GetFlag(e, vars.PlayerTeamFlag) {
			active = true

			break
		}
	}
	g.active = active
}
