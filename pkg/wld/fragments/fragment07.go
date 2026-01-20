package fragments

import (
	"fmt"
)

// Fragment07 (0x07)
// Internal Name: None
// Only found in gequip files. References a 0x06 fragment.
// This fragment can be referenced by an actor fragment.
type Fragment07 struct {
	BaseFragment

	// Fragment06Ref is the referenced Fragment06.
	Fragment06Ref *Fragment06
}

// FragmentType returns the fragment type ID.
func (f *Fragment07) FragmentType() uint32 {
	return 0x07
}

// Initialize parses the fragment data.
func (f *Fragment07) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	// Read Fragment06 reference
	refIndex, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read fragment06 reference: %w", err)
	}

	fragIndex := int(refIndex) - 1
	if fragIndex >= 0 && fragIndex < len(fragments) {
		if frag06, ok := fragments[fragIndex].(*Fragment06); ok {
			f.Fragment06Ref = frag06
		}
	}

	// Read unknown value (always 0)
	_, err = r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read value08: %w", err)
	}

	return nil
}
