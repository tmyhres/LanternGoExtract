package datatypes

// RegionVisNode represents a visibility node for a region.
type RegionVisNode struct {
	// NormalAbcd: NORMALABCD %f %f %f %f
	NormalAbcd Vec4

	// VisListIndex: VISLISTINDEX %d
	VisListIndex int

	// FrontTree: FRONTTREE %d
	FrontTree int

	// BackTree: BACKTREE %d
	BackTree int
}
