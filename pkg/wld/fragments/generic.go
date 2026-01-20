package fragments

// Generic is a fallback fragment type for unknown/unhandled fragment IDs.
// It stores the raw fragment data for debugging purposes.
type Generic struct {
	BaseFragment
	FragmentID int
	Data       []byte
}

// FragmentType returns the fragment type ID.
func (f *Generic) FragmentType() uint32 {
	return uint32(f.FragmentID)
}

// Initialize parses the fragment data.
func (f *Generic) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)
	f.FragmentID = id
	f.Data = data
	return nil
}
