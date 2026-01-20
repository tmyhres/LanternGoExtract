package datatypes

// RegionWall represents a wall within a region.
type RegionWall struct {
	// Flags contains bit flags:
	// bit 0 - has FLOOR (is floor?)
	// bit 1 - has RENDERMETHOD and NORMALABCD (is renderable?)
	Flags int

	// NumVertices: NUMVERTICES %d
	NumVertices int

	// RenderMethod: RENDERMETHOD ...
	RenderMethod *RenderMethod

	// RenderInfo: RENDERINFO
	RenderInfo *RenderInfo

	// NormalAbcd: NORMALABCD %f %f %f %f
	NormalAbcd Vec4

	// VertexList: VERTEXLIST %d ...%d
	// Binary values are 0 based. "VERTEXLIST 1" becomes vertex_list[0]
	VertexList []int
}
