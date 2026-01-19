package fragments

// VertexColorsReference (0x33) references a VertexColors fragment.
// Internal name: None
// Referenced by an ObjectInstance fragment.
type VertexColorsReference struct {
	BaseFragment

	// VertexColors is the reference to the vertex colors fragment.
	VertexColors *VertexColors
}

// FragmentType returns the fragment type ID (0x33).
func (v *VertexColorsReference) FragmentType() uint32 {
	return 0x33
}

// Initialize parses the vertex colors reference fragment data.
func (v *VertexColorsReference) Initialize(index int, id int, size int, data []byte, fragments []Fragment,
	stringHash map[int]string, isNewFormat bool) error {
	v.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return err
	}
	v.Name = GetStringFromHash(stringHash, nameRef)

	// Read vertex colors reference
	vertexColorsRef, err := r.ReadInt32()
	if err != nil {
		return err
	}

	fragIdx := int(vertexColorsRef) - 1
	if fragIdx >= 0 && fragIdx < len(fragments) {
		if vc, ok := fragments[fragIdx].(*VertexColors); ok {
			v.VertexColors = vc
		}
	}

	return nil
}
