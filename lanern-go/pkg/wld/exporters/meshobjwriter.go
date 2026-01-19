package exporters

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/lanterneq/lanern-go/pkg/wld/datatypes"
	"github.com/lanterneq/lanern-go/pkg/wld/fragments"
	"github.com/lanterneq/lanern-go/pkg/wld/helpers"
)

// MeshObjWriter exports meshes to the OBJ format.
type MeshObjWriter struct {
	TextAssetWriter

	// isFirstMesh indicates if this is the first mesh in the export.
	// Zones are made of multiple meshes. The OBJ header is only added when this is set.
	isFirstMesh bool

	// activeMaterial tracks the currently active material.
	// Prevents multiple usemtl declarations of the same material.
	activeMaterial *fragments.Material

	// baseVertex is used when dealing with multiple meshes.
	// The base vertex is added to submesh vertex to get the correct vertex value.
	baseVertex int

	// exportGroups divides the mesh into submeshes if true.
	// Only applies to the zone mesh.
	exportGroups bool

	// exportHiddenGeometry exports invisible and boundary surfaces if true.
	// Only applies to the zone mesh.
	exportHiddenGeometry bool

	// objExportType determines the export type (textured or collision).
	objExportType datatypes.ObjExportType

	// usedVertices tracks the count of used vertices.
	usedVertices int

	// forcedMeshList overrides the mesh list name if set.
	forcedMeshList string

	// frames stores animation frames.
	frames []*strings.Builder

	// isCharacterModel indicates if this is a character model.
	isCharacterModel bool

	// hasCollisionModel tracks if a collision model exists.
	hasCollisionModel bool
}

// NewMeshObjWriter creates a new MeshObjWriter.
func NewMeshObjWriter(exportType datatypes.ObjExportType, exportHiddenGeometry, exportGroups bool, zoneName, forcedMeshList string) *MeshObjWriter {
	return &MeshObjWriter{
		objExportType:        exportType,
		exportHiddenGeometry: exportHiddenGeometry,
		exportGroups:         exportGroups,
		forcedMeshList:       forcedMeshList,
		isFirstMesh:          true,
		frames:               make([]*strings.Builder, 0),
	}
}

// SetIsCharacterModel sets the character model flag.
func (w *MeshObjWriter) SetIsCharacterModel(state bool) {
	w.isCharacterModel = state
}

// AddFragmentData adds mesh fragment data to the export.
func (w *MeshObjWriter) AddFragmentData(data fragments.Fragment) {
	w.AddFragmentDataWithObject(data, nil)
}

// AddFragmentDataWithObject adds mesh fragment data to the export with an optional object instance.
func (w *MeshObjWriter) AddFragmentDataWithObject(fragment fragments.Fragment, associatedObject *fragments.ObjectInstance) {
	mesh, ok := fragment.(*fragments.Mesh)
	if !ok || mesh == nil {
		return
	}

	// Sometimes we get the lowest value of signed int16 - dropped trees, no use trying to adjust
	if associatedObject != nil && math.Round(float64(associatedObject.Position.Z)) <= float64(math.MinInt16) {
		return
	}

	var offset, rotation datatypes.Vec3
	var scale float32 = 1.0

	if associatedObject != nil {
		offset = associatedObject.Position
		rotation = associatedObject.Rotation
		scale = associatedObject.Scale.Y
	}

	// Rotation matrix transform
	pitch := float64(math.Pi/180) * float64(rotation.X)
	roll := float64(math.Pi/180) * float64(rotation.Y)
	yaw := float64(math.Pi/180) * float64(rotation.Z) * -1

	cosa := math.Cos(yaw)
	sina := math.Sin(yaw)
	cosb := math.Cos(pitch)
	sinb := math.Sin(pitch)
	cosc := math.Cos(roll)
	sinc := math.Sin(roll)

	axx := cosa * cosb
	axy := cosa*sinb*sinc - sina*cosc
	axz := cosa*sinb*cosc + sina*sinc

	ayx := sina * cosb
	ayy := sina*sinb*sinc + cosa*cosc
	ayz := sina*sinb*cosc - cosa*sinc

	azx := -sinb
	azy := cosb * sinc
	azz := cosb * cosc

	// Add OBJ header on first mesh
	if w.isFirstMesh && w.objExportType == datatypes.ObjExportTypeTextured {
		var name string
		if w.forcedMeshList != "" {
			name = ObjMaterialHeader + w.forcedMeshList + ".mtl"
		} else {
			suffix := ""
			if w.isCharacterModel {
				suffix = "_0"
			}
			materialListName := helpers.CleanName(mesh.MaterialList.GetName(), "MaterialList", true)
			name = ObjMaterialHeader + materialListName + suffix + ".mtl"
		}
		w.export.WriteString(name + "\n")
		w.isFirstMesh = false
	}

	// Add group name
	if w.exportGroups {
		meshName := helpers.CleanName(mesh.GetName(), "Mesh", true)
		w.export.WriteString("g " + meshName + "\n")
	}

	if mesh.ExportSeparateCollision {
		w.hasCollisionModel = true
	}

	usedVertices := make([]int, 0)
	unusedVertices := make([]int, 0)

	currentPolygon := 0
	faceOutput := &strings.Builder{}

	// First assemble the faces that are needed
	for _, group := range mesh.MaterialGroups {
		textureIndex := group.MaterialIndex
		polygonCount := group.PolygonCount

		shouldExport := true

		if mesh.MaterialList != nil && textureIndex >= 0 && textureIndex < len(mesh.MaterialList.Materials) {
			shaderType := mesh.MaterialList.Materials[textureIndex].ShaderType
			if shaderType == fragments.ShaderTypeBoundary || shaderType == fragments.ShaderTypeInvisible {
				if w.objExportType != datatypes.ObjExportTypeCollision || !w.exportHiddenGeometry {
					shouldExport = false
				}
			}
		}

		activeArray := &usedVertices
		if !shouldExport {
			activeArray = &unusedVertices
		}

		if mesh.MaterialList == nil || textureIndex < 0 || textureIndex >= len(mesh.MaterialList.Materials) {
			continue
		}

		filenameWithoutExtension := getFirstBitmapNameWithoutExtension(mesh.MaterialList.Materials[textureIndex])

		textureChange := ""

		if shouldExport {
			// Material change
			if w.activeMaterial != mesh.MaterialList.Materials[textureIndex] && w.objExportType == datatypes.ObjExportTypeTextured {
				if filenameWithoutExtension == "" {
					textureChange = ObjUseMtlPrefix + "null"
				} else {
					materialPrefix := fragments.GetMaterialPrefix(mesh.MaterialList.Materials[textureIndex].ShaderType)
					textureChange = ObjUseMtlPrefix + materialPrefix + filenameWithoutExtension
				}
				w.activeMaterial = mesh.MaterialList.Materials[textureIndex]
			}
		}

		for j := 0; j < polygonCount; j++ {
			if currentPolygon < 0 || currentPolygon >= len(mesh.Indices) {
				continue
			}

			// Check for non-solid polygons in collision mode
			if !mesh.Indices[currentPolygon].IsSolid && w.objExportType == datatypes.ObjExportTypeCollision {
				activeArray = &unusedVertices
				addIfNotContained(activeArray, mesh.Indices[currentPolygon].Vertex1)
				addIfNotContained(activeArray, mesh.Indices[currentPolygon].Vertex2)
				addIfNotContained(activeArray, mesh.Indices[currentPolygon].Vertex3)
				currentPolygon++
				continue
			}

			if textureChange != "" {
				faceOutput.WriteString(textureChange + "\n")
				textureChange = ""
			}

			vertex1 := mesh.Indices[currentPolygon].Vertex1 + w.baseVertex + 1
			vertex2 := mesh.Indices[currentPolygon].Vertex2 + w.baseVertex + 1
			vertex3 := mesh.Indices[currentPolygon].Vertex3 + w.baseVertex + 1

			if activeArray == &usedVertices {
				index1 := vertex1 - len(unusedVertices)
				index2 := vertex2 - len(unusedVertices)
				index3 := vertex3 - len(unusedVertices)

				// Vertex + UV
				if w.objExportType != datatypes.ObjExportTypeCollision {
					faceOutput.WriteString(fmt.Sprintf("f %d/%d %d/%d %d/%d\n",
						index3, index3, index2, index2, index1, index1))
				} else {
					faceOutput.WriteString(fmt.Sprintf("f %d %d %d\n", index3, index2, index1))
				}
			}

			addIfNotContained(activeArray, mesh.Indices[currentPolygon].Vertex1)
			addIfNotContained(activeArray, mesh.Indices[currentPolygon].Vertex2)
			addIfNotContained(activeArray, mesh.Indices[currentPolygon].Vertex3)

			currentPolygon++
		}
	}

	// For character models, use all vertices
	if w.isCharacterModel {
		usedVertices = make([]int, len(mesh.Vertices))
		for i := range usedVertices {
			usedVertices[i] = i
		}
	} else {
		sort.Ints(usedVertices)
	}

	frameCount := 1
	animatedVertices := getAnimatedVerticesFromMesh(mesh)

	// Avoid OOM errors for zones with objects
	if associatedObject == nil && animatedVertices != nil {
		frameCount += len(animatedVertices.GetFrames())
	}

	for frameIdx := 0; frameIdx < frameCount; frameIdx++ {
		vertexOutput := &strings.Builder{}

		// Add each vertex
		for _, usedVertex := range usedVertices {
			var vertex fragments.Vec3

			if frameIdx == 0 {
				if usedVertex < 0 || usedVertex >= len(mesh.Vertices) {
					continue
				}
				vertex = mesh.Vertices[usedVertex]
			} else {
				if animatedVertices == nil {
					continue
				}
				frames := animatedVertices.GetFrames()
				if frameIdx-1 >= len(frames) || usedVertex >= len(frames[frameIdx-1]) {
					continue
				}
				vertex = frames[frameIdx-1][usedVertex]
			}

			// Apply transformation for scale
			if scale != 1 {
				vertex.X = vertex.X * scale
				vertex.Y = vertex.Y * scale
				vertex.Z = vertex.Z * scale
			}

			// Apply transformation for rotation
			if rotation.X != 0 || rotation.Y != 0 || rotation.Z != 0 {
				px := float64(vertex.X)
				py := float64(vertex.Y)
				pz := float64(vertex.Z)

				x := float32(axx*px + axy*py + axz*pz)
				y := float32(ayx*px + ayy*py + ayz*pz)
				z := float32(azx*px + azy*py + azz*pz)
				vertex = fragments.Vec3{X: x, Y: y, Z: z}
			}

			vertexOutput.WriteString(fmt.Sprintf("v %f %f %f\n",
				-(float64(vertex.X)+float64(mesh.Center.X)+float64(offset.X)),
				float64(vertex.Z)+float64(mesh.Center.Z)+float64(offset.Z),
				float64(vertex.Y)+float64(mesh.Center.Y)+float64(offset.Y)))

			if w.objExportType == datatypes.ObjExportTypeCollision {
				continue
			}

			if usedVertex >= len(mesh.TextureUvCoordinates) {
				vertexOutput.WriteString(fmt.Sprintf("vt %f %f", 0.0, 0.0))
				continue
			}

			vertexUvs := mesh.TextureUvCoordinates[usedVertex]
			vertexOutput.WriteString(fmt.Sprintf("vt %f %f\n", vertexUvs.X, vertexUvs.Y))
		}

		frameContent := vertexOutput.String() + faceOutput.String()

		if frameIdx == 0 {
			w.export.WriteString(frameContent)
		} else {
			frameSb := &strings.Builder{}
			frameSb.WriteString(frameContent)
			w.frames = append(w.frames, frameSb)
		}
	}

	w.baseVertex += len(usedVertices)
}

// WriteAllFrames writes all animation frames to separate files.
func (w *MeshObjWriter) WriteAllFrames(fileName string) error {
	if len(w.frames) <= 1 {
		return nil
	}

	baseName := strings.TrimSuffix(fileName, ".obj")

	for i := 1; i < len(w.frames); i++ {
		w.export = *w.frames[i]
		frameName := fmt.Sprintf("%s_frame%d.obj", baseName, i)
		if err := w.TextAssetWriter.WriteAssetToFile(frameName); err != nil {
			return err
		}
	}

	return nil
}

// WriteAssetToFile writes the OBJ mesh data to a file.
func (w *MeshObjWriter) WriteAssetToFile(fileName string) error {
	if w.objExportType == datatypes.ObjExportTypeCollision && !w.hasCollisionModel {
		return nil
	}

	return w.TextAssetWriter.WriteAssetToFile(fileName)
}

// ClearExportData clears the export data and resets state.
func (w *MeshObjWriter) ClearExportData() {
	w.TextAssetWriter.ClearExportData()
	w.activeMaterial = nil
	w.usedVertices = 0
	w.baseVertex = 0
	w.isFirstMesh = true
	w.frames = make([]*strings.Builder, 0)
}

// addIfNotContained adds an element to a slice if it's not already present.
func addIfNotContained(list *[]int, element int) {
	for _, v := range *list {
		if v == element {
			return
		}
	}
	*list = append(*list, element)
}

// getFirstBitmapNameWithoutExtension returns the first bitmap name without extension.
func getFirstBitmapNameWithoutExtension(material *fragments.Material) string {
	if material == nil || material.BitmapInfoReference == nil {
		return ""
	}

	// Try BitmapInfoReference
	if biRef, ok := material.BitmapInfoReference.(*fragments.BitmapInfoReference); ok {
		if biRef.BitmapInfo != nil && len(biRef.BitmapInfo.BitmapNames) > 0 {
			filename := biRef.BitmapInfo.BitmapNames[0].Filename
			return strings.TrimSuffix(strings.ToLower(filename), ".bmp")
		}
	}

	// Try direct BitmapInfo
	if bi, ok := material.BitmapInfoReference.(*fragments.BitmapInfo); ok {
		if len(bi.BitmapNames) > 0 {
			filename := bi.BitmapNames[0].Filename
			return strings.TrimSuffix(strings.ToLower(filename), ".bmp")
		}
	}

	return ""
}

// getFirstBitmapExportFilename returns the first bitmap export filename.
func getFirstBitmapExportFilename(material *fragments.Material) string {
	if material == nil || material.BitmapInfoReference == nil {
		return ""
	}

	// Try BitmapInfoReference
	if biRef, ok := material.BitmapInfoReference.(*fragments.BitmapInfoReference); ok {
		if biRef.BitmapInfo != nil && len(biRef.BitmapInfo.BitmapNames) > 0 {
			filename := biRef.BitmapInfo.BitmapNames[0].Filename
			return strings.ToLower(strings.TrimSuffix(filename, ".bmp")) + ".png"
		}
	}

	// Try direct BitmapInfo
	if bi, ok := material.BitmapInfoReference.(*fragments.BitmapInfo); ok {
		if len(bi.BitmapNames) > 0 {
			filename := bi.BitmapNames[0].Filename
			return strings.ToLower(strings.TrimSuffix(filename, ".bmp")) + ".png"
		}
	}

	return ""
}

// getAnimatedVerticesFromMesh retrieves animated vertices from a mesh.
func getAnimatedVerticesFromMesh(mesh *fragments.Mesh) fragments.IAnimatedVertices {
	if mesh == nil || mesh.AnimatedVerticesReference == nil {
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
