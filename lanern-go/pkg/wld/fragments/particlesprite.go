package fragments

import (
	"fmt"
)

// ParticleSprite (0x26)
// Internal name: _SPB
// Assumed to be a particle sprite fragment.
type ParticleSprite struct {
	BaseFragment
	// BitmapReference is a reference to the bitmap info
	BitmapReference Fragment
}

// FragmentType returns the fragment type ID.
func (f *ParticleSprite) FragmentType() uint32 {
	return 0x26
}

// Initialize parses the fragment data.
func (f *ParticleSprite) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)
	reader := NewFragmentReader(data)

	nameRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	_, err = reader.ReadInt32() // flags, always 0
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	fragmentRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read fragment reference: %w", err)
	}

	_, err = reader.ReadInt32() // value12, always the same value
	if err != nil {
		return fmt.Errorf("failed to read value12: %w", err)
	}

	if fragmentRef > 0 && int(fragmentRef-1) < len(fragments) {
		f.BitmapReference = fragments[fragmentRef-1]
	}

	return nil
}
