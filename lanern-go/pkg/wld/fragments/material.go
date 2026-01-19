package fragments

import (
	"github.com/lanterneq/lanern-go/pkg/wld/datatypes"
)

// Material (0x30) contains information about a material's shader and textures.
// Internal name: _MDF
type Material struct {
	BaseFragment

	// BitmapInfoReference is the reference to the bitmap info used by this material.
	BitmapInfoReference Fragment

	// ShaderType is the shader type used when rendering this material.
	ShaderType ShaderType

	// Brightness is the material brightness value.
	Brightness float32

	// ScaledAmbient is the scaled ambient value.
	ScaledAmbient float32

	// IsHandled indicates if a material has been processed.
	// Used for alternate character skins.
	IsHandled bool
}

// FragmentType returns the fragment type ID (0x30).
func (m *Material) FragmentType() uint32 {
	return 0x30
}

// Initialize parses the material fragment data.
func (m *Material) Initialize(index int, id int, size int, data []byte, fragments []Fragment,
	stringHash map[int]string, isNewFormat bool) error {
	m.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return err
	}
	m.Name = GetStringFromHash(stringHash, nameRef)

	// Read flags
	_, err = r.ReadInt32()
	if err != nil {
		return err
	}

	// Read parameters
	parameters, err := r.ReadInt32()
	if err != nil {
		return err
	}

	// Read RGBA color (pen color - usage unknown)
	_, err = r.ReadByte() // colorR
	if err != nil {
		return err
	}
	_, err = r.ReadByte() // colorG
	if err != nil {
		return err
	}
	_, err = r.ReadByte() // colorB
	if err != nil {
		return err
	}
	_, err = r.ReadByte() // colorA
	if err != nil {
		return err
	}

	// Read brightness
	m.Brightness, err = r.ReadFloat32()
	if err != nil {
		return err
	}

	// Read scaled ambient
	m.ScaledAmbient, err = r.ReadFloat32()
	if err != nil {
		return err
	}

	// Read fragment reference
	fragmentReference, err := r.ReadInt32()
	if err != nil {
		return err
	}

	if fragmentReference != 0 && int(fragmentReference)-1 < len(fragments) {
		m.BitmapInfoReference = fragments[fragmentReference-1]
	}

	// Determine shader type from material type
	// Thanks to PixelBound for figuring this out
	// Use int32 math to avoid overflow: ^int32(0x7FFFFFFF) == int32(0x80000000)
	materialType := datatypes.MaterialType(parameters & 0x7FFFFFFF)

	switch materialType {
	case datatypes.MaterialTypeBoundary:
		m.ShaderType = ShaderTypeBoundary

	case datatypes.MaterialTypeInvisibleUnknown,
		datatypes.MaterialTypeInvisibleUnknown2,
		datatypes.MaterialTypeInvisibleUnknown3:
		m.ShaderType = ShaderTypeInvisible

	case datatypes.MaterialTypeDiffuse,
		datatypes.MaterialTypeDiffuse2,
		datatypes.MaterialTypeDiffuse3,
		datatypes.MaterialTypeDiffuse4,
		datatypes.MaterialTypeDiffuse6,
		datatypes.MaterialTypeDiffuse7,
		datatypes.MaterialTypeDiffuse8,
		datatypes.MaterialTypeCompleteUnknown,
		datatypes.MaterialTypeTransparentMaskedPassable:
		m.ShaderType = ShaderTypeDiffuse

	case datatypes.MaterialTypeTransparent25:
		m.ShaderType = ShaderTypeTransparent25

	case datatypes.MaterialTypeTransparent50:
		m.ShaderType = ShaderTypeTransparent50

	case datatypes.MaterialTypeTransparent75:
		m.ShaderType = ShaderTypeTransparent75

	case datatypes.MaterialTypeTransparentAdditive:
		m.ShaderType = ShaderTypeTransparentAdditive

	case datatypes.MaterialTypeTransparentAdditiveUnlit:
		m.ShaderType = ShaderTypeTransparentAdditiveUnlit

	case datatypes.MaterialTypeTransparentMasked,
		datatypes.MaterialTypeDiffuse5:
		m.ShaderType = ShaderTypeTransparentMasked

	case datatypes.MaterialTypeDiffuseSkydome:
		m.ShaderType = ShaderTypeDiffuseSkydome

	case datatypes.MaterialTypeTransparentSkydome:
		m.ShaderType = ShaderTypeTransparentSkydome

	case datatypes.MaterialTypeTransparentAdditiveUnlitSkydome:
		m.ShaderType = ShaderTypeTransparentAdditiveUnlitSkydome

	default:
		if m.BitmapInfoReference == nil {
			m.ShaderType = ShaderTypeInvisible
		} else {
			m.ShaderType = ShaderTypeDiffuse
		}
	}

	return nil
}

// GetAllBitmapNames returns all bitmap names referenced by this material.
// If includeExtension is false, the .bmp extension is removed.
func (m *Material) GetAllBitmapNames(includeExtension bool) []string {
	var bitmapNames []string

	if m.BitmapInfoReference == nil {
		return bitmapNames
	}

	// This requires access to the BitmapInfoReference's BitmapInfo and its BitmapNames.
	// The actual implementation depends on how these types are defined.
	// For now, return empty slice - caller should cast BitmapInfoReference appropriately.

	return bitmapNames
}
