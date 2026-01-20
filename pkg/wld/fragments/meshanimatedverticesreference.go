package fragments

import (
	"fmt"
)

// MeshAnimatedVerticesReference (0x2F)
// Internal name: None
// References a LegacyMeshAnimatedVertices or MeshAnimatedVertices fragment.
// This fragment is referenced from the Mesh fragment, if it's animated.
type MeshAnimatedVerticesReference struct {
	BaseFragment
	// LegacyMeshAnimatedVertices is the legacy mesh animated vertices reference
	LegacyMeshAnimatedVertices *LegacyMeshAnimatedVertices
	// MeshAnimatedVertices is the mesh animated vertices reference (Fragment 0x37)
	MeshAnimatedVertices Fragment
}

// FragmentType returns the fragment type ID.
func (f *MeshAnimatedVerticesReference) FragmentType() uint32 {
	return 0x2F
}

// Initialize parses the fragment data.
func (f *MeshAnimatedVerticesReference) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)
	reader := NewFragmentReader(data)

	nameRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	fragmentID, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read fragment id: %w", err)
	}
	fragIdx := int(fragmentID) - 1

	if fragIdx >= 0 && fragIdx < len(fragments) {
		// Try LegacyMeshAnimatedVertices first
		if legacyAnim, ok := fragments[fragIdx].(*LegacyMeshAnimatedVertices); ok {
			f.LegacyMeshAnimatedVertices = legacyAnim
		} else {
			// Otherwise store as generic reference (for MeshAnimatedVertices 0x37)
			f.MeshAnimatedVertices = fragments[fragIdx]
		}
	}

	_, err = reader.ReadInt32() // flags
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	return nil
}

// GetAnimatedVertices returns the IAnimatedVertices interface.
func (f *MeshAnimatedVerticesReference) GetAnimatedVertices() IAnimatedVertices {
	if f.LegacyMeshAnimatedVertices != nil {
		return f.LegacyMeshAnimatedVertices
	}
	// Check if MeshAnimatedVertices implements IAnimatedVertices
	if anim, ok := f.MeshAnimatedVertices.(IAnimatedVertices); ok {
		return anim
	}
	return nil
}
