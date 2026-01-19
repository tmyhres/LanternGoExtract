package datatypes

// ActorType represents the type of actor in the world.
type ActorType int

const (
	ActorTypeCamera   ActorType = 0
	ActorTypeStatic   ActorType = 1
	ActorTypeSkeletal ActorType = 2
	ActorTypeParticle ActorType = 3
	ActorTypeSprite   ActorType = 4
)
