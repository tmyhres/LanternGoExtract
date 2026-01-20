package fragments

import (
	"fmt"
)

// PolyhedronReference (0x18)
// Internal Name: None
// References a Polyhedron fragment.
type PolyhedronReference struct {
	BaseFragment
	Polyhedron *Polyhedron
	Params1    float32
}

// FragmentType returns the fragment type ID.
func (f *PolyhedronReference) FragmentType() uint32 {
	return 0x18
}

// Initialize parses the fragment data.
func (f *PolyhedronReference) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)
	reader := NewFragmentReader(data)

	nameRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	polyRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read polyhedron reference: %w", err)
	}
	if polyRef > 0 && int(polyRef-1) < len(fragments) {
		if poly, ok := fragments[polyRef-1].(*Polyhedron); ok {
			f.Polyhedron = poly
		}
	}

	f.Params1, err = reader.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read params1: %w", err)
	}

	return nil
}
