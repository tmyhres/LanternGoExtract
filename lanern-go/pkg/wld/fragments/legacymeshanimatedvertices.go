package fragments

import (
	"fmt"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/datatypes"
)

// LegacyMeshAnimatedVertices (0x2E)
// Internal name: _DMTRACKDEF
// Contains a list of frames each containing a position for each vertex.
// The frame vertices are cycled through, animating the model.
type LegacyMeshAnimatedVertices struct {
	BaseFragment
	// Frames contains the model frames
	Frames [][]datatypes.Vec3
	// Delay is the delay between the vertex swaps
	Delay int
}

// FragmentType returns the fragment type ID.
func (f *LegacyMeshAnimatedVertices) FragmentType() uint32 {
	return 0x2E
}

// Initialize parses the fragment data.
func (f *LegacyMeshAnimatedVertices) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)
	reader := NewFragmentReader(data)

	nameRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	_, err = reader.ReadInt32() // flags
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	vertexCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read vertex count: %w", err)
	}

	frameCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read frame count: %w", err)
	}

	delay, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read delay: %w", err)
	}
	f.Delay = int(delay)

	_, err = reader.ReadInt32() // param1
	if err != nil {
		return fmt.Errorf("failed to read param1: %w", err)
	}

	f.Frames = make([][]datatypes.Vec3, frameCount)
	for i := int32(0); i < frameCount; i++ {
		positions := make([]datatypes.Vec3, vertexCount)
		for v := int32(0); v < vertexCount; v++ {
			x, err := reader.ReadFloat32()
			if err != nil {
				return fmt.Errorf("failed to read position x: %w", err)
			}
			y, err := reader.ReadFloat32()
			if err != nil {
				return fmt.Errorf("failed to read position y: %w", err)
			}
			z, err := reader.ReadFloat32()
			if err != nil {
				return fmt.Errorf("failed to read position z: %w", err)
			}
			positions[v] = datatypes.Vec3{X: x, Y: y, Z: z}
		}
		f.Frames[i] = positions
	}

	return nil
}

// GetDelay returns the animation delay.
func (f *LegacyMeshAnimatedVertices) GetDelay() int {
	return f.Delay
}

// GetFrames returns the animation frames.
func (f *LegacyMeshAnimatedVertices) GetFrames() [][]Vec3 {
	result := make([][]Vec3, len(f.Frames))
	for i, frame := range f.Frames {
		result[i] = make([]Vec3, len(frame))
		for j, v := range frame {
			result[i][j] = Vec3{X: v.X, Y: v.Y, Z: v.Z}
		}
	}
	return result
}
