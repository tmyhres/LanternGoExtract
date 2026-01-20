package wld

import (
	"github.com/tmyhres/LanternGoExtract/pkg/archive"
	"github.com/tmyhres/LanternGoExtract/pkg/infrastructure/logger"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/datatypes"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
)

// WldFileZone represents a zone WLD file containing main geometry and BSP tree.
type WldFileZone struct {
	*BaseWldFile

	// BasePath is the base path for file operations.
	BasePath string

	// RootFolder is the root folder for the zone.
	RootFolder string

	// ShortName is the zone's short name.
	ShortName string

	// BaseS3DArchive is the base S3D archive for the zone.
	BaseS3DArchive archive.Archive

	// WldFileToInject is an additional WLD file to inject data from.
	WldFileToInject WldFile
}

// NewWldFileZone creates a new zone WLD file handler.
func NewWldFileZone(wldData archive.File, zoneName string, wldType WldType, log logger.Logger, settings *Settings, wldToInject WldFile) *WldFileZone {
	return &WldFileZone{
		BaseWldFile: NewBaseWldFile(wldData, zoneName, wldType, log, settings, wldToInject),
	}
}

// ProcessData processes the zone WLD data.
func (w *WldFileZone) ProcessData() {
	w.BaseWldFile.ProcessData()
	w.linkBspReferences()

	if w.WldToInject != nil {
		w.importVertexColors()
	}

	if w.WldType == WldTypeObjects {
		w.fixSkeletalObjectCollision()
	}
}

// ExportData exports the zone WLD data.
func (w *WldFileZone) ExportData() {
	w.BaseWldFile.ExportData()
	w.exportAmbientLightColor()
	w.exportBspTree()
}

// fixSkeletalObjectCollision removes collision from animated bones in skeletal objects.
func (w *WldFileZone) fixSkeletalObjectCollision() {
	actors := GetFragmentsByType[*fragments.Actor](w)

	for _, actor := range actors {
		if actor.ActorType != datatypes.ActorTypeSkeletal {
			continue
		}

		if actor.SkeletonReference == nil || actor.SkeletonReference.SkeletonHierarchy == nil {
			continue
		}

		skeleton := actor.SkeletonReference.SkeletonHierarchy.Skeleton

		for _, bone := range skeleton {
			if bone.Track != nil && bone.Track.TrackDefFragment != nil {
				if len(bone.Track.TrackDefFragment.Frames) != 1 {
					// Clear collision on mesh if it has animation frames
					if meshRef, ok := bone.MeshReference.(*fragments.MeshReference); ok && meshRef != nil {
						if mesh, ok := meshRef.Mesh.(*fragments.Mesh); ok && mesh != nil {
							mesh.ClearCollision()
						}
					}
				}
			}
		}
	}
}

// importVertexColors imports vertex colors from the injected WLD file.
func (w *WldFileZone) importVertexColors() {
	if w.WldToInject == nil {
		return
	}

	colors := GetFragmentsByType[*fragments.VertexColors](w.WldToInject)
	if len(colors) == 0 {
		return
	}

	for _, vc := range colors {
		if vc.Name == "" {
			continue
		}

		// Extract mesh name from vertex color name
		parts := splitFirst(vc.Name, "_")
		meshName := parts[0] + "_DMSPRITEDEF"

		frag := w.GetFragmentByName(meshName)
		if mesh, ok := frag.(*fragments.Mesh); ok && mesh != nil {
			mesh.Colors = vc.Colors
		}
	}
}

// splitFirst splits a string on the first occurrence of sep.
func splitFirst(s, sep string) []string {
	for i := 0; i < len(s)-len(sep)+1; i++ {
		if s[i:i+len(sep)] == sep {
			return []string{s[:i], s[i+len(sep):]}
		}
	}
	return []string{s}
}

// linkBspReferences links BSP tree references to regions.
func (w *WldFileZone) linkBspReferences() {
	bspTrees := GetFragmentsByType[*fragments.BspTree](w)
	bspRegions := GetFragmentsByType[*fragments.BspRegion](w)
	regionTypes := GetFragmentsByType[*fragments.BspRegionType](w)

	if len(bspTrees) == 0 || len(bspRegions) == 0 || len(regionTypes) == 0 {
		return
	}

	bspTrees[0].LinkBspRegions(bspRegions)

	for _, regionType := range regionTypes {
		regionType.LinkRegionType(bspRegions)
	}
}

// exportAmbientLightColor exports the ambient light color data.
func (w *WldFileZone) exportAmbientLightColor() {
	ambientLights := GetFragmentsByType[*fragments.GlobalAmbientLight](w)

	if len(ambientLights) == 0 {
		return
	}

	// Export logic would go here
	// In the original C#, this creates an AmbientLightColorWriter
	// For now, this is a placeholder for the export functionality
	_ = ambientLights[0]
}

// exportBspTree exports the BSP tree data.
func (w *WldFileZone) exportBspTree() {
	bspTrees := GetFragmentsByType[*fragments.BspTree](w)

	if len(bspTrees) == 0 {
		return
	}

	// Export logic would go here
	// In the original C#, this creates a BspTreeWriter
	// For now, this is a placeholder for the export functionality
	_ = bspTrees[0]
}

// Ensure WldFileZone implements WldFile interface.
var _ WldFile = (*WldFileZone)(nil)
