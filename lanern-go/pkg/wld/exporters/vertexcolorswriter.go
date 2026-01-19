package exporters

import (
	"strconv"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/fragments"
)

// VertexColorsWriter exports vertex color data to a text format.
type VertexColorsWriter struct {
	TextAssetWriter
}

// NewVertexColorsWriter creates a new VertexColorsWriter.
func NewVertexColorsWriter() *VertexColorsWriter {
	w := &VertexColorsWriter{}
	w.addHeader()
	return w
}

// addHeader adds the header to the export buffer.
func (w *VertexColorsWriter) addHeader() {
	w.AppendLine(ExportHeaderTitle + "Vertex Colors")
	w.AppendLine(ExportHeaderFormat + "Red, Green, Blue, Sunlight")
}

// AddFragmentData adds vertex colors fragment data to the export buffer.
func (w *VertexColorsWriter) AddFragmentData(data fragments.Fragment) {
	instance, ok := data.(*fragments.VertexColors)
	if !ok || instance == nil {
		return
	}

	for _, color := range instance.Colors {
		w.AppendString(strconv.Itoa(color.R))
		w.AppendString(",")
		w.AppendString(strconv.Itoa(color.G))
		w.AppendString(",")
		w.AppendString(strconv.Itoa(color.B))
		w.AppendString(",")
		w.AppendLine(strconv.Itoa(color.A))
	}
}

// ClearExportData clears the export buffer and re-adds the header.
func (w *VertexColorsWriter) ClearExportData() {
	w.TextAssetWriter.ClearExportData()
	w.addHeader()
}
