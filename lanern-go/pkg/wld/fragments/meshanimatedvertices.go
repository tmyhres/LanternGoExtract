package fragments

// MeshAnimatedVertices (0x37) contains a list of frames each containing a position for each vertex.
// Internal name: _DMTRACKDEF
// The frame vertices are cycled through, animating the model.
type MeshAnimatedVertices struct {
	BaseFragment

	// Frames contains the model animation frames.
	// Each frame is a list of vertex positions.
	frames [][]Vec3

	// Delay is the delay between vertex swaps in milliseconds.
	delay int
}

// FragmentType returns the fragment type ID (0x37).
func (m *MeshAnimatedVertices) FragmentType() uint32 {
	return 0x37
}

// Initialize parses the mesh animated vertices fragment data.
func (m *MeshAnimatedVertices) Initialize(index int, id int, size int, data []byte, fragments []Fragment,
	stringHash map[int]string, isNewFormat bool) error {
	m.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return err
	}
	m.Name = GetStringFromHash(stringHash, nameRef)

	// Read flags
	_, err = r.ReadInt32()
	if err != nil {
		return err
	}

	// Read vertex count
	vertexCount, err := r.ReadInt16()
	if err != nil {
		return err
	}

	// Read frame count
	frameCount, err := r.ReadInt16()
	if err != nil {
		return err
	}

	// Read delay
	delayValue, err := r.ReadInt16()
	if err != nil {
		return err
	}
	m.delay = int(delayValue)

	// Read param2 (unused)
	_, err = r.ReadInt16()
	if err != nil {
		return err
	}

	// Read scale
	scaleValue, err := r.ReadInt16()
	if err != nil {
		return err
	}
	scale := float32(1.0) / float32(int(1)<<scaleValue)

	// Read frames
	m.frames = make([][]Vec3, 0, frameCount)
	for i := int16(0); i < frameCount; i++ {
		positions := make([]Vec3, 0, vertexCount)

		for j := int16(0); j < vertexCount; j++ {
			x, err := r.ReadInt16()
			if err != nil {
				return err
			}
			y, err := r.ReadInt16()
			if err != nil {
				return err
			}
			z, err := r.ReadInt16()
			if err != nil {
				return err
			}

			positions = append(positions, Vec3{
				X: float32(x) * scale,
				Y: float32(y) * scale,
				Z: float32(z) * scale,
			})
		}

		m.frames = append(m.frames, positions)
	}

	return nil
}

// GetFrames returns the animation frames.
func (m *MeshAnimatedVertices) GetFrames() [][]Vec3 {
	return m.frames
}

// SetFrames sets the animation frames.
func (m *MeshAnimatedVertices) SetFrames(frames [][]Vec3) {
	m.frames = frames
}

// GetDelay returns the delay between vertex swaps.
func (m *MeshAnimatedVertices) GetDelay() int {
	return m.delay
}

// SetDelay sets the delay between vertex swaps.
func (m *MeshAnimatedVertices) SetDelay(delay int) {
	m.delay = delay
}
