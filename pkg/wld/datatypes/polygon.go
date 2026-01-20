package datatypes

// Polygon represents a triangle polygon with material information.
type Polygon struct {
	IsSolid       bool
	Vertex1       int
	Vertex2       int
	Vertex3       int
	MaterialIndex int
}

// Copy returns a copy of the polygon.
func (p *Polygon) Copy() *Polygon {
	return &Polygon{
		IsSolid:       p.IsSolid,
		Vertex1:       p.Vertex1,
		Vertex2:       p.Vertex2,
		Vertex3:       p.Vertex3,
		MaterialIndex: p.MaterialIndex,
	}
}
