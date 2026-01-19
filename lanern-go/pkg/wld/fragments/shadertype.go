package fragments

// ShaderType represents the shaders used by the EQ client to render surfaces.
type ShaderType int

const (
	// ShaderTypeDiffuse is the standard diffuse shader.
	ShaderTypeDiffuse ShaderType = 0

	// ShaderTypeTransparent25 is 25% transparent.
	ShaderTypeTransparent25 ShaderType = 1

	// ShaderTypeTransparent50 is 50% transparent.
	ShaderTypeTransparent50 ShaderType = 2

	// ShaderTypeTransparent75 is 75% transparent.
	ShaderTypeTransparent75 ShaderType = 3

	// ShaderTypeTransparentAdditive is transparent additive.
	ShaderTypeTransparentAdditive ShaderType = 4

	// ShaderTypeTransparentAdditiveUnlit is transparent additive unlit.
	ShaderTypeTransparentAdditiveUnlit ShaderType = 5

	// ShaderTypeTransparentMasked is transparent masked.
	ShaderTypeTransparentMasked ShaderType = 6

	// ShaderTypeDiffuseSkydome is diffuse for skydome.
	ShaderTypeDiffuseSkydome ShaderType = 7

	// ShaderTypeTransparentSkydome is transparent for skydome.
	ShaderTypeTransparentSkydome ShaderType = 8

	// ShaderTypeTransparentAdditiveUnlitSkydome is transparent additive unlit for skydome.
	ShaderTypeTransparentAdditiveUnlitSkydome ShaderType = 9

	// ShaderTypeInvisible is invisible/not rendered.
	ShaderTypeInvisible ShaderType = 10

	// ShaderTypeBoundary is for boundary surfaces.
	ShaderTypeBoundary ShaderType = 11
)

// String returns the string representation of the shader type.
func (s ShaderType) String() string {
	switch s {
	case ShaderTypeDiffuse:
		return "Diffuse"
	case ShaderTypeTransparent25:
		return "Transparent25"
	case ShaderTypeTransparent50:
		return "Transparent50"
	case ShaderTypeTransparent75:
		return "Transparent75"
	case ShaderTypeTransparentAdditive:
		return "TransparentAdditive"
	case ShaderTypeTransparentAdditiveUnlit:
		return "TransparentAdditiveUnlit"
	case ShaderTypeTransparentMasked:
		return "TransparentMasked"
	case ShaderTypeDiffuseSkydome:
		return "DiffuseSkydome"
	case ShaderTypeTransparentSkydome:
		return "TransparentSkydome"
	case ShaderTypeTransparentAdditiveUnlitSkydome:
		return "TransparentAdditiveUnlitSkydome"
	case ShaderTypeInvisible:
		return "Invisible"
	case ShaderTypeBoundary:
		return "Boundary"
	default:
		return "Unknown"
	}
}
