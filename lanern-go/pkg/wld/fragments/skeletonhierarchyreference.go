package fragments

import (
	"fmt"
)

// SkeletonHierarchyReference (0x11)
// Internal name: None
// A reference to a skeleton hierarchy fragment (0x10).
type SkeletonHierarchyReference struct {
	BaseFragment

	// SkeletonHierarchy is the referenced skeleton hierarchy.
	SkeletonHierarchy *SkeletonHierarchy
}

// FragmentType returns the fragment type ID.
func (f *SkeletonHierarchyReference) FragmentType() uint32 {
	return 0x11
}

// Initialize parses the fragment data.
func (f *SkeletonHierarchyReference) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)

	r := NewFragmentReader(data)

	// Reference is usually 0
	nameRef, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	reference, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read reference: %w", err)
	}

	fragIndex := int(reference) - 1
	if fragIndex >= 0 && fragIndex < len(fragments) {
		if skeleton, ok := fragments[fragIndex].(*SkeletonHierarchy); ok {
			f.SkeletonHierarchy = skeleton
		}
	}

	if f.SkeletonHierarchy == nil {
		// Log error: Bad skeleton hierarchy reference
	}

	// Params are 0 - confirmed
	_, err = r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read params1: %w", err)
	}

	return nil
}
