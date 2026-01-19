package fragments

import (
	"fmt"
)

// Fragment06 (0x06)
// Internal Name: None
// Only found in gequip files. Seems to represent 2D sprites in the world (coins).
type Fragment06 struct {
	BaseFragment
}

// FragmentType returns the fragment type ID.
func (f *Fragment06) FragmentType() uint32 {
	return 0x06
}

// Initialize parses the fragment data.
func (f *Fragment06) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	// Read unknown value (unused)
	_, err = r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read value: %w", err)
	}

	return nil
}
