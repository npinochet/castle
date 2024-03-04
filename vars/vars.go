package vars

var (
	// Collider.
	Gravity                     = 300.0
	DefaultMaxX, DefaultMaxY    = 20.0, 200.0
	GroundFriction, AirFriction = 12.0, 2.0 // TODO: Tune this variables. They might be too high.
	CollisionStiffness          = 1.0
	FrictionEpsilon             = 0.05
)
