package fragments

import (
	"fmt"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/datatypes"
)

// LightInstance (0x28)
// Internal name: None
// Defines the position and radius of a light.
type LightInstance struct {
	BaseFragment
	// LightReference is the light reference (0x1C) this fragment refers to
	LightReference *LightSourceReference
	// Position is the position of the light
	Position datatypes.Vec3
	// Radius is the radius of the light
	Radius float32
}

// FragmentType returns the fragment type ID.
func (f *LightInstance) FragmentType() uint32 {
	return 0x28
}

// Initialize parses the fragment data.
func (f *LightInstance) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
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
		if ref, ok := fragments[lightRef-1].(*LightSourceReference); ok {
			f.LightReference = ref
		}
	}

	_, err = reader.ReadInt32() // flags
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	x, err := reader.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read position x: %w", err)
	}
	y, err := reader.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read position y: %w", err)
	}
	z, err := reader.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read position z: %w", err)
	}
	f.Position = datatypes.Vec3{X: x, Y: y, Z: z}

	f.Radius, err = reader.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read radius: %w", err)
	}

	return nil
}
