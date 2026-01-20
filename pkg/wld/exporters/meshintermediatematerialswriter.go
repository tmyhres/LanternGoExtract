package exporters

import (
	"fmt"
	"strings"

	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
)

// MeshIntermediateMaterialsWriter exports material lists in the intermediate format.
type MeshIntermediateMaterialsWriter struct {
	TextAssetWriter
}

// NewMeshIntermediateMaterialsWriter creates a new MeshIntermediateMaterialsWriter.
func NewMeshIntermediateMaterialsWriter() *MeshIntermediateMaterialsWriter {
	return &MeshIntermediateMaterialsWriter{}
}

// AddFragmentData adds material list fragment data to the export.
func (w *MeshIntermediateMaterialsWriter) AddFragmentData(data fragments.Fragment) {
	list, ok := data.(*fragments.MaterialList)
	if !ok || list == nil {
		return
	}

	w.export.WriteString(ExportHeaderTitle + "Material List Intermediate Format\n")
	w.export.WriteString(ExportHeaderFormat + "Index, MaterialName, AnimationTextures, AnimationDelayMs, SkinTextures\n")

	for i := 0; i < len(list.Materials); i++ {
		material := list.Materials[i]
		if material == nil {
			continue
		}

		w.export.WriteString(fmt.Sprintf("%d,", i))

		// Collect all material variants
		allMaterials := make([]*fragments.Material, 0)
		allMaterials = append(allMaterials, material)
		allMaterials = append(allMaterials, list.GetMaterialVariants(material)...)

		for j, currentMaterial := range allMaterials {
			mat := currentMaterial
			if mat == nil && len(allMaterials) > 0 {
				mat = allMaterials[0]
			}
			if mat == nil {
				continue
			}

			w.export.WriteString(getMaterialString(mat))

			if j < list.VariantCount {
				w.export.WriteString(";")
			}
		}

		// Write animation delay
		animDelay := getAnimationDelay(material)
		w.export.WriteString(fmt.Sprintf(",%d\n", animDelay))
	}
}

// getMaterialString builds the material string with bitmap names.
func getMaterialString(material *fragments.Material) string {
	var sb strings.Builder

	materialName := getFullMaterialName(material)
	sb.WriteString(materialName)

	bitmapNames := getAllBitmapNames(material)
	for _, bitmap := range bitmapNames {
		sb.WriteString(":")
		sb.WriteString(bitmap)
	}

	return sb.String()
}

// getFullMaterialName returns the full material name.
func getFullMaterialName(material *fragments.Material) string {
	if material == nil {
		return ""
	}

	name := material.GetName()
	// Remove _MDF suffix if present
	name = strings.TrimSuffix(name, "_MDF")
	return strings.ToLower(name)
}

// getAllBitmapNames returns all bitmap names referenced by a material.
func getAllBitmapNames(material *fragments.Material) []string {
	var bitmapNames []string

	if material == nil || material.BitmapInfoReference == nil {
		return bitmapNames
	}

	// Try to get BitmapInfoReference
	if biRef, ok := material.BitmapInfoReference.(*fragments.BitmapInfoReference); ok {
		if biRef.BitmapInfo != nil {
			for _, bitmapName := range biRef.BitmapInfo.BitmapNames {
				if bitmapName != nil {
					bitmapNames = append(bitmapNames, bitmapName.Filename)
				}
			}
		}
	}

	// Also try direct BitmapInfo
	if bi, ok := material.BitmapInfoReference.(*fragments.BitmapInfo); ok {
		for _, bitmapName := range bi.BitmapNames {
			if bitmapName != nil {
				bitmapNames = append(bitmapNames, bitmapName.Filename)
			}
		}
	}

	return bitmapNames
}

// getAnimationDelay returns the animation delay for a material.
func getAnimationDelay(material *fragments.Material) int32 {
	if material == nil || material.BitmapInfoReference == nil {
		return 0
	}

	// Try BitmapInfoReference
	if biRef, ok := material.BitmapInfoReference.(*fragments.BitmapInfoReference); ok {
		if biRef.BitmapInfo != nil {
			return biRef.BitmapInfo.AnimationDelayMs
		}
	}

	// Try direct BitmapInfo
	if bi, ok := material.BitmapInfoReference.(*fragments.BitmapInfo); ok {
		return bi.AnimationDelayMs
	}

	return 0
}
