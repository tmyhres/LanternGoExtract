// Package datatypes contains data structures for WLD file parsing.
package datatypes

// Vec2 represents a 2D vector.
type Vec2 struct {
	X float32
	Y float32
}

// Vec3 represents a 3D vector.
type Vec3 struct {
	X float32
	Y float32
	Z float32
}

// Vec4 represents a 4D vector.
type Vec4 struct {
	X float32
	Y float32
	Z float32
	W float32
}

// Quat represents a quaternion for rotations.
type Quat struct {
	X float32
	Y float32
	Z float32
	W float32
}

// Mat4 represents a 4x4 matrix.
type Mat4 [16]float32
