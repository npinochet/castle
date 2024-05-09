package entity

import (
	"fmt"
	"game/assets"
	"game/comps/render"
	"game/comps/textbox"
	"game/core"
	"game/ext"
	"game/libs/bump"
	"game/utils"
	"game/vars"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const graveW, graveH = 24, 19

var graveImage, _, _ = ebitenutil.NewImageFromFileSystem(assets.FS, "grave.png")

type Grave struct {
	*core.BaseEntity
	render  *render.Comp
	textbox *textbox.Comp
}

func NewGrave(x, y, _, _ float64, props *core.Properties) *Grave {
	y += tileSize - graveH
	entity := &core.BaseEntity{X: x, Y: y, W: graveW, H: graveH}
	text := props.Custom["text"]
	if text == "" {
		text = "Here lies a hero thas saved the world from the darkness that consumed it. Rest in peace.\n" + fmt.Sprintf("Press %s to save", vars.Pad[utils.KeyUp].String())
	}
	grave := &Grave{
		BaseEntity: entity,
		render:     &render.Comp{Image: graveImage},
		textbox: &textbox.Comp{
			Text:      text,
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
	if active && vars.Pad.KeyDown(utils.KeyUp) {
		// TODO: Transition, save and reset game
		vars.Game.Save()
		vars.Game.Reset()
	}
}
