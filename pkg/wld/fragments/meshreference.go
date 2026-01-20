package fragments

import (
	"fmt"
)

// MeshReference (0x2D)
// Internal name: None
// Contains a reference to either a Mesh or LegacyMesh fragment.
// This fragment is referenced from a Skeleton fragment.
type MeshReference struct {
	BaseFragment
	// Mesh is the mesh fragment reference
	Mesh Fragment
	// LegacyMesh is the legacy mesh fragment reference
	LegacyMesh *LegacyMesh
}

// FragmentType returns the fragment type ID.
func (f *MeshReference) FragmentType() uint32 {
	return 0x2D
}

// Initialize parses the fragment data.
func (f *MeshReference) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)
	reader := NewFragmentReader(data)

	nameRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	reference, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read reference: %w", err)
	}
	refIdx := int(reference) - 1

	if refIdx >= 0 && refIdx < len(fragments) {
		// Try LegacyMesh first since we have a concrete type for it
		if legacyMesh, ok := fragments[refIdx].(*LegacyMesh); ok {
			f.LegacyMesh = legacyMesh
			return nil
		}

		// Otherwise, store the generic mesh reference
		f.Mesh = fragments[refIdx]
	}

	return nil
}
