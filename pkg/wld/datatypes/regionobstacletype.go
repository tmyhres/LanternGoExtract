package datatypes

// RegionObstacleType represents the type of region obstacle.
type RegionObstacleType int

const (
	RegionObstacleTypeXyVertex              RegionObstacleType = 8
	RegionObstacleTypeXyzVertex             RegionObstacleType = 9
	RegionObstacleTypeXyLine                RegionObstacleType = 10
	RegionObstacleTypeXyEdge                RegionObstacleType = 11
	RegionObstacleTypeXyzEdge               RegionObstacleType = 12
	RegionObstacleTypePlane                 RegionObstacleType = 13
	RegionObstacleTypeEdgePolygon           RegionObstacleType = 14
	RegionObstacleTypeEdgeWall              RegionObstacleType = 18
	RegionObstacleTypeEdgePolygonNormalAbcd RegionObstacleType = -15
)
