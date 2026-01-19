package fragments

import (
	"fmt"
)

// BitmapInfoReference (0x05)
// Internal name: None
// Contains a reference to a BitmapInfo fragment.
type BitmapInfoReference struct {
	BaseFragment

	// BitmapInfo is the reference to the BitmapInfo fragment.
	BitmapInfo *BitmapInfo
}

// FragmentType returns the fragment type ID.
func (f *BitmapInfoReference) FragmentType() uint32 {
	return 0x05
}

// Initialize parses the fragment data.
func (f *BitmapInfoReference) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	// Read BitmapInfo reference
	refIndex, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read bitmap info reference: %w", err)
	}

	fragIndex := int(refIndex) - 1
	if fragIndex >= 0 && fragIndex < len(fragments) {
		if bitmapInfo, ok := fragments[fragIndex].(*BitmapInfo); ok {
			f.BitmapInfo = bitmapInfo
		}
	}

	// Either 0 or 80 - unknown purpose
	_, err = r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	return nil
}
