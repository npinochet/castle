package anim

import "game/vars"

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

type AlphaScalerColor struct{ A uint32 }

func (c AlphaScalerColor) RGBA() (uint32, uint32, uint32, uint32) { return c.A, c.A, c.A, c.A }

var WhiteScalerColor = AlphaScalerColor{0xffff * 4}
