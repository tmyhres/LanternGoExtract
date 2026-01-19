package fragments

import (
	"fmt"
)

// Camera (0x08)
// Internal Name: CAMERA_DUMMY
// Unknown fragment purpose. Contains 26 parameters.
type Camera struct {
	BaseFragment

	// Params stores the 26 parameter values.
	// Their purpose is currently unknown.
	Params [26]interface{}
}

// FragmentType returns the fragment type ID.
func (f *Camera) FragmentType() uint32 {
	return 0x08
}

// Initialize parses the fragment data.
func (f *Camera) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	// 26 fields - unknown what they reference
	// Read integers for params 0-4
	for i := 0; i < 5; i++ {
		v, err := r.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read param %d: %w", i, err)
		}
		f.Params[i] = v
	}

	// params5, params6 - floats
	for i := 5; i < 7; i++ {
		v, err := r.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read param %d: %w", i, err)
		}
		f.Params[i] = v
	}

	// params7 - int
	v7, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read param 7: %w", err)
	}
	f.Params[7] = v7

	// params8, params9 - floats
	for i := 8; i < 10; i++ {
		v, err := r.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read param %d: %w", i, err)
		}
		f.Params[i] = v
	}

	// params10 - int
	v10, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read param 10: %w", err)
	}
	f.Params[10] = v10

	// params11, params12 - floats
	for i := 11; i < 13; i++ {
		v, err := r.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read param %d: %w", i, err)
		}
		f.Params[i] = v
	}

	// params13 - int
	v13, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read param 13: %w", err)
	}
	f.Params[13] = v13

	// params14, params15 - floats
	for i := 14; i < 16; i++ {
		v, err := r.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read param %d: %w", i, err)
		}
		f.Params[i] = v
	}

	// params16-25 - all ints
	for i := 16; i < 26; i++ {
		v, err := r.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read param %d: %w", i, err)
		}
		f.Params[i] = v
	}

	return nil
}
