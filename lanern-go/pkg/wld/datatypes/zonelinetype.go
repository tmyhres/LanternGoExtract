package datatypes

// ZonelineType represents the type of zone line.
type ZonelineType int

const (
	// ZonelineTypeReference is a reference-based zone line.
	ZonelineTypeReference ZonelineType = 0
	// ZonelineTypeAbsolute is an absolute zone line.
	ZonelineTypeAbsolute ZonelineType = 1
)
