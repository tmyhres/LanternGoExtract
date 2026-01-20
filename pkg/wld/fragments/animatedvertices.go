package fragments

// AnimatedVertices is the interface for fragments that contain animated vertex data.
// Implemented by MeshAnimatedVertices (0x37) and LegacyMeshAnimatedVertices.
// This extends the IAnimatedVertices interface defined in fragment.go.
type AnimatedVertices interface {
	IAnimatedVertices

	// SetFrames sets the animation frames.
	SetFrames(frames [][]Vec3)

	// SetDelay sets the delay between vertex swaps.
	SetDelay(delay int)
}
