package fragments

import (
	"encoding/binary"
	"fmt"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/datatypes"
)

// LegacyMesh (0x2C)
// Internal name: _DMSPRITEDEF
// This fragment is only found in the gequip archives and while it exists and is functional, it is not used.
// It looks like an earlier version of the Mesh fragment with fewer data points.
type LegacyMesh struct {
	BaseFragment
	Center                    datatypes.Vec3
	Vertices                  []datatypes.Vec3
	TexCoords                 []datatypes.Vec2
	Normals                   []datatypes.Vec3
	Polygons                  []*datatypes.Polygon
	VertexTex                 [][2]int
	Colors                    []datatypes.Color
	RenderGroups              []*datatypes.RenderGroup
	MaterialList              Fragment
	PolyhedronReference       *PolyhedronReference
	MobPieces                 map[int]*datatypes.MobVertexPiece
	AnimatedVerticesReference *MeshAnimatedVerticesReference
	ExportSeparateCollision   bool
}

// FragmentType returns the fragment type ID.
func (f *LegacyMesh) FragmentType() uint32 {
	return 0x2C
}

// Initialize parses the fragment data.
func (f *LegacyMesh) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)
	reader := NewFragmentReader(data)

	nameRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	// TODO: investigate flags further
	// looks like some flags will zero and 1.0 fields if they are missing.
	flags, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	hasCenterOffset := (flags & (1 << 0)) != 0
	hasBoundingRadius := (flags & (1 << 1)) != 0
	hasBit9 := (flags & (1 << 9)) != 0
	hasColors := (flags & (1 << 10)) != 0
	hasRenderGroups := (flags & (1 << 11)) != 0
	hasVertexTex := (flags & (1 << 12)) != 0
	hasBit13 := (flags & (1 << 13)) != 0
	hasBoundingBox := (flags & (1 << 14)) != 0
	// hasBit15 := (flags & (1 << 15)) != 0

	vertexCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read vertex count: %w", err)
	}

	texCoordCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read tex coord count: %w", err)
	}

	normalsCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read normals count: %w", err)
	}

	colorsCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read colors count: %w", err)
	}

	polygonCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read polygon count: %w", err)
	}

	size6, err := reader.ReadInt16()
	if err != nil {
		return fmt.Errorf("failed to read size6: %w", err)
	}

	_, err = reader.ReadInt16() // fragment1Maybe
	if err != nil {
		return fmt.Errorf("failed to read fragment1Maybe: %w", err)
	}

	vertexPieceCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read vertex piece count: %w", err)
	}

	materialListRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read material list ref: %w", err)
	}
	if materialListRef > 0 && int(materialListRef-1) < len(fragments) {
		f.MaterialList = fragments[materialListRef-1]
	}

	meshAnimation, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read mesh animation: %w", err)
	}
	// Vertex animation only
	if meshAnimation != 0 {
		if int(meshAnimation-1) < len(fragments) {
			if animRef, ok := fragments[meshAnimation-1].(*MeshAnimatedVerticesReference); ok {
				f.AnimatedVerticesReference = animRef
			}
		}
	}

	_, err = reader.ReadFloat32() // something1
	if err != nil {
		return fmt.Errorf("failed to read something1: %w", err)
	}

	// This might also be able to take a sphere (0x16) or sphere list (0x1a) collision volume
	polyhedronReference, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read polyhedron reference: %w", err)
	}
	if polyhedronReference > 0 {
		refIdx := int(polyhedronReference - 1)
		if refIdx < len(fragments) {
			if polyRef, ok := fragments[refIdx].(*PolyhedronReference); ok {
				f.PolyhedronReference = polyRef
			}
		}
		f.ExportSeparateCollision = true
	}

	centerX, err := reader.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read center x: %w", err)
	}
	centerY, err := reader.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read center y: %w", err)
	}
	centerZ, err := reader.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read center z: %w", err)
	}
	if hasCenterOffset {
		f.Center = datatypes.Vec3{X: centerX, Y: centerY, Z: centerZ}
	} else {
		f.Center = datatypes.Vec3{X: 0, Y: 0, Z: 0}
	}

	boundingRadius, err := reader.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read bounding radius: %w", err)
	}
	if !hasBoundingRadius {
		_ = boundingRadius // unused if bit not set
	}

	// Read vertices
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

	// Read tex coords
	f.TexCoords = make([]datatypes.Vec2, texCoordCount)
	for i := int32(0); i < texCoordCount; i++ {
		u, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read tex coord u: %w", err)
		}
		v, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read tex coord v: %w", err)
		}
		f.TexCoords[i] = datatypes.Vec2{X: u, Y: v}
	}

	// Read normals
	f.Normals = make([]datatypes.Vec3, normalsCount)
	for i := int32(0); i < normalsCount; i++ {
		x, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read normal x: %w", err)
		}
		y, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read normal y: %w", err)
		}
		z, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read normal z: %w", err)
		}
		f.Normals[i] = datatypes.Vec3{X: x, Y: y, Z: z}
	}

	// Read unknown color data (count matches vertexCount when present)
	for i := int32(0); i < colorsCount; i++ {
		_, err := reader.ReadBytes(4) // unkBytes
		if err != nil {
			return fmt.Errorf("failed to read color bytes: %w", err)
		}
	}

	// Read faces/polygons
	f.Polygons = make([]*datatypes.Polygon, polygonCount)
	for i := int32(0); i < polygonCount; i++ {
		_, err := reader.ReadInt16() // flag
		if err != nil {
			return fmt.Errorf("failed to read polygon flag: %w", err)
		}

		_, err = reader.ReadInt16() // unk1
		if err != nil {
			return fmt.Errorf("failed to read polygon unk1: %w", err)
		}

		materialIndex, err := reader.ReadInt16()
		if err != nil {
			return fmt.Errorf("failed to read polygon material index: %w", err)
		}

		_, err = reader.ReadInt16() // unk3
		if err != nil {
			return fmt.Errorf("failed to read polygon unk3: %w", err)
		}

		_, err = reader.ReadInt16() // unk4
		if err != nil {
			return fmt.Errorf("failed to read polygon unk4: %w", err)
		}

		i1, err := reader.ReadInt16()
		if err != nil {
			return fmt.Errorf("failed to read polygon i1: %w", err)
		}

		i2, err := reader.ReadInt16()
		if err != nil {
			return fmt.Errorf("failed to read polygon i2: %w", err)
		}

		i3, err := reader.ReadInt16()
		if err != nil {
			return fmt.Errorf("failed to read polygon i3: %w", err)
		}

		f.Polygons[i] = &datatypes.Polygon{
			IsSolid:       true,
			Vertex1:       int(i1),
			Vertex2:       int(i2),
			Vertex3:       int(i3),
			MaterialIndex: int(materialIndex),
		}
	}

	// Read meshops
	for i := int16(0); i < size6; i++ {
		datatype, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read meshop datatype: %w", err)
		}

		if datatype != 4 {
			_, err := reader.ReadInt32() // vertexIndex
			if err != nil {
				return fmt.Errorf("failed to read meshop vertex index: %w", err)
			}
			_, err = reader.ReadInt16() // data6Param1
			if err != nil {
				return fmt.Errorf("failed to read meshop param1: %w", err)
			}
			_, err = reader.ReadInt16() // data6Param2
			if err != nil {
				return fmt.Errorf("failed to read meshop param2: %w", err)
			}
		} else {
			_, err := reader.ReadFloat32() // offset
			if err != nil {
				return fmt.Errorf("failed to read meshop offset: %w", err)
			}
			_, err = reader.ReadInt32() // something
			if err != nil {
				return fmt.Errorf("failed to read meshop something: %w", err)
			}
		}
	}

	// Read mob pieces
	f.MobPieces = make(map[int]*datatypes.MobVertexPiece)
	for i := int32(0); i < vertexPieceCount; i++ {
		count, err := reader.ReadInt16()
		if err != nil {
			return fmt.Errorf("failed to read mob piece count: %w", err)
		}
		start, err := reader.ReadInt16()
		if err != nil {
			return fmt.Errorf("failed to read mob piece start: %w", err)
		}

		mobVertexPiece := &datatypes.MobVertexPiece{
			Count: int(count),
			Start: int(start),
		}
		f.MobPieces[int(start)] = mobVertexPiece
	}

	if hasBit9 {
		size8, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read size8: %w", err)
		}
		err = reader.Skip(int64(size8 * 4))
		if err != nil {
			return fmt.Errorf("failed to skip bit9 data: %w", err)
		}
	}

	// Found in qrg R1. count matches vertex count
	// this might be vertex colors?
	if hasColors {
		unkCount, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read unk count: %w", err)
		}
		f.Colors = make([]datatypes.Color, unkCount)
		for i := int32(0); i < unkCount; i++ {
			colorBytes, err := reader.ReadBytes(4)
			if err != nil {
				return fmt.Errorf("failed to read color bytes: %w", err)
			}
			colorVal := binary.LittleEndian.Uint32(colorBytes)
			b := int(colorVal & 0xFF)
			g := int((colorVal >> 8) & 0xFF)
			r := int((colorVal >> 16) & 0xFF)
			a := int((colorVal >> 24) & 0xFF)
			f.Colors[i] = datatypes.NewColor(r, g, b, a)
		}
	}

	if hasRenderGroups {
		polygonTexCount, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read polygon tex count: %w", err)
		}
		f.RenderGroups = make([]*datatypes.RenderGroup, polygonTexCount)
		for i := int32(0); i < polygonTexCount; i++ {
			polyCount, err := reader.ReadInt16()
			if err != nil {
				return fmt.Errorf("failed to read render group polygon count: %w", err)
			}
			matIndex, err := reader.ReadInt16()
			if err != nil {
				return fmt.Errorf("failed to read render group material index: %w", err)
			}
			f.RenderGroups[i] = &datatypes.RenderGroup{
				PolygonCount:  int(polyCount),
				MaterialIndex: int(matIndex),
			}
		}
	}

	if hasVertexTex {
		vertexTexCount, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read vertex tex count: %w", err)
		}
		f.VertexTex = make([][2]int, vertexTexCount)
		for i := int32(0); i < vertexTexCount; i++ {
			x, err := reader.ReadInt16()
			if err != nil {
				return fmt.Errorf("failed to read vertex tex x: %w", err)
			}
			y, err := reader.ReadInt16()
			if err != nil {
				return fmt.Errorf("failed to read vertex tex y: %w", err)
			}
			f.VertexTex[i] = [2]int{int(x), int(y)}
		}
	}

	// TODO: Research: Instead of controlling the presence, the fields might be zeroed if the bit isn't set
	if hasBit13 {
		_, err := reader.ReadFloat32() // params31
		if err != nil {
			return fmt.Errorf("failed to read params31: %w", err)
		}
		_, err = reader.ReadFloat32() // params32
		if err != nil {
			return fmt.Errorf("failed to read params32: %w", err)
		}
		_, err = reader.ReadFloat32() // params33
		if err != nil {
			return fmt.Errorf("failed to read params33: %w", err)
		}
	}

	// Bounding Box?
	if hasBoundingBox {
		for j := 0; j < 6; j++ {
			_, err := reader.ReadFloat32() // 2 vec3s
			if err != nil {
				return fmt.Errorf("failed to read bounding box: %w", err)
			}
		}
	}

	// In some rare cases, the number of uvs does not match the number of vertices
	if vertexCount != texCoordCount {
		difference := vertexCount - texCoordCount
		for i := int32(0); i < difference; i++ {
			f.TexCoords = append(f.TexCoords, datatypes.Vec2{X: 0.0, Y: 0.0})
		}
	}

	return nil
}
