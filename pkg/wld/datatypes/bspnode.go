package datatypes

// BspRegion is a forward declaration interface for BSP region fragments.
// The actual implementation should be in the fragments package.
type BspRegion interface{}

// BspNode represents a node in a binary space partitioning tree.
type BspNode struct {
	NormalX       float32
	NormalY       float32
	NormalZ       float32
	SplitDistance float32
	RegionID      int
	LeftNode      int
	RightNode     int
	Region        BspRegion
}
