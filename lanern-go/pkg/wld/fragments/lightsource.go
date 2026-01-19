package fragments

import (
	"fmt"

	"github.com/lanterneq/lanern-go/pkg/wld/datatypes"
)

// LightSource (0x1B)
// Internal name: _LIGHTDEF/_LDEF
// Defines color information about a light.
type LightSource struct {
	BaseFragment
	// IsPlacedLightSource indicates if this is a placed light source when used in light.wld
	IsPlacedLightSource bool
	// IsColoredLight indicates if this is a colored light (fragment size is larger)
	IsColoredLight bool
	// Color is the color of the light if it is colored
	Color datatypes.Vec4
	// Attenuation is a guess from Windcatcher, not sure what it is
	Attenuation int
	// SomeValue is an unknown value
	SomeValue float32
}

// FragmentType returns the fragment type ID.
func (f *LightSource) FragmentType() uint32 {
	return 0x1B
}

// Initialize parses the fragment data.
func (f *LightSource) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)
	reader := NewFragmentReader(data)

	nameRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	flags, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	if (flags & (1 << 1)) != 0 {
		f.IsPlacedLightSource = true
	}

	if (flags & (1 << 4)) != 0 {
		f.IsColoredLight = true
	}

	if !f.IsPlacedLightSource {
		if !f.IsColoredLight {
			_, err := reader.ReadInt32() // something1
			if err != nil {
				return fmt.Errorf("failed to read something1: %w", err)
			}
			f.SomeValue, err = reader.ReadFloat32()
			if err != nil {
				return fmt.Errorf("failed to read some value: %w", err)
			}
			return nil
		}

		attenuation, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read attenuation: %w", err)
		}
		f.Attenuation = int(attenuation)

		alpha, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read alpha: %w", err)
		}
		red, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read red: %w", err)
		}
		green, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read green: %w", err)
		}
		blue, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read blue: %w", err)
		}
		f.Color = datatypes.Vec4{X: red, Y: green, Z: blue, W: alpha}

		return nil
	}

	if !f.IsColoredLight {
		_, err := reader.ReadInt32() // something1
		if err != nil {
			return fmt.Errorf("failed to read something1: %w", err)
		}
		_, err = reader.ReadFloat32() // something2
		if err != nil {
			return fmt.Errorf("failed to read something2: %w", err)
		}
		return nil
	}

	// Not sure yet what the purpose of this fragment is in the main zone file
	// For now, return
	if !f.IsPlacedLightSource && f.Name == "DEFAULT_LIGHTDEF" {
		_, err := reader.ReadInt32() // unknown
		if err != nil {
			return fmt.Errorf("failed to read unknown: %w", err)
		}
		_, err = reader.ReadFloat32() // unknown6
		if err != nil {
			return fmt.Errorf("failed to read unknown6: %w", err)
		}
		return nil
	}

	_, err = reader.ReadInt32() // unknown1
	if err != nil {
		return fmt.Errorf("failed to read unknown1: %w", err)
	}

	if !f.IsColoredLight {
		_, err := reader.ReadInt32() // unknown
		if err != nil {
			return fmt.Errorf("failed to read unknown: %w", err)
		}
		f.Color = datatypes.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 1.0}
		_, err = reader.ReadInt32() // unknown2
		if err != nil {
			return fmt.Errorf("failed to read unknown2: %w", err)
		}
		_, err = reader.ReadInt32() // unknown3
		if err != nil {
			return fmt.Errorf("failed to read unknown3: %w", err)
		}
	} else {
		attenuation, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read attenuation: %w", err)
		}
		f.Attenuation = int(attenuation)

		alpha, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read alpha: %w", err)
		}
		red, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read red: %w", err)
		}
		green, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read green: %w", err)
		}
		blue, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read blue: %w", err)
		}
		f.Color = datatypes.Vec4{X: red, Y: green, Z: blue, W: alpha}
	}

	return nil
}
