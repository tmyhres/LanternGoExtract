package datatypes

// MeshReference is a forward declaration interface for mesh reference fragments.
// The actual implementation should be in the fragments package.
type MeshReference interface{}

// ParticleCloud is a forward declaration interface for particle cloud fragments.
// The actual implementation should be in the fragments package.
type ParticleCloud interface{}

// SkeletonBone represents a node in the skeleton tree.
type SkeletonBone struct {
	Index           int
	Name            string
	FullPath        string
	CleanedName     string
	CleanedFullPath string
	Children        []int
	Track           TrackFragment
	MeshReference   MeshReference
	ParticleCloud   ParticleCloud
	AnimationTracks map[string]TrackFragment
	Parent          *SkeletonBone
}
