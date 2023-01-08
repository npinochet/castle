package anim

const (
	IdleTag    = "Idle"
	WalkTag    = "Walk"
	AttackTag  = "Attack"
	BlockTag   = "Block"
	StaggerTag = "Stagger"
)

var DefaultAnimFsm = &Fsm{
	Transitions: map[string]string{WalkTag: IdleTag, AttackTag: IdleTag, StaggerTag: IdleTag, BlockTag: ""},
	Exit: map[string]func(*Comp){
		StaggerTag: func(ac *Comp) { ac.Data.PlaySpeed = 1 },
	},
}

func DefaultFsm() *Fsm {
	return DefaultAnimFsm
}
