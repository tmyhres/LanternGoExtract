package fragments

import (
	"fmt"
)

// ParticleSpriteReference (0x27)
// Internal name: None
// Assumed to reference a particle sprite fragment.
type ParticleSpriteReference struct {
	BaseFragment
	// Reference is the particle sprite fragment reference
	Reference *ParticleSprite
}

// FragmentType returns the fragment type ID.
func (f *ParticleSpriteReference) FragmentType() uint32 {
	return 0x27
}

// Initialize parses the fragment data.
func (f *ParticleSpriteReference) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)
	reader := NewFragmentReader(data)

	nameRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	fragmentRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read fragment reference: %w", err)
	}

	_, err = reader.ReadInt32() // value08, always 0
	if err != nil {
		return fmt.Errorf("failed to read value08: %w", err)
	}

	if fragmentRef > 0 && int(fragmentRef-1) < len(fragments) {
		if ps, ok := fragments[fragmentRef-1].(*ParticleSprite); ok {
			f.Reference = ps
		}
	}

	return nil
}
