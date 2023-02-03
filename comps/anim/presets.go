package anim

const (
	IdleTag    = "Idle"
	WalkTag    = "Walk"
	AttackTag  = "Attack"
	BlockTag   = "Block"
	StaggerTag = "Stagger"
	ClimbTag   = "Climb"
	ConsumeTag = "Consume"
)

var DefaultAnimFsm = &Fsm{
	Transitions: map[string]string{WalkTag: IdleTag, AttackTag: IdleTag, StaggerTag: IdleTag, BlockTag: "", ClimbTag: ""},
	Exit: map[string]func(*Comp){
		StaggerTag: func(ac *Comp) { ac.Data.PlaySpeed = 1 },
	},
}

func DefaultFsm() *Fsm {
	return DefaultAnimFsm
}
