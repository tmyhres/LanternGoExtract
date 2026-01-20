package fragments

import (
	"fmt"
)

// AmbientLight (0x2A)
// Internal name: _AMBIENTLIGHT
// Defines the ambient light for a group of regions. This fragment is found in the Trilogy client but is UNUSED.
type AmbientLight struct {
	BaseFragment
	// LightReference is a reference to a 0x1C light source reference fragment which defines the light for the regions
	LightReference *LightSourceReference
	// Regions contains the regions that the light reference apply to
	Regions []int
}

// FragmentType returns the fragment type ID.
func (f *AmbientLight) FragmentType() uint32 {
	return 0x2A
}

// Initialize parses the fragment data.
func (f *AmbientLight) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
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
		if ref, ok := fragments[refIdx].(*LightSourceReference); ok {
			f.LightReference = ref
		}
	}

	_, err = reader.ReadInt32() // flags
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	regionCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read region count: %w", err)
	}

	f.Regions = make([]int, regionCount)
	for i := int32(0); i < regionCount; i++ {
		regionID, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read region id: %w", err)
		}
		f.Regions[i] = int(regionID)
	}

	return nil
}
