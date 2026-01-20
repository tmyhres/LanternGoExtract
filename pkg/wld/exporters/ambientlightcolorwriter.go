package exporters

import (
	"strconv"

	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
)

// AmbientLightColorWriter exports ambient light color data to a text format.
type AmbientLightColorWriter struct {
	TextAssetWriter
}

// NewAmbientLightColorWriter creates a new AmbientLightColorWriter.
func NewAmbientLightColorWriter() *AmbientLightColorWriter {
	w := &AmbientLightColorWriter{}
	w.AppendLine(ExportHeaderTitle + "Ambient Light Color")
	w.AppendLine(ExportHeaderFormat + "R, G, B")
	return w
}

// AddFragmentData adds global ambient light fragment data to the export buffer.
func (w *AmbientLightColorWriter) AddFragmentData(data fragments.Fragment) {
	globalAmbientLight, ok := data.(*fragments.GlobalAmbientLight)
	if !ok || globalAmbientLight == nil {
		return
	}

	w.AppendString(strconv.Itoa(globalAmbientLight.Color.R))
	w.AppendString(",")
	w.AppendString(strconv.Itoa(globalAmbientLight.Color.G))
	w.AppendString(",")
	w.AppendString(strconv.Itoa(globalAmbientLight.Color.B))
}
