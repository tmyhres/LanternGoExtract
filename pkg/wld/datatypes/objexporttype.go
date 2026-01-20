package datatypes

// ObjExportType represents the type of OBJ export.
// Used by the mesh class to determine which surfaces/textures are needed.
type ObjExportType int

const (
	// ObjExportTypeTextured exports with textures.
	ObjExportTypeTextured ObjExportType = iota
	// ObjExportTypeCollision exports collision geometry.
	ObjExportTypeCollision
)
