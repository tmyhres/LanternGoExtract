package exporters

import (
	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/helpers"
)

// ActorWriterNew exports actor data in a new format.
type ActorWriterNew struct {
	TextAssetWriter
}

// NewActorWriterNew creates a new ActorWriterNew.
func NewActorWriterNew() *ActorWriterNew {
	return &ActorWriterNew{}
}

// AddFragmentData adds actor fragment data to the export buffer.
func (w *ActorWriterNew) AddFragmentData(data fragments.Fragment) {
	actor, ok := data.(*fragments.Actor)
	if !ok || actor == nil {
		return
	}

	w.AppendString(actorTypeToString(actor.ActorType))
	w.AppendString(",")
	w.AppendString(actor.ReferenceName)
	w.AppendLine(helpers.CleanName(actor.GetName(), "Actor", true))
}

// WriteAssetToFile writes the actor data to a file with a header.
func (w *ActorWriterNew) WriteAssetToFile(fileName string) error {
	if w.GetExportByteCount() == 0 {
		return nil
	}

	// Prepend header
	header := ExportHeaderTitle + "Actor\n"
	content := header + w.GetExport().String()

	// Temporarily replace export content
	originalExport := w.GetExport().String()
	w.ClearExportData()
	w.AppendString(content)

	err := w.TextAssetWriter.WriteAssetToFile(fileName)

	// Restore original content
	w.ClearExportData()
	w.AppendString(originalExport)

	return err
}
