package fragments

import (
	"math"

	"github.com/tmyhres/LanternGoExtract/pkg/wld/datatypes"
)

// Mesh (0x36) contains geometric data for a mesh.
// Internal name: _DMSPRITEDEF
// This is the primary mesh format used for zone geometry and objects.
type Mesh struct {
	BaseFragment

	// Center is the center of the mesh - used to calculate absolute coordinates of vertices.
	// Animated meshes use coordinates relative to the center.
	Center Vec3

	// MaxDistance is the maximum distance between the center and any vertex (bounding radius).
	MaxDistance float32

	// MinPosition is the minimum vertex positions in the model (bounding box min).
	MinPosition Vec3

	// MaxPosition is the maximum vertex positions in the model (bounding box max).
	MaxPosition Vec3

	// MaterialList is the texture list used to render this mesh.
	// In zone meshes, it's always the same one.
	// In object meshes, it can be unique.
	MaterialList *MaterialList

	// Vertices is the list of mesh vertices.
	Vertices []Vec3

	// Normals is the list of mesh normals.
	Normals []Vec3

	// Indices is the list of polygon indices.
	Indices []datatypes.Polygon

	// Colors is the list of vertex colors.
	Colors []datatypes.Color

	// TextureUvCoordinates is the list of UV texture coordinates.
	TextureUvCoordinates []datatypes.Vec2

	// MaterialGroups defines which texture index corresponds with groups of vertices.
	MaterialGroups []datatypes.RenderGroup

	// AnimatedVerticesReference is the reference to animated vertex fragment (0x37).
	AnimatedVerticesReference Fragment

	// ExportSeparateCollision is true if there are non-solid polygons in the mesh.
	// This means collision should be exported separately (e.g. trees, fire).
	ExportSeparateCollision bool

	// IsHandled indicates if this mesh has been processed.
	IsHandled bool

	// StartTextureIndex is the starting texture index.
	StartTextureIndex int

	// MobPieces are the render components of a mob skeleton.
	MobPieces map[int]datatypes.MobVertexPiece
}

// FragmentType returns the fragment type ID (0x36).
func (m *Mesh) FragmentType() uint32 {
	return 0x36
}

// Initialize parses the mesh fragment data.
func (m *Mesh) Initialize(index int, id int, size int, data []byte, fragments []Fragment,
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
	// Zone: 0x00018003, Objects: 0x00014003
	_, err = r.ReadInt32()
	if err != nil {
		return err
	}

	// Read material list reference
	materialListRef, err := r.ReadInt32()
	if err != nil {
		return err
	}
	if int(materialListRef)-1 >= 0 && int(materialListRef)-1 < len(fragments) {
		if ml, ok := fragments[materialListRef-1].(*MaterialList); ok {
			m.MaterialList = ml
		}
	}

	// Read mesh animation reference
	meshAnimation, err := r.ReadInt32()
	if err != nil {
		return err
	}

	// Vertex animation only
	if meshAnimation != 0 && int(meshAnimation)-1 < len(fragments) {
		m.AnimatedVerticesReference = fragments[meshAnimation-1]
	}

	// Read unknown values
	_, err = r.ReadInt32() // unknown
	if err != nil {
		return err
	}

	_, err = r.ReadInt32() // unknown2 - maybe references the first 0x03 in the WLD
	if err != nil {
		return err
	}

	// Read center position
	centerX, err := r.ReadFloat32()
	if err != nil {
		return err
	}
	centerY, err := r.ReadFloat32()
	if err != nil {
		return err
	}
	centerZ, err := r.ReadFloat32()
	if err != nil {
		return err
	}
	m.Center = Vec3{X: centerX, Y: centerY, Z: centerZ}

	// Read 3 unknown dwords (seems to be related to lighting models like torches)
	_, err = r.ReadInt32() // unknownDword1
	if err != nil {
		return err
	}
	_, err = r.ReadInt32() // unknownDword2
	if err != nil {
		return err
	}
	_, err = r.ReadInt32() // unknownDword3
	if err != nil {
		return err
	}

	// Read max distance (bounding radius)
	m.MaxDistance, err = r.ReadFloat32()
	if err != nil {
		return err
	}

	// Read min position
	minX, err := r.ReadFloat32()
	if err != nil {
		return err
	}
	minY, err := r.ReadFloat32()
	if err != nil {
		return err
	}
	minZ, err := r.ReadFloat32()
	if err != nil {
		return err
	}
	m.MinPosition = Vec3{X: minX, Y: minY, Z: minZ}

	// Read max position
	maxX, err := r.ReadFloat32()
	if err != nil {
		return err
	}
	maxY, err := r.ReadFloat32()
	if err != nil {
		return err
	}
	maxZ, err := r.ReadFloat32()
	if err != nil {
		return err
	}
	m.MaxPosition = Vec3{X: maxX, Y: maxY, Z: maxZ}

	// Read counts
	vertexCount, err := r.ReadInt16()
	if err != nil {
		return err
	}
	textureCoordinateCount, err := r.ReadInt16()
	if err != nil {
		return err
	}
	normalsCount, err := r.ReadInt16()
	if err != nil {
		return err
	}
	colorsCount, err := r.ReadInt16()
	if err != nil {
		return err
	}
	polygonCount, err := r.ReadInt16()
	if err != nil {
		return err
	}
	vertexPieceCount, err := r.ReadInt16()
	if err != nil {
		return err
	}
	polygonTextureCount, err := r.ReadInt16()
	if err != nil {
		return err
	}
	vertexTextureCount, err := r.ReadInt16()
	if err != nil {
		return err
	}
	size9, err := r.ReadInt16()
	if err != nil {
		return err
	}
	scaleValue, err := r.ReadInt16()
	if err != nil {
		return err
	}
	scale := float32(1.0) / float32(int(1)<<scaleValue)

	// Initialize slices
	m.Vertices = make([]Vec3, 0, vertexCount)
	m.Normals = make([]Vec3, 0, normalsCount)
	m.Colors = make([]datatypes.Color, 0, colorsCount)
	m.TextureUvCoordinates = make([]datatypes.Vec2, 0, textureCoordinateCount)

	// Read vertices
	for i := int16(0); i < vertexCount; i++ {
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
		m.Vertices = append(m.Vertices, Vec3{
			X: float32(x) * scale,
			Y: float32(y) * scale,
			Z: float32(z) * scale,
		})
	}

	// Read texture coordinates
	for i := int16(0); i < textureCoordinateCount; i++ {
		var u, v float32
		if isNewFormat {
			u, err = r.ReadFloat32()
			if err != nil {
				return err
			}
			v, err = r.ReadFloat32()
			if err != nil {
				return err
			}
		} else {
			uInt, err := r.ReadInt16()
			if err != nil {
				return err
			}
			vInt, err := r.ReadInt16()
			if err != nil {
				return err
			}
			u = float32(uInt) / 256.0
			v = float32(vInt) / 256.0
		}
		m.TextureUvCoordinates = append(m.TextureUvCoordinates, datatypes.Vec2{X: u, Y: v})
	}

	// Read normals
	for i := int16(0); i < normalsCount; i++ {
		xByte, err := r.ReadByte()
		if err != nil {
			return err
		}
		yByte, err := r.ReadByte()
		if err != nil {
			return err
		}
		zByte, err := r.ReadByte()
		if err != nil {
			return err
		}
		// Convert from signed byte to float
		x := float32(int8(xByte)) / 128.0
		y := float32(int8(yByte)) / 128.0
		z := float32(int8(zByte)) / 128.0
		m.Normals = append(m.Normals, Vec3{X: x, Y: y, Z: z})
	}

	// Read colors
	for i := int16(0); i < colorsCount; i++ {
		colorValue, err := r.ReadInt32()
		if err != nil {
			return err
		}
		// Color is stored as BGRA
		b := int(colorValue & 0xFF)
		g := int((colorValue >> 8) & 0xFF)
		red := int((colorValue >> 16) & 0xFF)
		a := int((colorValue >> 24) & 0xFF)
		m.Colors = append(m.Colors, datatypes.NewColor(red, g, b, a))
	}

	// Read polygons
	m.Indices = make([]datatypes.Polygon, 0, polygonCount)
	for i := int16(0); i < polygonCount; i++ {
		solidFlag, err := r.ReadInt16()
		if err != nil {
			return err
		}
		isSolid := solidFlag == 0

		if !isSolid {
			m.ExportSeparateCollision = true
		}

		v1, err := r.ReadInt16()
		if err != nil {
			return err
		}
		v2, err := r.ReadInt16()
		if err != nil {
			return err
		}
		v3, err := r.ReadInt16()
		if err != nil {
			return err
		}

		m.Indices = append(m.Indices, datatypes.Polygon{
			IsSolid: isSolid,
			Vertex1: int(v1),
			Vertex2: int(v2),
			Vertex3: int(v3),
		})
	}

	// Read mob pieces (vertex pieces)
	m.MobPieces = make(map[int]datatypes.MobVertexPiece)
	mobStart := 0
	for i := int16(0); i < vertexPieceCount; i++ {
		count, err := r.ReadInt16()
		if err != nil {
			return err
		}
		index1, err := r.ReadInt16()
		if err != nil {
			return err
		}
		m.MobPieces[int(index1)] = datatypes.MobVertexPiece{
			Count: int(count),
			Start: mobStart,
		}
		mobStart += int(count)
	}

	// Read material groups (polygon texture groups)
	m.MaterialGroups = make([]datatypes.RenderGroup, 0, polygonTextureCount)
	m.StartTextureIndex = math.MaxInt32

	for i := int16(0); i < polygonTextureCount; i++ {
		polyCount, err := r.ReadUint16()
		if err != nil {
			return err
		}
		matIndex, err := r.ReadUint16()
		if err != nil {
			return err
		}
		group := datatypes.RenderGroup{
			PolygonCount:  int(polyCount),
			MaterialIndex: int(matIndex),
		}
		m.MaterialGroups = append(m.MaterialGroups, group)

		if int(matIndex) < m.StartTextureIndex {
			m.StartTextureIndex = int(matIndex)
		}
	}

	// Skip vertex texture data
	for i := int16(0); i < vertexTextureCount; i++ {
		err = r.Skip(4)
		if err != nil {
			return err
		}
	}

	// Skip size9 data (12 bytes each)
	for i := int16(0); i < size9; i++ {
		err = r.Skip(12)
		if err != nil {
			return err
		}
	}

	// In some rare cases, the number of uvs does not match the number of vertices
	if len(m.Vertices) != len(m.TextureUvCoordinates) {
		difference := len(m.Vertices) - len(m.TextureUvCoordinates)
		for i := 0; i < difference; i++ {
			m.TextureUvCoordinates = append(m.TextureUvCoordinates, datatypes.Vec2{X: 0.0, Y: 0.0})
		}
	}

	return nil
}

// ClearCollision sets all polygons to non-solid.
func (m *Mesh) ClearCollision() {
	for i := range m.Indices {
		m.Indices[i].IsSolid = false
	}
	m.ExportSeparateCollision = true
}
