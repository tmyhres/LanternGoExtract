package exporters

import (
	"fmt"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/fragments"
)

// ObjectInstanceWriter exports object instance data to a text format.
type ObjectInstanceWriter struct {
	TextAssetWriter
}

// NewObjectInstanceWriter creates a new ObjectInstanceWriter.
func NewObjectInstanceWriter() *ObjectInstanceWriter {
	w := &ObjectInstanceWriter{}
	w.AppendLine(ExportHeaderTitle + "Object Instances")
	w.AppendLine(ExportHeaderFormat + "ModelName, PosX, PosY, PosZ, RotX, RotY, RotZ, ScaleX, ScaleY, ScaleZ, ColorIndex")
	return w
}

// AddFragmentData adds object instance fragment data to the export buffer.
func (w *ObjectInstanceWriter) AddFragmentData(data fragments.Fragment) {
	instance, ok := data.(*fragments.ObjectInstance)
	if !ok || instance == nil {
		return
	}

	// Model name
	w.AppendString(instance.ObjectName)
	w.AppendString(",")

	// Position (note coordinate swap: x, z, y)
	w.AppendString(formatFloat(instance.Position.X))
	w.AppendString(",")
	w.AppendString(formatFloat(instance.Position.Z))
	w.AppendString(",")
	w.AppendString(formatFloat(instance.Position.Y))
	w.AppendString(",")

	// Rotation (note coordinate swap: x, z, y)
	w.AppendString(formatFloat(instance.Rotation.X))
	w.AppendString(",")
	w.AppendString(formatFloat(instance.Rotation.Z))
	w.AppendString(",")
	w.AppendString(formatFloat(instance.Rotation.Y))
	w.AppendString(",")

	// Scale
	w.AppendString(formatFloat(instance.Scale.X))
	w.AppendString(",")
	w.AppendString(formatFloat(instance.Scale.Y))
	w.AppendString(",")
	w.AppendString(formatFloat(instance.Scale.Z))
	w.AppendString(",")

	// Color index
	colorIndex := -1
	if instance.Colors != nil {
		if vertexColors, ok := instance.Colors.(*fragments.VertexColorsReference); ok && vertexColors != nil {
			colorIndex = vertexColors.GetIndex()
		} else {
			colorIndex = instance.Colors.GetIndex()
		}
	}
	w.AppendLine(fmt.Sprintf("%d", colorIndex))
}
