package datatypes

// RegionVisList represents a visibility list for a region.
type RegionVisList struct {
	// RangeCount: RANGE %d
	RangeCount int

	// Ranges contains the visibility ranges
	Ranges []int
}
