package exporters

import (
	"fmt"

	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/helpers"
)

// LegacyMeshIntermediateAssetWriter exports legacy meshes in the intermediate mesh format.
type LegacyMeshIntermediateAssetWriter struct {
	TextAssetWriter
	useGroups        bool
	isCollisionMesh  bool
	isFirstMesh      bool
	currentBaseIndex int
}

// NewLegacyMeshIntermediateAssetWriter creates a new LegacyMeshIntermediateAssetWriter.
func NewLegacyMeshIntermediateAssetWriter(useGroups, isCollisionMesh bool) *LegacyMeshIntermediateAssetWriter {
	return &LegacyMeshIntermediateAssetWriter{
		useGroups:        useGroups,
		isCollisionMesh:  isCollisionMesh,
		isFirstMesh:      true,
		currentBaseIndex: 0,
	}
}

// ClearExportData clears the export data and resets state.
func (w *LegacyMeshIntermediateAssetWriter) ClearExportData() {
	w.TextAssetWriter.ClearExportData()
	w.isFirstMesh = true
	w.currentBaseIndex = 0
}

// AddFragmentData adds legacy mesh fragment data to the export.
func (w *LegacyMeshIntermediateAssetWriter) AddFragmentData(data fragments.Fragment) {
	lm, ok := data.(*fragments.LegacyMesh)
	if !ok || lm == nil {
		return
	}

	// Handle collision mesh with polyhedron reference
	if w.isCollisionMesh && lm.PolyhedronReference != nil {
		polyhedron := lm.PolyhedronReference.Polyhedron
		if polyhedron == nil {
			return
		}

		// Write polyhedron vertices
		for _, vertex := range polyhedron.Vertices {
			w.export.WriteString(fmt.Sprintf("v,%v,%v,%v\n",
				vertex.X+lm.Center.X,
				vertex.Z+lm.Center.Z,
				vertex.Y+lm.Center.Y))
		}

		// Write polyhedron faces
		for _, polygon := range polyhedron.Faces {
			w.export.WriteString(fmt.Sprintf("i,0,%d,%d,%d\n",
				w.currentBaseIndex+polygon.Vertex1,
				w.currentBaseIndex+polygon.Vertex2,
				w.currentBaseIndex+polygon.Vertex3))
		}

		return
	}

	// Write material list reference
	if !w.isCollisionMesh && (w.isFirstMesh || w.useGroups) {
		if lm.MaterialList != nil {
			materialName := helpers.CleanName(lm.MaterialList.GetName(), "MaterialList", true)
			w.export.WriteString("ml,")
			w.export.WriteString(materialName)
			w.export.WriteString("\n")
		}
		w.isFirstMesh = false
	}

	// Write vertices
	for _, vertex := range lm.Vertices {
		w.export.WriteString(fmt.Sprintf("v,%v,%v,%v\n",
			vertex.X+lm.Center.X,
			vertex.Z+lm.Center.Z,
			vertex.Y+lm.Center.Y))
	}

	// Write texture UV coordinates
	for _, uv := range lm.TexCoords {
		w.export.WriteString(fmt.Sprintf("uv,%v,%v\n", uv.X, uv.Y))
	}

	// Write normals
	for _, normal := range lm.Normals {
		w.export.WriteString(fmt.Sprintf("n,%v,%v,%v\n", normal.X, normal.Y, normal.Z))
	}

	// Write indices via render groups
	currentPolygon := 0
	for _, renderGroup := range lm.RenderGroups {
		for j := 0; j < renderGroup.PolygonCount; j++ {
			if currentPolygon >= len(lm.Polygons) {
				break
			}
			polygon := lm.Polygons[currentPolygon]
			currentPolygon++

			w.export.WriteString(fmt.Sprintf("i,%d,%d,%d,%d\n",
				renderGroup.MaterialIndex,
				w.currentBaseIndex+polygon.Vertex1,
				w.currentBaseIndex+polygon.Vertex2,
				w.currentBaseIndex+polygon.Vertex3))
		}
	}

	// If no render groups, write indices directly from polygons
	if len(lm.RenderGroups) == 0 {
		for _, polygon := range lm.Polygons {
			w.export.WriteString(fmt.Sprintf("i,%d,%d,%d,%d\n",
				polygon.MaterialIndex,
				w.currentBaseIndex+polygon.Vertex1,
				w.currentBaseIndex+polygon.Vertex2,
				w.currentBaseIndex+polygon.Vertex3))
		}
	}

	// Write bone assignments
	for boneKey, boneValue := range lm.MobPieces {
		w.export.WriteString(fmt.Sprintf("b,%d,%d,%d\n",
			boneKey, boneValue.Start, boneValue.Count))
	}

	// Write animated vertices
	animatedVertices := w.getAnimatedVertices(lm)
	if animatedVertices != nil && !w.isCollisionMesh {
		w.export.WriteString(fmt.Sprintf("ad,%d\n", animatedVertices.GetDelay()))

		frames := animatedVertices.GetFrames()
		for frameIdx, frame := range frames {
			for _, position := range frame {
				w.export.WriteString(fmt.Sprintf("av,%d,%v,%v,%v\n",
					frameIdx,
					position.X+lm.Center.X,
					position.Z+lm.Center.Z,
					position.Y+lm.Center.Y))
			}
		}
	}

	if !w.useGroups {
		w.currentBaseIndex += len(lm.Vertices)
	}
}

// getAnimatedVertices retrieves animated vertices from a legacy mesh reference.
func (w *LegacyMeshIntermediateAssetWriter) getAnimatedVertices(lm *fragments.LegacyMesh) fragments.IAnimatedVertices {
	if lm.AnimatedVerticesReference == nil {
		return nil
	}
	return lm.AnimatedVerticesReference.GetAnimatedVertices()
}

// WriteAssetToFile writes the intermediate mesh data to a file.
func (w *LegacyMeshIntermediateAssetWriter) WriteAssetToFile(fileName string) error {
	if w.export.Len() == 0 {
		return nil
	}

	// Insert header at the beginning
	header := ExportHeaderTitle + "Alternate Mesh Intermediate Format\n"
	content := header + w.export.String()

	// Clear and write the complete content
	w.export.Reset()
	w.export.WriteString(content)

	return w.TextAssetWriter.WriteAssetToFile(fileName)
}
