package fragments

import (
	"fmt"
)

// Fragment16 (0x16)
// Internal Name: None
// An unknown fragment. Found in zone files.
type Fragment16 struct {
	BaseFragment
	Unknown float32
}

// FragmentType returns the fragment type ID.
func (f *Fragment16) FragmentType() uint32 {
	return 0x16
}

// Initialize parses the fragment data.
func (f *Fragment16) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)
	reader := NewFragmentReader(data)

	nameRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	// Should be 0.1
	f.Unknown, err = reader.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read unknown float: %w", err)
	}

	return nil
}
