package datatypes

// ZonelineInfo contains information about a zone line.
type ZonelineInfo struct {
	Type      ZonelineType
	Index     int
	Position  Vec3
	Heading   int
	ZoneIndex int
}
