package actor

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

func DefaultAFsm() *AFsm {
	return &AFsm{
		Initial: IdleTag,
		Transitions: map[string]string{
			WalkTag:       IdleTag,
			AttackTag:     IdleTag,
			StaggerTag:    IdleTag,
			ParryBlockTag: BlockTag,
			BlockTag:      "",
			ClimbTag:      "",
		},
		Exit: map[string]func(*Actor){
			StaggerTag: func(a *Actor) { a.Anim.Data.PlaySpeed = 1 },
		},
	}
}
