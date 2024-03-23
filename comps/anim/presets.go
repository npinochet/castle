package anim

import "game/vars"

func DefaultFsm() *Fsm {
	return &Fsm{
		Transitions: map[string]string{
			vars.ParryBlockTag: vars.BlockTag,
			vars.BlockTag:      "",
			vars.ClimbTag:      "",
		},
	}
}
