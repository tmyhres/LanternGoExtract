package datatypes

// RegionType represents the type of region.
type RegionType int

const (
	RegionTypeNormal        RegionType = 0
	RegionTypeWater         RegionType = 1
	RegionTypeLava          RegionType = 2
	RegionTypePvp           RegionType = 3
	RegionTypeZoneline      RegionType = 4
	RegionTypeWaterBlockLos RegionType = 5
	RegionTypeFreezingWater RegionType = 6
	RegionTypeSlippery      RegionType = 7
	RegionTypeUnknown       RegionType = 8
)
