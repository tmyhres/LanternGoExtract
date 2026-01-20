package fragments

import (
	"fmt"

	"github.com/tmyhres/LanternGoExtract/pkg/wld/datatypes"
)

// Polyhedron (0x17)
// Internal Name: _POLYHDEF
// POLYHEDRONDEFINITION
// BOUNDINGRADIUS %f
// SCALEFACTOR %f
// NUMVERTICES %d
// XYZ %f %f %f
// NUMFACES %d
// FACE
//
//	NORMALABCD %f %f %f %f
//	NUMVERTICES %d
//	VERTEXLIST %d ...%d
//
// ENDFACE
// ENDPOLYHEDRONDEFINITION
type Polyhedron struct {
	BaseFragment
	BoundingRadius float32
	ScaleFactor    float32
	Vertices       []datatypes.Vec3
	Faces          []*datatypes.Polygon
}

// FragmentType returns the fragment type ID.
func (f *Polyhedron) FragmentType() uint32 {
	return 0x17
}

// Initialize parses the fragment data.
func (f *Polyhedron) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)
	reader := NewFragmentReader(data)

	nameRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	flags, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	hasScaleFactor := (flags & (1 << 0)) != 0
	hasNormalAbcd := (flags & (1 << 1)) != 0

	vertexCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read vertex count: %w", err)
	}

	faceCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read face count: %w", err)
	}

	f.BoundingRadius, err = reader.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read bounding radius: %w", err)
	}

	if hasScaleFactor {
		f.ScaleFactor, err = reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read scale factor: %w", err)
		}
	} else {
		f.ScaleFactor = 1.0
	}

	f.Vertices = make([]datatypes.Vec3, vertexCount)
	for i := int32(0); i < vertexCount; i++ {
		x, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read vertex x: %w", err)
		}
		y, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read vertex y: %w", err)
		}
		z, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read vertex z: %w", err)
		}
		f.Vertices[i] = datatypes.Vec3{X: x, Y: y, Z: z}
	}

	f.Faces = make([]*datatypes.Polygon, 0)
	for i := int32(0); i < faceCount; i++ {
		faceVertexCount, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read face vertex count: %w", err)
		}

		faceVertices := make([]int, faceVertexCount)
		for v := int32(0); v < faceVertexCount; v++ {
			idx, err := reader.ReadInt32()
			if err != nil {
				return fmt.Errorf("failed to read face vertex index: %w", err)
			}
			faceVertices[v] = int(idx)
		}

		if hasNormalAbcd {
			// Read normalAbcd but don't store it
			for j := 0; j < 4; j++ {
				_, err := reader.ReadFloat32()
				if err != nil {
					return fmt.Errorf("failed to read normal: %w", err)
				}
			}
		}

		// 4 vertices will result in 2 triangles
		polygonCount := faceVertexCount - 2
		for p := int32(0); p < polygonCount; p++ {
			polygon := &datatypes.Polygon{
				IsSolid: true,
				Vertex1: faceVertices[0],
				Vertex2: faceVertices[p+1],
				Vertex3: faceVertices[p+2],
			}
			f.Faces = append(f.Faces, polygon)
		}
	}

	return nil
}
