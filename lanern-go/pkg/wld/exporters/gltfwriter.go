// Package exporters provides export functionality for WLD data to various formats.
package exporters

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/lanterneq/lanern-go/pkg/wld/datatypes"
	"github.com/lanterneq/lanern-go/pkg/wld/fragments"
	"github.com/lanterneq/lanern-go/pkg/wld/helpers"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

// GltfExportFormat defines the output format for glTF export.
type GltfExportFormat int

const (
	// GltfExportFormatGlTF exports separate .gltf json file, .bin binary, and images externally referenced.
	GltfExportFormatGlTF GltfExportFormat = 0
	// GltfExportFormatGlb exports one binary file with json metadata and images packaged within.
	GltfExportFormatGlb GltfExportFormat = 1
)

// ModelGenerationMode defines how meshes are combined.
type ModelGenerationMode int

const (
	// ModelGenerationModeCombine combines all meshes into one.
	ModelGenerationModeCombine ModelGenerationMode = 0
	// ModelGenerationModeSeparate keeps every mesh separated.
	ModelGenerationModeSeparate ModelGenerationMode = 1
)

const (
	materialRoughness      = 0.9
	materialInvisName      = "Invis"
	materialBlankName      = "Blank"
	defaultModelPoseAnimKey = "pos"
)

// defaultVertexColor is black with full alpha.
var defaultVertexColor = [4]float32{0.0, 0.0, 0.0, 1.0}

// shaderTypesThatNeedAlpha are shader types that need alpha added to images.
var shaderTypesThatNeedAlpha = map[fragments.ShaderType]bool{
	fragments.ShaderTypeTransparent25:                   true,
	fragments.ShaderTypeTransparent50:                   true,
	fragments.ShaderTypeTransparent75:                   true,
	fragments.ShaderTypeTransparentAdditive:             true,
	fragments.ShaderTypeTransparentAdditiveUnlit:        true,
	fragments.ShaderTypeTransparentAdditiveUnlitSkydome: true,
	fragments.ShaderTypeTransparentSkydome:              true,
}

// loopedAnimationKeys are animation keys that should loop.
var loopedAnimationKeys = map[string]bool{
	"pos": true, // name is used for animated objects
	"p01": true, // Stand
	"l01": true, // Walk
	"l02": true, // Run
	"l05": true, // falling
	"l06": true, // crouch walk
	"l07": true, // climbing
	"l09": true, // swim treading
	"p03": true, // rotating
	"p06": true, // swim
	"p07": true, // sitting
	"p08": true, // stand (arms at sides)
	"sky": true,
}

// GltfWriter handles exporting WLD data to glTF format.
type GltfWriter struct {
	// Materials is the map of material names to glTF material indices.
	Materials map[string]uint32

	exportVertexColors    bool
	exportFormat          GltfExportFormat
	meshMaterialsToSkip   map[string]bool
	sharedMeshes          map[string]*meshData
	skeletons             map[string]*skeletonData
	doc                   *gltf.Document
	combinedMesh          *meshBuilder
	textureIndices        map[string]uint32
	rootNode              uint32
	nodeCount             uint32
}

// meshData holds mesh building data.
type meshData struct {
	positions  [][3]float32
	normals    [][3]float32
	uvs        [][2]float32
	colors     [][4]float32
	joints     [][4]uint16
	weights    [][4]float32
	indices    []uint32
	primitives map[uint32]*primitiveData
}

// primitiveData holds primitive data for a material.
type primitiveData struct {
	indices []uint32
}

// meshBuilder builds mesh data.
type meshBuilder struct {
	name       string
	isSkinned  bool
	hasColors  bool
	meshData   *meshData
	vertexMap  map[vertexKey]uint32
}

// vertexKey uniquely identifies a vertex.
type vertexKey struct {
	posX, posY, posZ    float32
	normX, normY, normZ float32
	u, v                float32
	colorR, colorG, colorB, colorA float32
	joint               uint16
}

// skeletonData holds skeleton node data.
type skeletonData struct {
	nodes      []uint32
	boneNames  []string
	inverseBindMatrices []float32
}

// NewGltfWriter creates a new GltfWriter.
func NewGltfWriter(exportVertexColors bool, exportFormat GltfExportFormat) *GltfWriter {
	doc := gltf.NewDocument()
	doc.Asset.Generator = "LanternGoExtract"

	// Create a root scene
	doc.Scenes = append(doc.Scenes, &gltf.Scene{
		Name: "Scene",
	})
	doc.Scene = gltf.Index(0)

	return &GltfWriter{
		Materials:           make(map[string]uint32),
		exportVertexColors:  exportVertexColors,
		exportFormat:        exportFormat,
		meshMaterialsToSkip: make(map[string]bool),
		sharedMeshes:        make(map[string]*meshData),
		skeletons:           make(map[string]*skeletonData),
		doc:                 doc,
		textureIndices:      make(map[string]uint32),
	}
}

// CopyMaterialList copies materials from another GltfWriter.
func (w *GltfWriter) CopyMaterialList(other *GltfWriter) {
	w.Materials = other.Materials
	w.doc.Materials = other.doc.Materials
	w.doc.Textures = other.doc.Textures
	w.doc.Images = other.doc.Images
	w.textureIndices = other.textureIndices
}

// GenerateGltfMaterials generates glTF materials from material lists.
func (w *GltfWriter) GenerateGltfMaterials(materialLists []*fragments.MaterialList, textureImageFolder string) {
	if len(w.Materials) == 0 {
		w.addBlankMaterial()
	}

	for _, materialList := range materialLists {
		if materialList == nil {
			continue
		}

		for _, eqMaterial := range materialList.Materials {
			if eqMaterial == nil {
				continue
			}

			materialName := getMaterialName(eqMaterial)

			if _, exists := w.Materials[materialName]; exists {
				continue
			}

			if eqMaterial.ShaderType == fragments.ShaderTypeBoundary {
				w.meshMaterialsToSkip[materialName] = true
				continue
			}

			if eqMaterial.ShaderType == fragments.ShaderTypeInvisible {
				w.addInvisibleMaterial(materialName)
				continue
			}

			imageFileNameWithoutExtension := gltfGetBitmapNameWithoutExtension(eqMaterial)
			if imageFileNameWithoutExtension == "" {
				continue
			}

			imagePath := textureImageFolder + gltfGetBitmapExportFilename(eqMaterial)

			// Create material
			mat := &gltf.Material{
				Name: materialName,
				PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
					MetallicFactor:  gltf.Float(0.0),
					RoughnessFactor: gltf.Float(materialRoughness),
				},
				DoubleSided: false,
			}

			// Add texture if image exists
			if _, err := os.Stat(imagePath); err == nil {
				textureIdx := w.addTexture(imagePath, imageFileNameWithoutExtension)
				mat.PBRMetallicRoughness.BaseColorTexture = &gltf.TextureInfo{
					Index: textureIdx,
				}
			}

			// Set alpha mode based on shader type
			switch eqMaterial.ShaderType {
			case fragments.ShaderTypeTransparent25:
				mat.AlphaMode = gltf.AlphaMask
				mat.AlphaCutoff = gltf.Float(0.25)
			case fragments.ShaderTypeTransparent50, fragments.ShaderTypeTransparentMasked:
				mat.AlphaMode = gltf.AlphaMask
				mat.AlphaCutoff = gltf.Float(0.5)
			case fragments.ShaderTypeTransparent75:
				mat.AlphaMode = gltf.AlphaMask
				mat.AlphaCutoff = gltf.Float(0.75)
			case fragments.ShaderTypeTransparentAdditive,
				fragments.ShaderTypeTransparentAdditiveUnlit,
				fragments.ShaderTypeTransparentSkydome,
				fragments.ShaderTypeTransparentAdditiveUnlitSkydome:
				mat.AlphaMode = gltf.AlphaBlend
			default:
				mat.AlphaMode = gltf.AlphaOpaque
			}

			// Set unlit extension for certain shader types
			if eqMaterial.ShaderType == fragments.ShaderTypeTransparentAdditiveUnlit ||
				eqMaterial.ShaderType == fragments.ShaderTypeDiffuseSkydome ||
				eqMaterial.ShaderType == fragments.ShaderTypeTransparentAdditiveUnlitSkydome {
				mat.Extensions = map[string]interface{}{
					"KHR_materials_unlit": map[string]interface{}{},
				}
				// Add the extension to the document if not already present
				hasUnlit := false
				for _, ext := range w.doc.ExtensionsUsed {
					if ext == "KHR_materials_unlit" {
						hasUnlit = true
						break
					}
				}
				if !hasUnlit {
					w.doc.ExtensionsUsed = append(w.doc.ExtensionsUsed, "KHR_materials_unlit")
				}
			}

			matIdx := uint32(len(w.doc.Materials))
			w.doc.Materials = append(w.doc.Materials, mat)
			w.Materials[materialName] = matIdx
		}
	}
}

// addTexture adds a texture to the document and returns its index.
func (w *GltfWriter) addTexture(imagePath, imageName string) uint32 {
	if idx, exists := w.textureIndices[imagePath]; exists {
		return idx
	}

	// Add image
	imageIdx := uint32(len(w.doc.Images))
	w.doc.Images = append(w.doc.Images, &gltf.Image{
		Name: imageName,
		URI:  "Textures/" + filepath.Base(imagePath),
	})

	// Add texture
	textureIdx := uint32(len(w.doc.Textures))
	w.doc.Textures = append(w.doc.Textures, &gltf.Texture{
		Name:   imageName,
		Source: gltf.Index(imageIdx),
	})

	w.textureIndices[imagePath] = textureIdx
	return textureIdx
}

// addBlankMaterial adds a blank white material.
func (w *GltfWriter) addBlankMaterial() {
	mat := &gltf.Material{
		Name: materialBlankName,
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0.0),
			RoughnessFactor: gltf.Float(materialRoughness),
		},
		DoubleSided: false,
	}
	matIdx := uint32(len(w.doc.Materials))
	w.doc.Materials = append(w.doc.Materials, mat)
	w.Materials[materialBlankName] = matIdx
}

// addInvisibleMaterial adds an invisible material.
func (w *GltfWriter) addInvisibleMaterial(name string) {
	mat := &gltf.Material{
		Name: name,
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 0},
			MetallicFactor:  gltf.Float(0.0),
			RoughnessFactor: gltf.Float(materialRoughness),
		},
		AlphaMode:   gltf.AlphaMask,
		DoubleSided: false,
	}
	matIdx := uint32(len(w.doc.Materials))
	w.doc.Materials = append(w.doc.Materials, mat)
	w.Materials[name] = matIdx
}

// AddFragmentData adds mesh fragment data with default settings.
func (w *GltfWriter) AddFragmentData(mesh *fragments.Mesh) {
	w.AddFragmentDataWithOptions(mesh, ModelGenerationModeSeparate, false, "", -1, nil, 0, false)
}

// AddFragmentDataWithSkeleton adds mesh fragment data with skeleton.
func (w *GltfWriter) AddFragmentDataWithSkeleton(mesh *fragments.Mesh, skeleton *fragments.SkeletonHierarchy, meshNameOverride string, singularBoneIndex int) {
	if _, exists := w.skeletons[skeleton.ModelBase]; !exists {
		w.addNewSkeleton(skeleton)
	}

	w.AddFragmentDataWithOptions(mesh, ModelGenerationModeCombine, true, meshNameOverride, singularBoneIndex, nil, 0, false)
}

// AddFragmentDataWithOptions adds mesh fragment data with full options.
func (w *GltfWriter) AddFragmentDataWithOptions(
	mesh *fragments.Mesh,
	generationMode ModelGenerationMode,
	isSkinned bool,
	meshNameOverride string,
	singularBoneIndex int,
	objectInstance *fragments.ObjectInstance,
	instanceIndex int,
	isZoneMesh bool,
) {
	meshName := meshNameOverride
	if meshName == "" {
		meshName = helpers.CleanMeshName(mesh.GetName())
	}

	canExportVertexColors := w.exportVertexColors &&
		((objectInstance != nil && hasObjectColors(objectInstance)) ||
			len(mesh.Colors) > 0)

	// Check for shared mesh reuse
	if mesh.AnimatedVerticesReference != nil && !canExportVertexColors && objectInstance != nil {
		if _, exists := w.sharedMeshes[meshName]; exists {
			if generationMode == ModelGenerationModeSeparate {
				// Add instance of existing mesh - for now just skip
				// Full implementation would add a node referencing the mesh
			}
			return
		}
	}

	// Update mesh name for instances with unique data
	if objectInstance != nil && (canExportVertexColors || mesh.AnimatedVerticesReference != nil) {
		meshName = fmt.Sprintf("%s.%02d", meshName, instanceIndex)
	}

	// Get or create mesh builder
	var builder *meshBuilder
	if generationMode == ModelGenerationModeCombine {
		if w.combinedMesh == nil {
			w.combinedMesh = newMeshBuilder(meshName, isSkinned, canExportVertexColors)
		}
		builder = w.combinedMesh
	} else {
		builder = newMeshBuilder(meshName, isSkinned, canExportVertexColors)
	}

	// Process triangles
	polygonIndex := 0
	for _, materialGroup := range mesh.MaterialGroups {
		if mesh.MaterialList == nil || materialGroup.MaterialIndex >= len(mesh.MaterialList.Materials) {
			polygonIndex += materialGroup.PolygonCount
			continue
		}

		material := mesh.MaterialList.Materials[materialGroup.MaterialIndex]
		materialName := getMaterialName(material)

		if w.meshMaterialsToSkip[materialName] {
			polygonIndex += materialGroup.PolygonCount
			continue
		}

		matIdx, ok := w.Materials[materialName]
		if !ok {
			matIdx = w.Materials[materialBlankName]
		}

		for i := 0; i < materialGroup.PolygonCount; i++ {
			if polygonIndex >= len(mesh.Indices) {
				break
			}
			w.addTriangleToMesh(builder, mesh, polygonIndex, canExportVertexColors, isSkinned, singularBoneIndex, objectInstance, matIdx)
			polygonIndex++
		}
	}

	if generationMode == ModelGenerationModeSeparate {
		// Build and add the mesh to the scene
		transform := identityMatrix()
		if isZoneMesh {
			transform = correctedWorldMatrix()
		} else {
			transform = mirrorXAxisMatrix()
		}

		if objectInstance != nil {
			objTransform := createTransformMatrixForObjectInstance(objectInstance)
			transform = multiplyMatrices(objTransform, transform)
		}

		w.addMeshToScene(builder, transform)
		w.sharedMeshes[meshName] = builder.meshData
	}
}

// addTriangleToMesh adds a triangle to the mesh builder.
func (w *GltfWriter) addTriangleToMesh(
	builder *meshBuilder,
	mesh *fragments.Mesh,
	polygonIndex int,
	canExportVertexColors bool,
	isSkinned bool,
	singularBoneIndex int,
	objectInstance *fragments.ObjectInstance,
	matIdx uint32,
) {
	triangle := mesh.Indices[polygonIndex]
	vertexIndices := [3]int{triangle.Vertex1, triangle.Vertex2, triangle.Vertex3}

	// Get vertex positions
	var positions [3][3]float32
	for i, vi := range vertexIndices {
		if vi >= len(mesh.Vertices) {
			return
		}
		v := mesh.Vertices[vi]
		// Add center offset and swap Y/Z
		positions[i] = [3]float32{
			v.X + mesh.Center.X,
			v.Z + mesh.Center.Z, // Swap Y and Z
			v.Y + mesh.Center.Y,
		}
	}

	// Get vertex normals
	var normals [3][3]float32
	for i, vi := range vertexIndices {
		if vi >= len(mesh.Normals) {
			normals[i] = [3]float32{0, 1, 0}
			continue
		}
		n := mesh.Normals[vi]
		// Negate and normalize
		normals[i] = [3]float32{-n.X, -n.Z, -n.Y}
	}

	// Get texture coordinates
	var uvs [3][2]float32
	for i, vi := range vertexIndices {
		if vi >= len(mesh.TextureUvCoordinates) {
			uvs[i] = [2]float32{0, 0}
			continue
		}
		uv := mesh.TextureUvCoordinates[vi]
		uvs[i] = [2]float32{uv.X, -uv.Y} // Negate V
	}

	// Get bone indices
	var boneIndices [3]int
	if isSkinned && singularBoneIndex == -1 {
		for i, vi := range vertexIndices {
			boneIndices[i] = getBoneIndexForVertex(mesh, vi)
		}
	} else if singularBoneIndex >= 0 {
		for i := range boneIndices {
			boneIndices[i] = singularBoneIndex
		}
	}

	// Get vertex colors
	var colors [3][4]float32
	if canExportVertexColors {
		for i, vi := range vertexIndices {
			colors[i] = getVertexColor(mesh, vi, objectInstance)
		}
	}

	// Add vertices to builder
	var indices [3]uint32
	for i := 0; i < 3; i++ {
		key := vertexKey{
			posX: positions[i][0], posY: positions[i][1], posZ: positions[i][2],
			normX: normals[i][0], normY: normals[i][1], normZ: normals[i][2],
			u: uvs[i][0], v: uvs[i][1],
		}
		if canExportVertexColors {
			key.colorR = colors[i][0]
			key.colorG = colors[i][1]
			key.colorB = colors[i][2]
			key.colorA = colors[i][3]
		}
		if isSkinned {
			key.joint = uint16(boneIndices[i])
		}

		if existingIdx, ok := builder.vertexMap[key]; ok {
			indices[i] = existingIdx
		} else {
			idx := uint32(len(builder.meshData.positions))
			builder.meshData.positions = append(builder.meshData.positions, positions[i])
			builder.meshData.normals = append(builder.meshData.normals, normals[i])
			builder.meshData.uvs = append(builder.meshData.uvs, uvs[i])

			if canExportVertexColors {
				builder.meshData.colors = append(builder.meshData.colors, colors[i])
			}

			if isSkinned {
				builder.meshData.joints = append(builder.meshData.joints, [4]uint16{uint16(boneIndices[i]), 0, 0, 0})
				builder.meshData.weights = append(builder.meshData.weights, [4]float32{1.0, 0.0, 0.0, 0.0})
			}

			builder.vertexMap[key] = idx
			indices[i] = idx
		}
	}

	// Add indices to primitive
	prim, ok := builder.meshData.primitives[matIdx]
	if !ok {
		prim = &primitiveData{}
		builder.meshData.primitives[matIdx] = prim
	}

	// For skinned meshes, reverse winding order
	if isSkinned {
		prim.indices = append(prim.indices, indices[2], indices[1], indices[0])
	} else {
		prim.indices = append(prim.indices, indices[0], indices[1], indices[2])
	}
}

// AddCombinedMeshToScene adds the combined mesh to the scene.
func (w *GltfWriter) AddCombinedMeshToScene(isZoneMesh bool, meshName string, skeletonModelBase string, objectInstance *fragments.ObjectInstance) {
	var builder *meshBuilder
	if meshName != "" {
		if data, exists := w.sharedMeshes[meshName]; exists {
			builder = &meshBuilder{
				name:     meshName,
				meshData: data,
			}
		}
	}
	if builder == nil {
		builder = w.combinedMesh
	}
	if builder == nil {
		return
	}

	transform := identityMatrix()
	if objectInstance != nil {
		objTransform := createTransformMatrixForObjectInstance(objectInstance)
		transform = multiplyMatrices(objTransform, correctedWorldMatrix())
	} else if isZoneMesh {
		transform = correctedWorldMatrix()
	} else {
		// Mirror Z axis
		transform = [16]float32{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, -1, 0,
			0, 0, 0, 1,
		}
	}

	if skeletonModelBase != "" {
		if skelData, exists := w.skeletons[skeletonModelBase]; exists {
			w.addSkinnedMeshToScene(builder, transform, skelData)
		} else {
			w.addMeshToScene(builder, transform)
		}
	} else {
		w.addMeshToScene(builder, transform)
	}

	if meshName != "" {
		w.sharedMeshes[meshName] = builder.meshData
	}
	w.combinedMesh = nil
}

// addMeshToScene adds a mesh to the scene with transform.
func (w *GltfWriter) addMeshToScene(builder *meshBuilder, transform [16]float32) {
	if builder == nil || len(builder.meshData.positions) == 0 {
		return
	}

	meshIdx := w.buildGltfMesh(builder)

	// Create node
	node := &gltf.Node{
		Name: builder.name,
		Mesh: gltf.Index(meshIdx),
	}

	// Apply transform if not identity
	if !isIdentityMatrix(transform) {
		node.Matrix = transform
	}

	nodeIdx := uint32(len(w.doc.Nodes))
	w.doc.Nodes = append(w.doc.Nodes, node)
	w.doc.Scenes[0].Nodes = append(w.doc.Scenes[0].Nodes, nodeIdx)
}

// addSkinnedMeshToScene adds a skinned mesh to the scene.
func (w *GltfWriter) addSkinnedMeshToScene(builder *meshBuilder, transform [16]float32, skelData *skeletonData) {
	if builder == nil || len(builder.meshData.positions) == 0 {
		return
	}

	meshIdx := w.buildGltfMesh(builder)

	// Create skin
	skinIdx := uint32(len(w.doc.Skins))
	skin := &gltf.Skin{
		Name:   builder.name + "_skin",
		Joints: skelData.nodes,
	}

	// Add inverse bind matrices accessor
	if len(skelData.inverseBindMatrices) > 0 {
		ibmAccessor := modeler.WriteAccessor(w.doc, gltf.TargetNone, skelData.inverseBindMatrices)
		w.doc.Accessors[ibmAccessor].Type = gltf.AccessorMat4
		skin.InverseBindMatrices = gltf.Index(ibmAccessor)
	}

	w.doc.Skins = append(w.doc.Skins, skin)

	// Create node
	node := &gltf.Node{
		Name: builder.name,
		Mesh: gltf.Index(meshIdx),
		Skin: gltf.Index(skinIdx),
	}

	if !isIdentityMatrix(transform) {
		node.Matrix = transform
	}

	nodeIdx := uint32(len(w.doc.Nodes))
	w.doc.Nodes = append(w.doc.Nodes, node)
	w.doc.Scenes[0].Nodes = append(w.doc.Scenes[0].Nodes, nodeIdx)
}

// buildGltfMesh builds a glTF mesh from the builder.
func (w *GltfWriter) buildGltfMesh(builder *meshBuilder) uint32 {
	mesh := &gltf.Mesh{
		Name: builder.name,
	}

	// Create accessors for vertex attributes
	posAccessor := modeler.WritePosition(w.doc, builder.meshData.positions)
	normAccessor := modeler.WriteNormal(w.doc, builder.meshData.normals)
	uvAccessor := modeler.WriteTextureCoord(w.doc, builder.meshData.uvs)

	var colorAccessor, jointAccessor, weightAccessor uint32
	hasColors := len(builder.meshData.colors) > 0
	hasSkinning := len(builder.meshData.joints) > 0

	if hasColors {
		colorAccessor = modeler.WriteColor(w.doc, builder.meshData.colors)
	}

	if hasSkinning {
		jointAccessor = modeler.WriteJoints(w.doc, builder.meshData.joints)
		weightAccessor = modeler.WriteWeights(w.doc, builder.meshData.weights)
	}

	// Create primitives for each material
	for matIdx, primData := range builder.meshData.primitives {
		if len(primData.indices) == 0 {
			continue
		}

		indicesAccessor := modeler.WriteIndices(w.doc, primData.indices)

		prim := &gltf.Primitive{
			Attributes: map[string]uint32{
				gltf.POSITION:   posAccessor,
				gltf.NORMAL:     normAccessor,
				gltf.TEXCOORD_0: uvAccessor,
			},
			Indices:  gltf.Index(indicesAccessor),
			Material: gltf.Index(matIdx),
			Mode:     gltf.PrimitiveTriangles,
		}

		if hasColors {
			prim.Attributes[gltf.COLOR_0] = colorAccessor
		}

		if hasSkinning {
			prim.Attributes[gltf.JOINTS_0] = jointAccessor
			prim.Attributes[gltf.WEIGHTS_0] = weightAccessor
		}

		mesh.Primitives = append(mesh.Primitives, prim)
	}

	meshIdx := uint32(len(w.doc.Meshes))
	w.doc.Meshes = append(w.doc.Meshes, mesh)
	return meshIdx
}

// ApplyAnimationToSkeleton applies animation data to skeleton nodes.
func (w *GltfWriter) ApplyAnimationToSkeleton(skeleton *fragments.SkeletonHierarchy, animationKey string, isCharacterAnimation bool, staticPose bool) {
	if isCharacterAnimation && !staticPose && animationKey == defaultModelPoseAnimKey {
		return
	}

	skelData, exists := w.skeletons[skeleton.ModelBase]
	if !exists {
		skelData = w.addNewSkeleton(skeleton)
	}

	animation, ok := skeleton.Animations[animationKey]
	if !ok {
		return
	}

	var trackArray, poseArray map[string]datatypes.TrackFragment
	if isCharacterAnimation {
		trackArray = animation.TracksCleanedStripped
		if poseAnim, ok := skeleton.Animations[defaultModelPoseAnimKey]; ok {
			poseArray = poseAnim.TracksCleanedStripped
		}
	} else {
		trackArray = animation.TracksCleaned
		if poseAnim, ok := skeleton.Animations[defaultModelPoseAnimKey]; ok {
			poseArray = poseAnim.TracksCleaned
		}
	}

	if poseArray == nil {
		return
	}

	// Create animation if not static pose
	var gltfAnim *gltf.Animation
	if !staticPose {
		gltfAnim = &gltf.Animation{
			Name: animationKey,
		}
	}

	for i := 0; i < len(skeleton.Skeleton); i++ {
		var boneName string
		if isCharacterAnimation {
			boneName = datatypes.CleanBoneAndStripBase(skeleton.BoneMapping[i], skeleton.ModelBase)
		} else {
			boneName = datatypes.CleanBoneName(skeleton.BoneMapping[i])
		}

		nodeIdx := skelData.nodes[i]

		if staticPose || trackArray[boneName] == nil {
			poseTrack := poseArray[boneName]
			if poseTrack == nil {
				continue
			}
			trackDef := poseTrack.GetTrackDefFragment()
			if trackDef == nil {
				continue
			}
			frames := getTrackDefFrames(trackDef)
			if len(frames) == 0 {
				continue
			}

			w.applyBoneTransformation(nodeIdx, &frames[0], staticPose)
			continue
		}

		track := trackArray[boneName]
		if track == nil {
			continue
		}
		trackDef := track.GetTrackDefFragment()
		if trackDef == nil {
			continue
		}
		frames := getTrackDefFrames(trackDef)

		totalTimeForBone := 0
		for frame := 0; frame < animation.FrameCount; frame++ {
			if frame >= len(frames) {
				break
			}

			boneTransform := &frames[frame]

			if !staticPose && gltfAnim != nil {
				w.addAnimationKeyframe(gltfAnim, nodeIdx, boneTransform, animationKey, float32(totalTimeForBone)/1000.0)
				if frame == 0 && loopedAnimationKeys[animationKey] {
					w.addAnimationKeyframe(gltfAnim, nodeIdx, boneTransform, animationKey, float32(animation.AnimationTimeMs)/1000.0)
				}
			} else {
				w.applyBoneTransformation(nodeIdx, boneTransform, staticPose)
			}

			if isCharacterAnimation {
				totalTimeForBone += animation.AnimationTimeMs / animation.FrameCount
			} else {
				totalTimeForBone += skeleton.Skeleton[i].Track.GetFrameMs()
			}
		}
	}

	if !staticPose && gltfAnim != nil && len(gltfAnim.Channels) > 0 {
		w.doc.Animations = append(w.doc.Animations, gltfAnim)
	}
}

// applyBoneTransformation applies a bone transformation to a node.
func (w *GltfWriter) applyBoneTransformation(nodeIdx uint32, transform *datatypes.BoneTransform, staticPose bool) {
	if nodeIdx >= uint32(len(w.doc.Nodes)) {
		return
	}

	node := w.doc.Nodes[nodeIdx]

	// Convert rotation
	rotX := transform.Rotation.X * math.Pi / 180.0
	rotY := transform.Rotation.Z * math.Pi / 180.0 // Swap Y and Z
	rotZ := -transform.Rotation.Y * math.Pi / 180.0
	rotW := transform.Rotation.W * math.Pi / 180.0

	// Normalize quaternion
	length := float32(math.Sqrt(float64(rotX*rotX + rotY*rotY + rotZ*rotZ + rotW*rotW)))
	if length > 0 {
		rotX /= length
		rotY /= length
		rotZ /= length
		rotW /= length
	}

	// Convert translation (swap Y and Z, negate Z)
	transX := transform.Translation.X
	transY := transform.Translation.Z
	transZ := -transform.Translation.Y

	scale := transform.Scale
	if scale == 0 {
		scale = 1.0
	}

	if staticPose {
		node.Scale = [3]float32{scale, scale, scale}
		node.Rotation = [4]float32{rotX, rotY, rotZ, rotW}
		node.Translation = [3]float32{transX, transY, transZ}
	}
}

// addAnimationKeyframe adds a keyframe to an animation.
func (w *GltfWriter) addAnimationKeyframe(anim *gltf.Animation, nodeIdx uint32, transform *datatypes.BoneTransform, animKey string, time float32) {
	// Convert rotation
	rotX := transform.Rotation.X * math.Pi / 180.0
	rotY := transform.Rotation.Z * math.Pi / 180.0
	rotZ := -transform.Rotation.Y * math.Pi / 180.0
	rotW := transform.Rotation.W * math.Pi / 180.0

	// Normalize quaternion
	length := float32(math.Sqrt(float64(rotX*rotX + rotY*rotY + rotZ*rotZ + rotW*rotW)))
	if length > 0 {
		rotX /= length
		rotY /= length
		rotZ /= length
		rotW /= length
	}

	// Convert translation
	transX := transform.Translation.X
	transY := transform.Translation.Z
	transZ := -transform.Translation.Y

	scale := transform.Scale
	if scale == 0 {
		scale = 1.0
	}

	// Add scale sampler
	scaleInputAccessor := modeler.WriteAccessor(w.doc, gltf.TargetNone, []float32{time})
	scaleOutputAccessor := modeler.WriteAccessor(w.doc, gltf.TargetNone, []float32{scale, scale, scale})
	w.doc.Accessors[scaleOutputAccessor].Type = gltf.AccessorVec3

	scaleSamplerIdx := uint32(len(anim.Samplers))
	anim.Samplers = append(anim.Samplers, &gltf.AnimationSampler{
		Input:         scaleInputAccessor,
		Output:        scaleOutputAccessor,
		Interpolation: gltf.InterpolationLinear,
	})
	anim.Channels = append(anim.Channels, &gltf.Channel{
		Sampler: gltf.Index(scaleSamplerIdx),
		Target: gltf.ChannelTarget{
			Node: gltf.Index(nodeIdx),
			Path: gltf.TRSScale,
		},
	})

	// Add rotation sampler
	rotInputAccessor := modeler.WriteAccessor(w.doc, gltf.TargetNone, []float32{time})
	rotOutputAccessor := modeler.WriteAccessor(w.doc, gltf.TargetNone, []float32{rotX, rotY, rotZ, rotW})
	w.doc.Accessors[rotOutputAccessor].Type = gltf.AccessorVec4

	rotSamplerIdx := uint32(len(anim.Samplers))
	anim.Samplers = append(anim.Samplers, &gltf.AnimationSampler{
		Input:         rotInputAccessor,
		Output:        rotOutputAccessor,
		Interpolation: gltf.InterpolationLinear,
	})
	anim.Channels = append(anim.Channels, &gltf.Channel{
		Sampler: gltf.Index(rotSamplerIdx),
		Target: gltf.ChannelTarget{
			Node: gltf.Index(nodeIdx),
			Path: gltf.TRSRotation,
		},
	})

	// Add translation sampler
	transInputAccessor := modeler.WriteAccessor(w.doc, gltf.TargetNone, []float32{time})
	transOutputAccessor := modeler.WriteAccessor(w.doc, gltf.TargetNone, []float32{transX, transY, transZ})
	w.doc.Accessors[transOutputAccessor].Type = gltf.AccessorVec3

	transSamplerIdx := uint32(len(anim.Samplers))
	anim.Samplers = append(anim.Samplers, &gltf.AnimationSampler{
		Input:         transInputAccessor,
		Output:        transOutputAccessor,
		Interpolation: gltf.InterpolationLinear,
	})
	anim.Channels = append(anim.Channels, &gltf.Channel{
		Sampler: gltf.Index(transSamplerIdx),
		Target: gltf.ChannelTarget{
			Node: gltf.Index(nodeIdx),
			Path: gltf.TRSTranslation,
		},
	})
}

// addNewSkeleton adds a new skeleton to the writer.
func (w *GltfWriter) addNewSkeleton(skeleton *fragments.SkeletonHierarchy) *skeletonData {
	skelData := &skeletonData{
		nodes:     make([]uint32, len(skeleton.Skeleton)),
		boneNames: make([]string, len(skeleton.Skeleton)),
	}

	duplicateNameCount := make(map[string]int)

	// Create nodes for each bone
	for i, bone := range skeleton.Skeleton {
		boneName := bone.CleanedName
		if count, exists := duplicateNameCount[boneName]; exists {
			skelData.boneNames[i] = fmt.Sprintf("%s_%02d", boneName, count)
			duplicateNameCount[boneName]++
		} else {
			skelData.boneNames[i] = boneName
			duplicateNameCount[boneName] = 1
		}

		node := &gltf.Node{
			Name: skelData.boneNames[i],
		}

		nodeIdx := uint32(len(w.doc.Nodes))
		w.doc.Nodes = append(w.doc.Nodes, node)
		skelData.nodes[i] = nodeIdx
	}

	// Set up parent-child relationships
	for i, bone := range skeleton.Skeleton {
		nodeIdx := skelData.nodes[i]
		node := w.doc.Nodes[nodeIdx]

		for _, childIdx := range bone.Children {
			if childIdx < len(skelData.nodes) {
				childNodeIdx := skelData.nodes[childIdx]
				node.Children = append(node.Children, childNodeIdx)
			}
		}
	}

	// Add root node to scene
	if len(skelData.nodes) > 0 {
		w.doc.Scenes[0].Nodes = append(w.doc.Scenes[0].Nodes, skelData.nodes[0])
	}

	// Generate inverse bind matrices (identity for now)
	skelData.inverseBindMatrices = make([]float32, len(skeleton.Skeleton)*16)
	for i := 0; i < len(skeleton.Skeleton); i++ {
		identity := identityMatrix()
		copy(skelData.inverseBindMatrices[i*16:(i+1)*16], identity[:])
	}

	w.skeletons[skeleton.ModelBase] = skelData
	return skelData
}

// WriteAssetToFile writes the glTF document to a file.
func (w *GltfWriter) WriteAssetToFile(fileName string, useExistingImages bool, skeletonModelBase string) error {
	w.AddCombinedMeshToScene(false, "", skeletonModelBase, nil)

	outputPath := w.fixFilePath(fileName)

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if w.exportFormat == GltfExportFormatGlb {
		return gltf.SaveBinary(w.doc, outputPath)
	}

	return gltf.Save(w.doc, outputPath)
}

// ClearExportData clears all export data for reuse.
func (w *GltfWriter) ClearExportData() {
	w.doc = gltf.NewDocument()
	w.doc.Asset.Generator = "LanternGoExtract"
	w.doc.Scenes = append(w.doc.Scenes, &gltf.Scene{Name: "Scene"})
	w.doc.Scene = gltf.Index(0)

	w.Materials = make(map[string]uint32)
	w.sharedMeshes = make(map[string]*meshData)
	w.skeletons = make(map[string]*skeletonData)
	w.meshMaterialsToSkip = make(map[string]bool)
	w.textureIndices = make(map[string]uint32)
	w.combinedMesh = nil
}

// fixFilePath ensures the file has the correct extension.
func (w *GltfWriter) fixFilePath(filePath string) string {
	ext := ".gltf"
	if w.exportFormat == GltfExportFormatGlb {
		ext = ".glb"
	}

	// Remove existing extension
	base := strings.TrimSuffix(filePath, filepath.Ext(filePath))
	return base + ext
}

// Helper functions

func newMeshBuilder(name string, isSkinned, hasColors bool) *meshBuilder {
	return &meshBuilder{
		name:      name,
		isSkinned: isSkinned,
		hasColors: hasColors,
		meshData: &meshData{
			primitives: make(map[uint32]*primitiveData),
		},
		vertexMap: make(map[vertexKey]uint32),
	}
}

func getMaterialName(material *fragments.Material) string {
	prefix := fragments.GetMaterialPrefix(material.ShaderType)
	bitmapName := gltfGetBitmapNameWithoutExtension(material)
	return prefix + bitmapName
}

func gltfGetBitmapNameWithoutExtension(material *fragments.Material) string {
	if material.BitmapInfoReference == nil {
		return ""
	}

	bitmapInfoRef, ok := material.BitmapInfoReference.(*fragments.BitmapInfoReference)
	if !ok || bitmapInfoRef.BitmapInfo == nil {
		return ""
	}

	if len(bitmapInfoRef.BitmapInfo.BitmapNames) == 0 {
		return ""
	}

	return bitmapInfoRef.BitmapInfo.BitmapNames[0].GetFilenameWithoutExtension()
}

func gltfGetBitmapExportFilename(material *fragments.Material) string {
	if material.BitmapInfoReference == nil {
		return ""
	}

	bitmapInfoRef, ok := material.BitmapInfoReference.(*fragments.BitmapInfoReference)
	if !ok || bitmapInfoRef.BitmapInfo == nil {
		return ""
	}

	if len(bitmapInfoRef.BitmapInfo.BitmapNames) == 0 {
		return ""
	}

	return bitmapInfoRef.BitmapInfo.BitmapNames[0].GetExportFilename()
}

func getBoneIndexForVertex(mesh *fragments.Mesh, vertexIndex int) int {
	for boneIdx, piece := range mesh.MobPieces {
		if vertexIndex >= piece.Start && vertexIndex < piece.Start+piece.Count {
			return boneIdx
		}
	}
	return 0
}

func getVertexColor(mesh *fragments.Mesh, vertexIndex int, objectInstance *fragments.ObjectInstance) [4]float32 {
	// Try object instance colors first
	if objectInstance != nil && objectInstance.Colors != nil {
		if vc, ok := objectInstance.Colors.(*fragments.VertexColors); ok {
			if vertexIndex < len(vc.Colors) {
				c := vc.Colors[vertexIndex]
				return [4]float32{
					float32(c.R) / 255.0,
					float32(c.G) / 255.0,
					float32(c.B) / 255.0,
					float32(c.A) / 255.0,
				}
			}
		}
	}

	// Fall back to mesh colors
	if vertexIndex < len(mesh.Colors) {
		c := mesh.Colors[vertexIndex]
		return [4]float32{
			float32(c.R) / 255.0,
			float32(c.G) / 255.0,
			float32(c.B) / 255.0,
			float32(c.A) / 255.0,
		}
	}

	return defaultVertexColor
}

func hasObjectColors(objectInstance *fragments.ObjectInstance) bool {
	if objectInstance == nil || objectInstance.Colors == nil {
		return false
	}
	if vc, ok := objectInstance.Colors.(*fragments.VertexColors); ok {
		return len(vc.Colors) > 0
	}
	return false
}

func getTrackDefFrames(trackDef interface{}) []datatypes.BoneTransform {
	if td, ok := trackDef.(*fragments.TrackDefFragment); ok {
		return td.Frames
	}
	return nil
}

// Matrix helper functions

func identityMatrix() [16]float32 {
	return [16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func mirrorXAxisMatrix() [16]float32 {
	return [16]float32{
		-1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func correctedWorldMatrix() [16]float32 {
	// Mirror X axis * scale 0.1
	return [16]float32{
		-0.1, 0, 0, 0,
		0, 0.1, 0, 0,
		0, 0, 0.1, 0,
		0, 0, 0, 1,
	}
}

func isIdentityMatrix(m [16]float32) bool {
	identity := identityMatrix()
	for i := 0; i < 16; i++ {
		if m[i] != identity[i] {
			return false
		}
	}
	return true
}

func multiplyMatrices(a, b [16]float32) [16]float32 {
	var result [16]float32
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				result[i*4+j] += a[i*4+k] * b[k*4+j]
			}
		}
	}
	return result
}

func createTransformMatrixForObjectInstance(instance *fragments.ObjectInstance) [16]float32 {
	// Create scale matrix
	sx := instance.Scale.X
	sy := instance.Scale.Y
	sz := instance.Scale.Z

	// Create rotation angles (convert to radians)
	rx := float64(instance.Rotation.X) * math.Pi / 180.0
	ry := float64(instance.Rotation.Y) * math.Pi / 180.0
	rz := float64(instance.Rotation.Z) * math.Pi / 180.0

	// Compute rotation matrix (YawPitchRoll: Z, X, Y order)
	cosX, sinX := math.Cos(rx), math.Sin(rx)
	cosY, sinY := math.Cos(ry), math.Sin(ry)
	cosZ, sinZ := math.Cos(rz), math.Sin(rz)

	// Combined rotation matrix
	r00 := float32(cosZ*cosY - sinZ*sinX*sinY)
	r01 := float32(-sinZ * cosX)
	r02 := float32(cosZ*sinY + sinZ*sinX*cosY)
	r10 := float32(sinZ*cosY + cosZ*sinX*sinY)
	r11 := float32(cosZ * cosX)
	r12 := float32(sinZ*sinY - cosZ*sinX*cosY)
	r20 := float32(-cosX * sinY)
	r21 := float32(sinX)
	r22 := float32(cosX * cosY)

	// Translation (swap Y and Z for glTF coordinate system)
	tx := instance.Position.X
	ty := instance.Position.Z
	tz := instance.Position.Y

	// Combine: T * R * S
	return [16]float32{
		r00 * sx, r01 * sx, r02 * sx, 0,
		r10 * sy, r11 * sy, r12 * sy, 0,
		r20 * sz, r21 * sz, r22 * sz, 0,
		tx, ty, tz, 1,
	}
}
