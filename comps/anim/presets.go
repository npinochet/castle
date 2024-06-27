package anim

import (
	"game/vars"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2/colorm"
)

func DefaultFsm() *Fsm {
	return &Fsm{
		Initial: vars.IdleTag,
		Transitions: map[string]string{
			vars.ParryBlockTag: vars.BlockTag,
			vars.BlockTag:      "",
			vars.ClimbTag:      "",
		},
	}
}

type UberColor struct{ R, G, B, A uint32 }

func (c UberColor) RGBA() (uint32, uint32, uint32, uint32) { return c.R, c.G, c.B, c.A }

var (
	WhiteScalerColor     = UberColor{0xffff * 4, 0xffff * 4, 0xffff * 4, 0xffff * 4}
	NormalMaskColor      = color.NRGBA{127, 127, 255, 255}
	FillNormalMaskColorM = colorm.ColorM{}
)

func init() {
	FillNormalMaskColorM.Scale(0, 0, 0, 1)
	r, g, b := float64(NormalMaskColor.R)/math.MaxUint8, float64(NormalMaskColor.G)/math.MaxUint8, float64(NormalMaskColor.B)/math.MaxUint8
	FillNormalMaskColorM.Translate(r, g, b, 0)
}
