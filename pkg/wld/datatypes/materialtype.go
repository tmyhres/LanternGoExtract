package datatypes

// MaterialType represents the type of material/shader used for rendering.
type MaterialType int

const (
	// MaterialTypeBoundary is used for boundaries that are not rendered.
	// TextInfoReference can be null or have reference.
	MaterialTypeBoundary MaterialType = 0x0

	// MaterialTypeDiffuse is the standard diffuse shader.
	MaterialTypeDiffuse MaterialType = 0x01

	// MaterialTypeDiffuse2 is a diffuse variant.
	MaterialTypeDiffuse2 MaterialType = 0x02

	// MaterialTypeTransparent50 is transparent with 0.5 blend strength.
	MaterialTypeTransparent50 MaterialType = 0x05

	// MaterialTypeTransparent25 is transparent with 0.25 blend strength.
	MaterialTypeTransparent25 MaterialType = 0x09

	// MaterialTypeTransparent75 is transparent with 0.75 blend strength.
	MaterialTypeTransparent75 MaterialType = 0x0A

	// MaterialTypeTransparentMaskedPassable is for non-solid surfaces
	// that should not really be masked.
	MaterialTypeTransparentMaskedPassable MaterialType = 0x07

	// MaterialTypeTransparentAdditiveUnlit is transparent additive unlit.
	MaterialTypeTransparentAdditiveUnlit MaterialType = 0x0B

	// MaterialTypeTransparentMasked is transparent masked.
	MaterialTypeTransparentMasked MaterialType = 0x13

	// MaterialTypeDiffuse3 is a diffuse variant.
	MaterialTypeDiffuse3 MaterialType = 0x14

	// MaterialTypeDiffuse4 is a diffuse variant.
	MaterialTypeDiffuse4 MaterialType = 0x15

	// MaterialTypeTransparentAdditive is transparent additive.
	MaterialTypeTransparentAdditive MaterialType = 0x17

	// MaterialTypeDiffuse5 is a diffuse variant.
	MaterialTypeDiffuse5 MaterialType = 0x19

	// MaterialTypeInvisibleUnknown is an invisible unknown type.
	MaterialTypeInvisibleUnknown MaterialType = 0x53

	// MaterialTypeDiffuse6 is a diffuse variant.
	MaterialTypeDiffuse6 MaterialType = 0x553

	// MaterialTypeCompleteUnknown is a completely unknown type.
	// TODO: Analyze this
	MaterialTypeCompleteUnknown MaterialType = 0x1A

	// MaterialTypeDiffuse7 is a diffuse variant.
	MaterialTypeDiffuse7 MaterialType = 0x12

	// MaterialTypeDiffuse8 is a diffuse variant.
	MaterialTypeDiffuse8 MaterialType = 0x31

	// MaterialTypeInvisibleUnknown2 is an invisible unknown type.
	MaterialTypeInvisibleUnknown2 MaterialType = 0x4B

	// MaterialTypeDiffuseSkydome is for skydome diffuse. Need to confirm.
	MaterialTypeDiffuseSkydome MaterialType = 0x0D

	// MaterialTypeTransparentSkydome is for skydome transparent. Need to confirm.
	MaterialTypeTransparentSkydome MaterialType = 0x0F

	// MaterialTypeTransparentAdditiveUnlitSkydome is for skydome transparent additive unlit.
	MaterialTypeTransparentAdditiveUnlitSkydome MaterialType = 0x10

	// MaterialTypeInvisibleUnknown3 is an invisible unknown type.
	MaterialTypeInvisibleUnknown3 MaterialType = 0x03
)
