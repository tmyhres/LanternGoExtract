package fragments

import (
	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/datatypes"
)

// GlobalAmbientLight (0x35) contains the color value for ambient lighting.
// Internal name: None
// This fragment contains no name reference and is only found in zone WLDs (e.g. akanon.wld).
// Used to boost the darkness in some zones.
type GlobalAmbientLight struct {
	BaseFragment

	// Color is the ambient light color.
	Color datatypes.Color
}

// FragmentType returns the fragment type ID (0x35).
func (g *GlobalAmbientLight) FragmentType() uint32 {
	return 0x35
}

// Initialize parses the global ambient light fragment data.
func (g *GlobalAmbientLight) Initialize(index int, id int, size int, data []byte, fragments []Fragment,
	stringHash map[int]string, isNewFormat bool) error {
	g.initBase(index, size)

	r := NewFragmentReader(data)

	// Color is in BGRA format. A is always 255.
	colorValue, err := r.ReadInt32()
	if err != nil {
		return err
	}

	// Extract BGRA components and convert to RGBA
	b := int(colorValue & 0xFF)
	green := int((colorValue >> 8) & 0xFF)
	red := int((colorValue >> 16) & 0xFF)
	a := int((colorValue >> 24) & 0xFF)

	g.Color = datatypes.NewColor(red, green, b, a)

	return nil
}
