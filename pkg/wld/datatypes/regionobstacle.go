package datatypes

// RegionObstacle represents an obstacle within a region.
type RegionObstacle struct {
	// Flags contains bit flags:
	// bit 0 - is a FLOOR
	// bit 1 - is a GEOMETRYCUTTINGOBSTACLE
	// bit 2 - has USERDATA %s
	Flags int

	// NextRegion: NEXTREGION %d
	NextRegion int

	// ObstacleType determines the type of obstacle:
	// XY_VERTEX 0 %d
	// XYZ_VERTEX 0 %d
	// XY_LINE 0 %d %d
	// XY_EDGE 0 %d %d
	// XYZ_EDGE 0 %d %d
	// PLANE 0 %d
	// EDGEPOLYGON 0
	// EDGEWALL 0 %d
	ObstacleType RegionObstacleType

	// NumVertices: NUMVERTICES %d
	NumVertices int

	// VertexList: VERTEXLIST %d ...%d
	VertexList []int

	// NormalAbcd: NORMALABCD %f %f %f %f
	NormalAbcd Vec4

	// EdgeWall: EDGEWALL 0 %d
	// Binary values are 0 based. "EDGEWALL 0 1" becomes edge_wall[0]
	EdgeWall int

	// UserDataSize is the length of USERDATA string
	UserDataSize int

	// UserData: USERDATA %s
	UserData []byte
}
