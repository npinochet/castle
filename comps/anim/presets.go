package anim

const (
	IdleTag       = "Idle"
	WalkTag       = "Walk"
	AttackTag     = "Attack"
	BlockTag      = "Block"
	ParryBlockTag = "ParryBlock"
	StaggerTag    = "Stagger"
	ClimbTag      = "Climb"
	ConsumeTag    = "Consume"
)

func DefaultFsm() *Fsm {
	return &Fsm{
		Transitions: map[string]string{
			WalkTag:       IdleTag,
			AttackTag:     IdleTag,
			StaggerTag:    IdleTag,
			ParryBlockTag: BlockTag,
			BlockTag:      "",
			ClimbTag:      "",
		},
		Exit: map[string]func(*Comp){
			StaggerTag: func(ac *Comp) { ac.Data.PlaySpeed = 1 },
		},
	}
}
