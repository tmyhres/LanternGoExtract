package helpers

import (
	"github.com/lanterneq/lanern-go/pkg/wld/datatypes"
	"github.com/lanterneq/lanern-go/pkg/wld/fragments"
)

// MeshExportHelper provides utilities for mesh export operations,
// particularly for transforming mesh vertices for skeletal animations.

// ShiftMeshVertices transforms vertices of a mesh for the given animation and frame.
// It returns the original vertex positions before transformation.
//
// Parameters:
//   - mesh: The mesh that will have vertices shifted
//   - skeleton: The SkeletonHierarchy that contains the bone transformations
//   - isCharacterAnimation: If true, uses stripped track names for lookup
//   - animName: The name of the animation used for the transform
//   - frame: The frame of the animation
//   - singularBoneIndex: The bone index for the mesh when there is a 1:1 relationship (-1 if not applicable)
//
// Returns the original vertex positions before transformation.
func ShiftMeshVertices(mesh *fragments.Mesh, skeleton *fragments.SkeletonHierarchy,
	isCharacterAnimation bool, animName string, frame int, singularBoneIndex int) []fragments.Vec3 {

	originalVertices := make([]fragments.Vec3, 0)

	if skeleton == nil || mesh == nil {
		return originalVertices
	}

	// Check if animation exists
	animation, exists := skeleton.Animations[animName]
	if !exists || len(mesh.Vertices) == 0 {
		return originalVertices
	}

	// Select the appropriate track map based on animation type
	var tracks map[string]datatypes.TrackFragment
	if isCharacterAnimation {
		tracks = animation.TracksCleanedStripped
	} else {
		tracks = animation.TracksCleaned
	}

	// Handle singular bone case (1:1 relationship between mesh and bone)
	if singularBoneIndex > -1 {
		if singularBoneIndex >= len(skeleton.Skeleton) {
			return originalVertices
		}

		bone := skeleton.Skeleton[singularBoneIndex].CleanedName
		if _, exists := tracks[bone]; !exists {
			return originalVertices
		}

		modelMatrix := GetBoneMatrix(skeleton, singularBoneIndex, animName, frame)
		originalVertices = append(originalVertices,
			shiftMeshVerticesWithIndices(0, len(mesh.Vertices)-1, mesh, modelMatrix)...)

		return originalVertices
	}

	// Handle mesh with multiple bone influences (MobPieces)
	for boneIndex, mobVertexPiece := range mesh.MobPieces {
		if boneIndex >= len(skeleton.Skeleton) {
			continue
		}

		bone := skeleton.Skeleton[boneIndex].CleanedName
		if _, exists := tracks[bone]; !exists {
			continue
		}

		modelMatrix := GetBoneMatrix(skeleton, boneIndex, animName, frame)

		start := mobVertexPiece.Start
		end := start + mobVertexPiece.Count - 1

		originalVertices = append(originalVertices,
			shiftMeshVerticesWithIndices(start, end, mesh, modelMatrix)...)
	}

	return originalVertices
}

// shiftMeshVerticesWithIndices transforms vertices in a given range using a bone matrix.
// It returns the original vertex positions and modifies the mesh vertices in place.
func shiftMeshVerticesWithIndices(start, end int, mesh *fragments.Mesh, boneMatrix datatypes.Mat4) []fragments.Vec3 {
	originalVertices := make([]fragments.Vec3, 0, end-start+1)

	for i := start; i <= end; i++ {
		if i >= len(mesh.Vertices) {
			break
		}

		vertex := mesh.Vertices[i]
		originalVertices = append(originalVertices, vertex)

		// Transform vertex by bone matrix: newVertex = boneMatrix * vec4(vertex, 1.0)
		newVertex := transformVec3ByMat4(vertex, boneMatrix)
		mesh.Vertices[i] = newVertex
	}

	return originalVertices
}

// transformVec3ByMat4 multiplies a 3D vector by a 4x4 matrix.
// The vector is treated as a 4D vector with w=1.0 for the multiplication.
func transformVec3ByMat4(v fragments.Vec3, m datatypes.Mat4) fragments.Vec3 {
	// Mat4 is stored as [16]float32 in column-major order:
	// [0]  [4]  [8]  [12]
	// [1]  [5]  [9]  [13]
	// [2]  [6]  [10] [14]
	// [3]  [7]  [11] [15]
	x := m[0]*v.X + m[4]*v.Y + m[8]*v.Z + m[12]
	y := m[1]*v.X + m[5]*v.Y + m[9]*v.Z + m[13]
	z := m[2]*v.X + m[6]*v.Y + m[10]*v.Z + m[14]
	w := m[3]*v.X + m[7]*v.Y + m[11]*v.Z + m[15]

	// Perspective divide (normally w should be 1.0 for affine transforms)
	if w != 0 && w != 1 {
		x /= w
		y /= w
		z /= w
	}

	return fragments.Vec3{X: x, Y: y, Z: z}
}

// GetBoneMatrix retrieves the model matrix for a bone at a specific animation frame.
// This is a placeholder implementation - the actual implementation should be in
// the SkeletonHierarchy type or use the skeleton's internal matrix calculation.
//
// Note: This function needs to be implemented based on the skeleton's bone hierarchy
// and animation system. The implementation should:
// 1. Get the bone transform from the animation at the specified frame
// 2. Calculate the world matrix by multiplying parent transforms
// 3. Return the final model matrix
func GetBoneMatrix(skeleton *fragments.SkeletonHierarchy, boneIndex int, animName string, frame int) datatypes.Mat4 {
	// Return identity matrix as placeholder
	// The actual implementation should traverse the bone hierarchy and calculate
	// the cumulative transform matrix
	return datatypes.Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// RestoreVertices restores mesh vertices to their original positions.
// This can be used after exporting a frame to reset the mesh for the next frame.
func RestoreVertices(mesh *fragments.Mesh, originalVertices []fragments.Vec3, startIndex int) {
	for i, vertex := range originalVertices {
		idx := startIndex + i
		if idx < len(mesh.Vertices) {
			mesh.Vertices[idx] = vertex
		}
	}
}

// Mat4Identity returns a 4x4 identity matrix.
func Mat4Identity() datatypes.Mat4 {
	return datatypes.Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// Mat4Multiply multiplies two 4x4 matrices.
func Mat4Multiply(a, b datatypes.Mat4) datatypes.Mat4 {
	var result datatypes.Mat4

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			result[i+j*4] = a[i+0*4]*b[0+j*4] +
				a[i+1*4]*b[1+j*4] +
				a[i+2*4]*b[2+j*4] +
				a[i+3*4]*b[3+j*4]
		}
	}

	return result
}

// Mat4FromTranslation creates a translation matrix.
func Mat4FromTranslation(v datatypes.Vec3) datatypes.Mat4 {
	return datatypes.Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		v.X, v.Y, v.Z, 1,
	}
}

// Mat4FromQuaternion creates a rotation matrix from a quaternion.
func Mat4FromQuaternion(q datatypes.Quat) datatypes.Mat4 {
	x2 := q.X + q.X
	y2 := q.Y + q.Y
	z2 := q.Z + q.Z

	xx := q.X * x2
	xy := q.X * y2
	xz := q.X * z2
	yy := q.Y * y2
	yz := q.Y * z2
	zz := q.Z * z2
	wx := q.W * x2
	wy := q.W * y2
	wz := q.W * z2

	return datatypes.Mat4{
		1 - (yy + zz), xy + wz, xz - wy, 0,
		xy - wz, 1 - (xx + zz), yz + wx, 0,
		xz + wy, yz - wx, 1 - (xx + yy), 0,
		0, 0, 0, 1,
	}
}

// Mat4FromScale creates a scale matrix.
func Mat4FromScale(scale float32) datatypes.Mat4 {
	return datatypes.Mat4{
		scale, 0, 0, 0,
		0, scale, 0, 0,
		0, 0, scale, 0,
		0, 0, 0, 1,
	}
}

// Mat4FromTRS creates a transformation matrix from translation, rotation, and scale.
func Mat4FromTRS(translation datatypes.Vec3, rotation datatypes.Quat, scale float32) datatypes.Mat4 {
	// Scale first, then rotate, then translate
	scaleMatrix := Mat4FromScale(scale)
	rotMatrix := Mat4FromQuaternion(rotation)
	transMatrix := Mat4FromTranslation(translation)

	// Combine: T * R * S
	return Mat4Multiply(transMatrix, Mat4Multiply(rotMatrix, scaleMatrix))
}
