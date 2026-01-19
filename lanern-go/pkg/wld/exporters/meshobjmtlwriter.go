package exporters

import (
	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/fragments"
)

// MeshObjMtlWriter exports material lists to the OBJ MTL format.
type MeshObjMtlWriter struct {
	TextAssetWriter
	exportHiddenGeometry bool
	modelName            string
	skinID               int
}

// NewMeshObjMtlWriter creates a new MeshObjMtlWriter.
func NewMeshObjMtlWriter(exportHiddenGeometry bool, modelName string) *MeshObjMtlWriter {
	return &MeshObjMtlWriter{
		exportHiddenGeometry: exportHiddenGeometry,
		modelName:            modelName,
	}
}

// SetSkinID sets the skin ID for material variants.
func (w *MeshObjMtlWriter) SetSkinID(id int) {
	w.skinID = id
}

// AddFragmentData adds material list fragment data to the export.
func (w *MeshObjMtlWriter) AddFragmentData(data fragments.Fragment) {
	list, ok := data.(*fragments.MaterialList)
	if !ok || list == nil {
		return
	}

	createdNullMaterial := false

	for _, material := range list.Materials {
		skinMaterial := material

		if w.skinID != 0 {
			skinIndex := w.skinID - 1
			variants := list.GetMaterialVariants(skinMaterial)

			if skinIndex >= 0 && skinIndex < len(variants) && variants[skinIndex] != nil {
				skinMaterial = variants[skinIndex]
			}
		}

		filenameWithoutExtension := getFirstBitmapNameWithoutExtension(skinMaterial)

		if filenameWithoutExtension == "" {
			if !createdNullMaterial {
				w.export.WriteString(ObjNewMaterialPrefix + " null\n")
				w.export.WriteString("Ka 1.000 1.000 1.000\n")
				w.export.WriteString("Kd 1.000 1.000 1.000\n")
				w.export.WriteString("Ks 0.000 0.000 0.000\n")
				w.export.WriteString("d 1.0 \n")
				w.export.WriteString("illum 2\n")
				createdNullMaterial = true
			}
			continue
		}

		if skinMaterial.ShaderType == fragments.ShaderTypeInvisible && !w.exportHiddenGeometry {
			continue
		}

		materialPrefix := fragments.GetMaterialPrefix(material.ShaderType)
		baseBitmapName := getFirstBitmapNameWithoutExtension(material)

		w.export.WriteString(ObjNewMaterialPrefix + " " + materialPrefix + baseBitmapName + "\n")
		w.export.WriteString("Ka 1.000 1.000 1.000\n")
		w.export.WriteString("Kd 1.000 1.000 1.000\n")
		w.export.WriteString("Ks 0.000 0.000 0.000\n")
		w.export.WriteString("d 1.0 \n")
		w.export.WriteString("illum 2\n")

		textureFilename := getFirstBitmapExportFilename(skinMaterial)
		w.export.WriteString("map_Kd Textures/" + textureFilename + "\n")
	}
}
