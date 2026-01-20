package datatypes

// Color represents an RGBA color with integer components.
type Color struct {
	R int
	G int
	B int
	A int
}

// NewColor creates a new Color with the specified RGBA values.
func NewColor(r, g, b, a int) Color {
	return Color{
		R: r,
		G: g,
		B: b,
		A: a,
	}
}
