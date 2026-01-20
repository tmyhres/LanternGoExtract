package fragments

import (
	"fmt"
	"strings"
)

// BitmapName (0x03)
// Internal Name: None
// This fragment contains the name of a bitmap image.
// It supports more than one bitmap but this is never used.
// Fragment end is padded to end on a DWORD boundary.
type BitmapName struct {
	BaseFragment

	// Filename is the filename of the referenced bitmap.
	Filename string
}

// FragmentType returns the fragment type ID.
func (f *BitmapName) FragmentType() uint32 {
	return 0x03
}

// Initialize parses the fragment data.
func (f *BitmapName) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	// The client supports more than one bitmap reference but is never used
	bitmapCount, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read bitmap count: %w", err)
	}

	if bitmapCount > 1 {
		// Log warning: Bitmap count exceeds 1
	}

	nameLength, err := r.ReadInt16()
	if err != nil {
		return fmt.Errorf("failed to read name length: %w", err)
	}

	// Read and decode the bitmap name
	nameBytes, err := r.ReadBytes(int(nameLength))
	if err != nil {
		return fmt.Errorf("failed to read bitmap name bytes: %w", err)
	}

	// Decode the bitmap name and trim the null character (c style strings)
	f.Filename = DecodeString(nameBytes)
	if len(f.Filename) > 0 {
		f.Filename = strings.ToLower(f.Filename[:len(f.Filename)-1])
	}

	return nil
}

// GetExportFilename returns the filename with a .png extension for export.
func (f *BitmapName) GetExportFilename() string {
	return f.GetFilenameWithoutExtension() + ".png"
}

// GetFilenameWithoutExtension returns the filename without its extension.
func (f *BitmapName) GetFilenameWithoutExtension() string {
	if len(f.Filename) <= 4 {
		return f.Filename
	}
	return f.Filename[:len(f.Filename)-4]
}
