package fragments

import (
	"fmt"
)

// LightSourceReference (0x1C)
// Internal name: None
// References a LightSource fragment.
type LightSourceReference struct {
	BaseFragment
	// LightSource is the light source (0x1B) fragment reference
	LightSource *LightSource
}

// FragmentType returns the fragment type ID.
func (f *LightSourceReference) FragmentType() uint32 {
	return 0x1C
}

// Initialize parses the fragment data.
func (f *LightSourceReference) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)
	reader := NewFragmentReader(data)

	nameRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	lightRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read light reference: %w", err)
	}
	if lightRef > 0 && int(lightRef-1) < len(fragments) {
		if light, ok := fragments[lightRef-1].(*LightSource); ok {
			f.LightSource = light
		}
	}

	return nil
}
