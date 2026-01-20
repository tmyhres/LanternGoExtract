package wld

import (
	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
)

// FragmentFactory is a function type that creates new fragment instances.
type FragmentFactory func() fragments.Fragment

// fragmentRegistry maps fragment type IDs to their factory functions.
// This allows dynamic creation of the appropriate fragment type when parsing WLD files.
var fragmentRegistry = map[int]FragmentFactory{
	// Materials
	0x03: func() fragments.Fragment { return &fragments.BitmapName{} },
	0x04: func() fragments.Fragment { return &fragments.BitmapInfo{} },
	0x05: func() fragments.Fragment { return &fragments.BitmapInfoReference{} },
	0x30: func() fragments.Fragment { return &fragments.Material{} },
	0x31: func() fragments.Fragment { return &fragments.MaterialList{} },

	// BSP Tree
	0x21: func() fragments.Fragment { return &fragments.BspTree{} },
	0x22: func() fragments.Fragment { return &fragments.BspRegion{} },
	0x29: func() fragments.Fragment { return &fragments.BspRegionType{} },

	// Meshes
	0x36: func() fragments.Fragment { return &fragments.Mesh{} },
	0x37: func() fragments.Fragment { return &fragments.MeshAnimatedVertices{} },
	0x2E: func() fragments.Fragment { return &fragments.LegacyMeshAnimatedVertices{} },
	0x2F: func() fragments.Fragment { return &fragments.MeshAnimatedVerticesReference{} },
	0x2D: func() fragments.Fragment { return &fragments.MeshReference{} },
	0x2C: func() fragments.Fragment { return &fragments.LegacyMesh{} },

	// Animation
	0x10: func() fragments.Fragment { return &fragments.SkeletonHierarchy{} },
	0x11: func() fragments.Fragment { return &fragments.SkeletonHierarchyReference{} },
	0x12: func() fragments.Fragment { return &fragments.TrackDefFragment{} },
	0x13: func() fragments.Fragment { return &fragments.TrackFragment{} },
	0x14: func() fragments.Fragment { return &fragments.Actor{} },

	// Lights
	0x1B: func() fragments.Fragment { return &fragments.LightSource{} },
	0x1C: func() fragments.Fragment { return &fragments.LightSourceReference{} },
	0x28: func() fragments.Fragment { return &fragments.LightInstance{} },
	0x2A: func() fragments.Fragment { return &fragments.AmbientLight{} },
	0x35: func() fragments.Fragment { return &fragments.GlobalAmbientLight{} },

	// Vertex colors
	0x32: func() fragments.Fragment { return &fragments.VertexColors{} },
	0x33: func() fragments.Fragment { return &fragments.VertexColorsReference{} },

	// Particle Cloud
	0x26: func() fragments.Fragment { return &fragments.ParticleSprite{} },
	0x27: func() fragments.Fragment { return &fragments.ParticleSpriteReference{} },
	0x34: func() fragments.Fragment { return &fragments.ParticleCloud{} },

	// General
	0x15: func() fragments.Fragment { return &fragments.ObjectInstance{} },

	// Not used/unknown
	0x08: func() fragments.Fragment { return &fragments.Camera{} },
	0x09: func() fragments.Fragment { return &fragments.CameraReference{} },
	0x16: func() fragments.Fragment { return &fragments.Fragment16{} },
	0x17: func() fragments.Fragment { return &fragments.Polyhedron{} },
	0x18: func() fragments.Fragment { return &fragments.PolyhedronReference{} },
	0x06: func() fragments.Fragment { return &fragments.Fragment06{} },
	0x07: func() fragments.Fragment { return &fragments.Fragment07{} },
}

// GetFragmentFactory returns the factory function for the given fragment type ID.
// Returns nil if the fragment type is not registered.
func GetFragmentFactory(fragmentID int) FragmentFactory {
	if factory, ok := fragmentRegistry[fragmentID]; ok {
		return factory
	}
	return nil
}

// CreateFragment creates a new fragment instance for the given type ID.
// Returns a Generic fragment if the type is not registered.
func CreateFragment(fragmentID int) fragments.Fragment {
	if factory := GetFragmentFactory(fragmentID); factory != nil {
		return factory()
	}
	return &fragments.Generic{}
}

// IsKnownFragmentType returns true if the fragment type ID is registered.
func IsKnownFragmentType(fragmentID int) bool {
	_, ok := fragmentRegistry[fragmentID]
	return ok
}

// RegisterFragment registers a new fragment type with its factory function.
// This can be used to extend the fragment registry at runtime.
func RegisterFragment(fragmentID int, factory FragmentFactory) {
	fragmentRegistry[fragmentID] = factory
}

// GetRegisteredFragmentTypes returns a slice of all registered fragment type IDs.
func GetRegisteredFragmentTypes() []int {
	types := make([]int, 0, len(fragmentRegistry))
	for id := range fragmentRegistry {
		types = append(types, id)
	}
	return types
}
