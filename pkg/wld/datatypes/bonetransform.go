package datatypes

// BoneTransform represents a transformation applied to a bone.
type BoneTransform struct {
	Translation Vec3
	Rotation    Quat
	Scale       float32
	ModelMatrix Mat4
}
