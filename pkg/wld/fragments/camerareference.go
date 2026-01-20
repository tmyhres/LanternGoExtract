package fragments

import (
	"fmt"
)

// CameraReference (0x09)
// Internal Name: None
// References a Camera fragment.
type CameraReference struct {
	BaseFragment

	// Camera is the referenced Camera fragment.
	Camera *Camera
}

// FragmentType returns the fragment type ID.
func (f *CameraReference) FragmentType() uint32 {
	return 0x09
}

// Initialize parses the fragment data.
func (f *CameraReference) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	// Read Camera reference
	refIndex, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read camera reference: %w", err)
	}

	fragIndex := int(refIndex) - 1
	if fragIndex >= 0 && fragIndex < len(fragments) {
		if camera, ok := fragments[fragIndex].(*Camera); ok {
			f.Camera = camera
		}
	}

	// Usually 0 - flags
	_, err = r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	return nil
}
