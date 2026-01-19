package exporters

import (
	"fmt"

	"github.com/lanterneq/lanern-go/pkg/wld/datatypes"
	"github.com/lanterneq/lanern-go/pkg/wld/fragments"
	"github.com/lanterneq/lanern-go/pkg/wld/helpers"
)

// MeshIntermediateAssetWriter exports meshes in the intermediate mesh format.
type MeshIntermediateAssetWriter struct {
	TextAssetWriter
	useGroups       bool
	isCollisionMesh bool
	isFirstMesh     bool
	currentBaseIndex int
}

// NewMeshIntermediateAssetWriter creates a new MeshIntermediateAssetWriter.
func NewMeshIntermediateAssetWriter(useGroups, isCollisionMesh bool) *MeshIntermediateAssetWriter {
	return &MeshIntermediateAssetWriter{
		useGroups:        useGroups,
		isCollisionMesh:  isCollisionMesh,
		isFirstMesh:      true,
		currentBaseIndex: 0,
	}
}

// ClearExportData clears the export data and resets state.
func (w *MeshIntermediateAssetWriter) ClearExportData() {
	w.TextAssetWriter.ClearExportData()
	w.isFirstMesh = true
	w.currentBaseIndex = 0
}

// AddFragmentData adds mesh fragment data to the export.
func (w *MeshIntermediateAssetWriter) AddFragmentData(data fragments.Fragment) {
	mesh, ok := data.(*fragments.Mesh)
	if !ok || mesh == nil {
		return
	}

	usedVertices := make(map[int]bool)
	newIndices := make([]*datatypes.Polygon, 0, len(mesh.Indices))

	currentPolygon := 0

	// Process material groups
	for _, group := range mesh.MaterialGroups {
		for i := 0; i < group.PolygonCount; i++ {
			if currentPolygon >= len(mesh.Indices) {
				break
			}
			polygon := mesh.Indices[currentPolygon]
			newIndices = append(newIndices, polygon.Copy())
			currentPolygon++

			if !polygon.IsSolid && w.isCollisionMesh {
				continue
			}

			usedVertices[polygon.Vertex1] = true
			usedVertices[polygon.Vertex2] = true
			usedVertices[polygon.Vertex3] = true
		}
	}

	// For non-collision meshes, use all vertices
	if !w.isCollisionMesh {
		usedVertices = make(map[int]bool)
		for i := 0; i < len(mesh.Vertices); i++ {
			usedVertices[i] = true
		}
	}

	// Adjust vertex indices for unused vertices
	unusedVertices := 0
	for i := len(mesh.Vertices) - 1; i >= 0; i-- {
		if usedVertices[i] {
			continue
		}

		unusedVertices++

		for _, polygon := range newIndices {
			if polygon.Vertex1 >= i && polygon.Vertex1 != 0 {
				polygon.Vertex1--
			}
			if polygon.Vertex2 >= i && polygon.Vertex2 != 0 {
				polygon.Vertex2--
			}
			if polygon.Vertex3 >= i && polygon.Vertex3 != 0 {
				polygon.Vertex3--
			}
		}
	}

	// Write material list reference
	if !w.isCollisionMesh && (w.isFirstMesh || w.useGroups) {
		if mesh.MaterialList != nil {
			materialName := helpers.CleanName(mesh.MaterialList.GetName(), "MaterialList", true)
			w.export.WriteString("ml,")
			w.export.WriteString(materialName)
			w.export.WriteString("\n")
		}
		w.isFirstMesh = false
	}

	// Write vertices
	for i := 0; i < len(mesh.Vertices); i++ {
		if !usedVertices[i] {
			continue
		}

		vertex := mesh.Vertices[i]
		w.export.WriteString(fmt.Sprintf("v,%v,%v,%v\n",
			vertex.X+mesh.Center.X,
			vertex.Z+mesh.Center.Z,
			vertex.Y+mesh.Center.Y))
	}

	// Write texture UV coordinates
	for i := 0; i < len(mesh.TextureUvCoordinates); i++ {
		if !usedVertices[i] || w.isCollisionMesh {
			continue
		}

		textureUv := mesh.TextureUvCoordinates[i]
		w.export.WriteString(fmt.Sprintf("uv,%v,%v\n", textureUv.X, textureUv.Y))
	}

	// Write normals
	for i := 0; i < len(mesh.Normals); i++ {
		if !usedVertices[i] || w.isCollisionMesh {
			continue
		}

		normal := mesh.Normals[i]
		w.export.WriteString(fmt.Sprintf("n,%v,%v,%v\n", normal.X, normal.Y, normal.Z))
	}

	// Write colors
	for i := 0; i < len(mesh.Colors); i++ {
		if !usedVertices[i] || w.isCollisionMesh {
			continue
		}

		vertexColor := mesh.Colors[i]
		w.export.WriteString(fmt.Sprintf("c,%d,%d,%d,%d\n",
			vertexColor.B, vertexColor.G, vertexColor.R, vertexColor.A))
	}

	// Write indices
	currentPolygon = 0
	for _, group := range mesh.MaterialGroups {
		for i := 0; i < group.PolygonCount; i++ {
			if currentPolygon >= len(newIndices) {
				break
			}
			polygon := newIndices[currentPolygon]
			currentPolygon++

			w.export.WriteString(fmt.Sprintf("i,%d,%d,%d,%d\n",
				group.MaterialIndex,
				w.currentBaseIndex+polygon.Vertex1,
				w.currentBaseIndex+polygon.Vertex2,
				w.currentBaseIndex+polygon.Vertex3))
		}
	}

	// Write bone assignments
	for boneKey, boneValue := range mesh.MobPieces {
		w.export.WriteString(fmt.Sprintf("b,%d,%d,%d\n",
			boneKey, boneValue.Start, boneValue.Count))
	}

	// Write animated vertices
	animatedVertices := w.getAnimatedVertices(mesh)
	if animatedVertices != nil && !w.isCollisionMesh {
		w.export.WriteString(fmt.Sprintf("ad,%d\n", animatedVertices.GetDelay()))

		frames := animatedVertices.GetFrames()
		for frameIdx, frame := range frames {
			for _, position := range frame {
				w.export.WriteString(fmt.Sprintf("av,%d,%v,%v,%v\n",
					frameIdx,
					position.X+mesh.Center.X,
					position.Z+mesh.Center.Z,
					position.Y+mesh.Center.Y))
			}
		}
	}

	if !w.useGroups {
		w.currentBaseIndex += len(mesh.Vertices) - unusedVertices
	}
}

// getAnimatedVertices retrieves animated vertices from a mesh reference.
func (w *MeshIntermediateAssetWriter) getAnimatedVertices(mesh *fragments.Mesh) fragments.IAnimatedVertices {
	if mesh.AnimatedVerticesReference == nil {
		return nil
	}

	// Try direct cast to IAnimatedVertices
	if animVerts, ok := mesh.AnimatedVerticesReference.(fragments.IAnimatedVertices); ok {
		return animVerts
	}

	// Try via reference
	if animRef, ok := mesh.AnimatedVerticesReference.(*fragments.MeshAnimatedVerticesReference); ok {
		return animRef.GetAnimatedVertices()
	}

	return nil
}

// WriteAssetToFile writes the intermediate mesh data to a file.
func (w *MeshIntermediateAssetWriter) WriteAssetToFile(fileName string) error {
	if w.export.Len() == 0 {
		return nil
	}

	// Insert header at the beginning
	header := ExportHeaderTitle + "Mesh Intermediate Format\n"
	content := header + w.export.String()

	// Clear and write the complete content
	w.export.Reset()
	w.export.WriteString(content)

	return w.TextAssetWriter.WriteAssetToFile(fileName)
}
