package anim

import "game/vars"

func DefaultFsm() *Fsm {
	return &Fsm{
		Transitions: map[string]string{
			vars.WalkTag:       vars.IdleTag,
			vars.AttackTag:     vars.IdleTag,
			vars.StaggerTag:    vars.IdleTag,
			vars.ParryBlockTag: vars.BlockTag,
			vars.BlockTag:      "",
			vars.ClimbTag:      "",
		},
		Exit: map[string]func(*Comp){
			vars.StaggerTag: func(ac *Comp) { ac.Data.PlaySpeed = 1 },
		},
	}
}
