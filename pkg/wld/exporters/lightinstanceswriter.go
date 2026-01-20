package exporters

import (
	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
)

// LightInstancesWriter exports light instance data to a text format.
type LightInstancesWriter struct {
	TextAssetWriter
}

// NewLightInstancesWriter creates a new LightInstancesWriter.
func NewLightInstancesWriter() *LightInstancesWriter {
	w := &LightInstancesWriter{}
	w.AppendLine(ExportHeaderTitle + "Light Instances")
	w.AppendLine(ExportHeaderFormat + "PosX, PosY, PosZ, Radius, ColorR, ColorG, ColorB")
	return w
}

// AddFragmentData adds light instance fragment data to the export buffer.
func (w *LightInstancesWriter) AddFragmentData(data fragments.Fragment) {
	light, ok := data.(*fragments.LightInstance)
	if !ok || light == nil {
		return
	}

	// Position (note coordinate swap: x, z, y)
	w.AppendString(formatFloat(light.Position.X))
	w.AppendString(",")
	w.AppendString(formatFloat(light.Position.Z))
	w.AppendString(",")
	w.AppendString(formatFloat(light.Position.Y))
	w.AppendString(",")

	// Radius
	w.AppendString(formatFloat(light.Radius))
	w.AppendString(",")

	// Color (from light source reference)
	if light.LightReference != nil && light.LightReference.LightSource != nil {
		w.AppendString(formatFloat(light.LightReference.LightSource.Color.X))
		w.AppendString(",")
		w.AppendString(formatFloat(light.LightReference.LightSource.Color.Y))
		w.AppendString(",")
		w.AppendString(formatFloat(light.LightReference.LightSource.Color.Z))
	} else {
		w.AppendString("0,0,0")
	}

	w.AppendString("\n")
}
