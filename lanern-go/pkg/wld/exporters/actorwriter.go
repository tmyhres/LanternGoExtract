package exporters

import (
	"fmt"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/datatypes"
	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/fragments"
	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/helpers"
)

// ActorWriter exports actor data to a text format.
type ActorWriter struct {
	TextAssetWriter
	actorType  datatypes.ActorType
	actorCount int
}

// NewActorWriter creates a new ActorWriter for the specified actor type.
func NewActorWriter(actorType datatypes.ActorType) *ActorWriter {
	return &ActorWriter{
		actorType: actorType,
	}
}

// AddFragmentData adds actor fragment data to the export buffer.
func (w *ActorWriter) AddFragmentData(data fragments.Fragment) {
	actor, ok := data.(*fragments.Actor)
	if !ok || actor == nil {
		return
	}

	if actor.ActorType != w.actorType {
		return
	}

	switch w.actorType {
	case datatypes.ActorTypeSkeletal:
		w.AppendString(helpers.CleanName(actor.GetName(), "Actor", true))
		w.AppendString(",")
		if actor.SkeletonReference != nil && actor.SkeletonReference.SkeletonHierarchy != nil {
			w.AppendString(helpers.CleanName(actor.SkeletonReference.SkeletonHierarchy.GetName(), "SkeletonHierarchy", true))
		}
		w.AppendString("\n")

	case datatypes.ActorTypeStatic:
		w.AppendString(helpers.CleanName(actor.GetName(), "Actor", true))
		w.AppendString(",")

		if meshRef, ok := actor.MeshReference.(*fragments.Mesh); ok && meshRef != nil {
			w.AppendString(helpers.CleanName(meshRef.GetName(), "Mesh", true))
		} else if legacyMeshRef, ok := actor.MeshReference.(*fragments.LegacyMesh); ok && legacyMeshRef != nil {
			w.AppendString(helpers.CleanName(legacyMeshRef.GetName(), "LegacyMesh", true))
		}

		w.AppendString("\n")

	default:
		w.AppendLine(helpers.CleanName(actor.GetName(), "Actor", true))
	}

	w.actorCount++
}

// WriteAssetToFile writes the actor data to a file with a header.
func (w *ActorWriter) WriteAssetToFile(fileName string) error {
	if w.GetExportByteCount() == 0 {
		return nil
	}

	// Prepend header
	header := fmt.Sprintf("%sActor List\n# Total models: %d\n", ExportHeaderTitle, w.actorCount)
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
