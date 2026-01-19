package fragments

import (
	"github.com/lanterneq/lanern-go/pkg/wld/datatypes"
)

// VertexColors (0x32) contains a list of colors, one per vertex.
// Internal name: _DMT
// Represents baked lighting data for an object.
// Vertex color data for zone meshes are baked into the mesh as they are unique.
type VertexColors struct {
	BaseFragment

	// Colors is the list of vertex colors corresponding with each vertex.
	Colors []datatypes.Color
}

// FragmentType returns the fragment type ID (0x32).
func (v *VertexColors) FragmentType() uint32 {
	return 0x32
}

// Initialize parses the vertex colors fragment data.
func (v *VertexColors) Initialize(index int, id int, size int, data []byte, fragments []Fragment,
	stringHash map[int]string, isNewFormat bool) error {
	v.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return err
	}
	v.Name = GetStringFromHash(stringHash, nameRef)

	// Read unknown value
	_, err = r.ReadInt32()
	if err != nil {
		return err
	}

	// Read color count
	colorCount, err := r.ReadInt32()
	if err != nil {
		return err
	}

	// Read unknown2 - typically contains 1
	_, err = r.ReadInt32()
	if err != nil {
		return err
	}

	// Read unknown3 - typically contains 200
	_, err = r.ReadInt32()
	if err != nil {
		return err
	}

	// Read unknown4 - typically contains 0
	_, err = r.ReadInt32()
	if err != nil {
		return err
	}

	v.Colors = make([]datatypes.Color, 0, colorCount)

	for i := int32(0); i < colorCount; i++ {
		// Color is stored as BGRA in the file
		colorValue, err := r.ReadInt32()
		if err != nil {
			return err
		}

		// Extract BGRA components
		b := int(colorValue & 0xFF)
		g := int((colorValue >> 8) & 0xFF)
		red := int((colorValue >> 16) & 0xFF)
		a := int((colorValue >> 24) & 0xFF)

		v.Colors = append(v.Colors, datatypes.NewColor(red, g, b, a))
	}

	return nil
}
