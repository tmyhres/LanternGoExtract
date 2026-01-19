package helpers

import (
	"github.com/lanterneq/lanern-go/pkg/wld/fragments"
)

// MaterialFixer fixes issues with EverQuest materials.
// Currently, only fixes incorrect shader assignment but can be extended in the future.

// materialsToFix maps material names to their correct shader types.
var materialsToFix = map[string]fragments.ShaderType{
	// Tree materials that need transparent masked shader
	"TREE7_MDF":    fragments.ShaderTypeTransparentMasked,
	"TREE9B1_MDF":  fragments.ShaderTypeTransparentMasked,
	"TREE16_MDF":   fragments.ShaderTypeTransparentMasked,
	"TREE16B1_MDF": fragments.ShaderTypeTransparentMasked,
	"TREE17_MDF":   fragments.ShaderTypeTransparentMasked,
	"TREE18_MDF":   fragments.ShaderTypeTransparentMasked,
	"TREE18B1_MDF": fragments.ShaderTypeTransparentMasked,
	"TREE20_MDF":   fragments.ShaderTypeTransparentMasked,
	"TREE20B1_MDF": fragments.ShaderTypeTransparentMasked,
	"TREE21_MDF":   fragments.ShaderTypeTransparentMasked,
	"TREE22_MDF":   fragments.ShaderTypeTransparentMasked,
	"TREE22B1_MDF": fragments.ShaderTypeTransparentMasked,
	"TOP_MDF":      fragments.ShaderTypeTransparentMasked,
	"TOPB_MDF":     fragments.ShaderTypeTransparentMasked,
	"FURPILE1_MDF": fragments.ShaderTypeTransparentMasked,
	"BEARRUG_MDF":  fragments.ShaderTypeTransparentMasked,

	// Fire material needs transparent additive unlit
	"FIRE1_MDF": fragments.ShaderTypeTransparentAdditiveUnlit,

	// Ice material should be invisible
	"ICE1_MDF": fragments.ShaderTypeInvisible,

	// Cloud materials need transparent skydome
	"AIRCLOUD_MDF":    fragments.ShaderTypeTransparentSkydome,
	"NORMALCLOUD_MDF": fragments.ShaderTypeTransparentSkydome,
}

// WldFile is an interface representing a WLD file that can provide fragments.
// This allows MaterialFixer to work without a direct dependency on the WLD package.
type WldFile interface {
	GetFragmentsOfType(fragmentType uint32) []fragments.Fragment
	GetFragmentByName(name string) fragments.Fragment
}

// FixMaterials fixes shader assignments for known problematic materials in a WLD file.
// This function iterates through all materials and corrects their shader types
// based on known issues in EverQuest's material definitions.
func FixMaterials(materials []*fragments.Material) {
	for _, material := range materials {
		if material == nil {
			continue
		}

		if shaderType, needsFix := materialsToFix[material.Name]; needsFix {
			material.ShaderType = shaderType
		}
	}
}

// FixMaterialsFromFragments fixes shader assignments using a slice of fragments.
// This is useful when working directly with parsed fragment data.
func FixMaterialsFromFragments(frags []fragments.Fragment) {
	for _, frag := range frags {
		material, ok := frag.(*fragments.Material)
		if !ok || material == nil {
			continue
		}

		if shaderType, needsFix := materialsToFix[material.Name]; needsFix {
			material.ShaderType = shaderType
		}
	}
}

// FixMaterial fixes a single material's shader assignment if needed.
// Returns true if the material was fixed, false otherwise.
func FixMaterial(material *fragments.Material) bool {
	if material == nil {
		return false
	}

	if shaderType, needsFix := materialsToFix[material.Name]; needsFix {
		material.ShaderType = shaderType
		return true
	}

	return false
}

// GetMaterialFixList returns a copy of the materials that need fixing.
// This can be used for debugging or logging purposes.
func GetMaterialFixList() map[string]fragments.ShaderType {
	result := make(map[string]fragments.ShaderType, len(materialsToFix))
	for k, v := range materialsToFix {
		result[k] = v
	}
	return result
}
