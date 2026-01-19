package fragments

import (
	"fmt"

	"github.com/lanterneq/lanern-go/pkg/wld/datatypes"
)

// BspRegion (0x22)
// Internal Name: None
// Leaf nodes in the BSP tree. Can contain references to Mesh fragments.
// This fragment's PVS (potentially visible set) data is unhandled.
type BspRegion struct {
	BaseFragment
	// ContainsPolygons indicates if this fragment contains geometry
	ContainsPolygons bool
	// Mesh is a reference to the mesh fragment
	Mesh Fragment
	// LegacyMesh is a reference to the legacy mesh fragment
	LegacyMesh *LegacyMesh
	// RegionType is the type of this region
	RegionType *BspRegionType
	// RegionVertices contains the vertices for this region
	RegionVertices []datatypes.Vec3
}

// FragmentType returns the fragment type ID.
func (f *BspRegion) FragmentType() uint32 {
	return 0x22
}

// Initialize parses the fragment data.
func (f *BspRegion) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
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

	hasSphere := (flags & (1 << 0)) != 0
	hasReverbVolume := (flags & (1 << 1)) != 0
	hasReverbOffset := (flags & (1 << 2)) != 0
	// regionFog := (flags & (1 << 3)) != 0
	// enableGoraud2 := (flags & (1 << 4)) != 0
	// encodedVisibility := (flags & (1 << 5)) != 0
	hasLegacyMeshReference := (flags & (1 << 6)) != 0
	hasByteEntries := (flags & (1 << 7)) != 0
	hasMeshReference := (flags & (1 << 8)) != 0

	f.ContainsPolygons = hasMeshReference || hasLegacyMeshReference

	// Always 0
	_, err = reader.ReadInt32() // ambientLight
	if err != nil {
		return fmt.Errorf("failed to read ambient light: %w", err)
	}

	numRegionVertex, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read num region vertex: %w", err)
	}

	numProximalRegions, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read num proximal regions: %w", err)
	}

	// Always 0
	numRenderVertices, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read num render vertices: %w", err)
	}

	numWalls, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read num walls: %w", err)
	}

	numObstacles, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read num obstacles: %w", err)
	}

	// Always 0
	_, err = reader.ReadInt32() // numCuttingObstacles
	if err != nil {
		return fmt.Errorf("failed to read num cutting obstacles: %w", err)
	}

	numVisNode, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read num vis node: %w", err)
	}

	numVisList, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read num vis list: %w", err)
	}

	// Read region vertices
	f.RegionVertices = make([]datatypes.Vec3, numRegionVertex)
	for i := int32(0); i < numRegionVertex; i++ {
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
		f.RegionVertices[i] = datatypes.Vec3{X: x, Y: y, Z: z}
	}

	// Read proximal regions
	for i := int32(0); i < numProximalRegions; i++ {
		_, err := reader.ReadInt32() // region index
		if err != nil {
			return fmt.Errorf("failed to read proximal region index: %w", err)
		}
		_, err = reader.ReadFloat32() // distance
		if err != nil {
			return fmt.Errorf("failed to read proximal region distance: %w", err)
		}
	}

	// Read render vertices
	for i := int32(0); i < numRenderVertices; i++ {
		_, err := reader.ReadFloat32() // x
		if err != nil {
			return fmt.Errorf("failed to read render vertex x: %w", err)
		}
		_, err = reader.ReadFloat32() // y
		if err != nil {
			return fmt.Errorf("failed to read render vertex y: %w", err)
		}
		_, err = reader.ReadFloat32() // z
		if err != nil {
			return fmt.Errorf("failed to read render vertex z: %w", err)
		}
	}

	// Read walls
	for i := int32(0); i < numWalls; i++ {
		wallFlags, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read wall flags: %w", err)
		}
		isRenderable := (wallFlags & (1 << 1)) != 0

		wallNumVertices, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read wall num vertices: %w", err)
		}

		for v := int32(0); v < wallNumVertices; v++ {
			_, err := reader.ReadInt32() // vertex index
			if err != nil {
				return fmt.Errorf("failed to read wall vertex: %w", err)
			}
		}

		if isRenderable {
			_, err := reader.ReadInt32() // render method flags
			if err != nil {
				return fmt.Errorf("failed to read render method flags: %w", err)
			}

			// Parse RenderInfo
			_, err = datatypes.ParseRenderInfo(reader.Reader, toWldFragments(fragments))
			if err != nil {
				return fmt.Errorf("failed to parse render info: %w", err)
			}

			// Read normalAbcd
			for j := 0; j < 4; j++ {
				_, err := reader.ReadFloat32()
				if err != nil {
					return fmt.Errorf("failed to read wall normal: %w", err)
				}
			}
		}
	}

	// Read obstacles
	for i := int32(0); i < numObstacles; i++ {
		obstacleFlags, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read obstacle flags: %w", err)
		}
		hasUserData := (obstacleFlags & (1 << 2)) != 0

		_, err = reader.ReadInt32() // nextRegion
		if err != nil {
			return fmt.Errorf("failed to read next region: %w", err)
		}

		obstacleType, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read obstacle type: %w", err)
		}

		obstacleNumVertices := int32(0)
		if datatypes.RegionObstacleType(obstacleType) == datatypes.RegionObstacleTypeEdgePolygon ||
			datatypes.RegionObstacleType(obstacleType) == datatypes.RegionObstacleTypeEdgePolygonNormalAbcd {
			obstacleNumVertices, err = reader.ReadInt32()
			if err != nil {
				return fmt.Errorf("failed to read obstacle num vertices: %w", err)
			}
		}

		for v := int32(0); v < obstacleNumVertices; v++ {
			_, err := reader.ReadInt32() // vertex index
			if err != nil {
				return fmt.Errorf("failed to read obstacle vertex: %w", err)
			}
		}

		if datatypes.RegionObstacleType(obstacleType) == datatypes.RegionObstacleTypeEdgePolygonNormalAbcd {
			for j := 0; j < 4; j++ {
				_, err := reader.ReadFloat32() // normalAbcd
				if err != nil {
					return fmt.Errorf("failed to read obstacle normal: %w", err)
				}
			}
		}

		if datatypes.RegionObstacleType(obstacleType) == datatypes.RegionObstacleTypeEdgeWall {
			_, err := reader.ReadInt32() // edgeWall
			if err != nil {
				return fmt.Errorf("failed to read edge wall: %w", err)
			}
		}

		if hasUserData {
			userDataSize, err := reader.ReadInt32()
			if err != nil {
				return fmt.Errorf("failed to read user data size: %w", err)
			}
			_, err = reader.ReadBytes(int(userDataSize))
			if err != nil {
				return fmt.Errorf("failed to read user data: %w", err)
			}
		}
	}

	// Read vis nodes
	for i := int32(0); i < numVisNode; i++ {
		for j := 0; j < 4; j++ {
			_, err := reader.ReadFloat32() // normalAbcd
			if err != nil {
				return fmt.Errorf("failed to read vis node normal: %w", err)
			}
		}
		_, err := reader.ReadInt32() // visListIndex
		if err != nil {
			return fmt.Errorf("failed to read vis list index: %w", err)
		}
		_, err = reader.ReadInt32() // frontTree
		if err != nil {
			return fmt.Errorf("failed to read front tree: %w", err)
		}
		_, err = reader.ReadInt32() // backTree
		if err != nil {
			return fmt.Errorf("failed to read back tree: %w", err)
		}
	}

	// Read vis lists
	for i := int32(0); i < numVisList; i++ {
		rangeCount, err := reader.ReadInt16()
		if err != nil {
			return fmt.Errorf("failed to read range count: %w", err)
		}

		for r := int16(0); r < rangeCount; r++ {
			if hasByteEntries {
				_, err := reader.ReadByte()
				if err != nil {
					return fmt.Errorf("failed to read range byte: %w", err)
				}
			} else {
				_, err := reader.ReadInt16()
				if err != nil {
					return fmt.Errorf("failed to read range int16: %w", err)
				}
			}
		}
	}

	if hasSphere {
		for j := 0; j < 4; j++ {
			_, err := reader.ReadFloat32()
			if err != nil {
				return fmt.Errorf("failed to read sphere: %w", err)
			}
		}
	}

	if hasReverbVolume {
		_, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read reverb volume: %w", err)
		}
	}

	if hasReverbOffset {
		_, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read reverb offset: %w", err)
		}
	}

	userDataSize, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read user data size: %w", err)
	}
	_, err = reader.ReadBytes(int(userDataSize))
	if err != nil {
		return fmt.Errorf("failed to read user data: %w", err)
	}

	// Get the mesh reference index and link to it
	if f.ContainsPolygons {
		meshReference, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read mesh reference: %w", err)
		}
		meshIdx := int(meshReference) - 1

		if hasMeshReference {
			if meshIdx >= 0 && meshIdx < len(fragments) {
				f.Mesh = fragments[meshIdx]
			}
		} else if hasLegacyMeshReference {
			if meshIdx >= 0 && meshIdx < len(fragments) {
				if legacyMesh, ok := fragments[meshIdx].(*LegacyMesh); ok {
					f.LegacyMesh = legacyMesh
				}
			}
		}
	}

	return nil
}

// SetRegionFlag sets the region type for this BSP region.
func (f *BspRegion) SetRegionFlag(bspRegionType *BspRegionType) {
	f.RegionType = bspRegionType
}

// toWldFragments converts fragments to the interface slice expected by ParseRenderInfo.
func toWldFragments(fragments []Fragment) []datatypes.WldFragment {
	result := make([]datatypes.WldFragment, len(fragments))
	for i, f := range fragments {
		result[i] = f
	}
	return result
}
