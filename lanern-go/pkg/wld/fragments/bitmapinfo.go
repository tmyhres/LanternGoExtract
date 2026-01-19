package fragments

import (
	"fmt"
)

// BitmapInfo (0x04)
// Internal name: _SPRITE
// This fragment contains a reference to a 0x03 fragment and information about animation.
type BitmapInfo struct {
	BaseFragment

	// IsAnimated indicates whether the texture is animated.
	IsAnimated bool

	// BitmapNames contains the bitmap names referenced.
	BitmapNames []*BitmapName

	// AnimationDelayMs is the number of milliseconds before the next texture is swapped.
	AnimationDelayMs int32
}

// FragmentType returns the fragment type ID.
func (f *BitmapInfo) FragmentType() uint32 {
	return 0x04
}

// Initialize parses the fragment data.
func (f *BitmapInfo) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	// Read flags
	flags, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	f.IsAnimated = IsBitSet(flags, 3)

	// Read bitmap count
	bitmapCount, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read bitmap count: %w", err)
	}

	f.BitmapNames = make([]*BitmapName, 0, bitmapCount)

	if f.IsAnimated {
		f.AnimationDelayMs, err = r.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read animation delay: %w", err)
		}
	}

	// Read bitmap references
	for i := int32(0); i < bitmapCount; i++ {
		refIndex, err := r.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read bitmap reference %d: %w", i, err)
		}

		fragIndex := int(refIndex) - 1
		if fragIndex >= 0 && fragIndex < len(fragments) {
			if bitmapName, ok := fragments[fragIndex].(*BitmapName); ok {
				f.BitmapNames = append(f.BitmapNames, bitmapName)
			}
		}
	}

	return nil
}
