package wld

import (
	"github.com/tmyhres/LanternGoExtract/pkg/archive"
	"github.com/tmyhres/LanternGoExtract/pkg/infrastructure/logger"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
)

// WldFileZoneObjects represents a zone objects WLD file containing object instance data.
type WldFileZoneObjects struct {
	*BaseWldFile
}

// NewWldFileZoneObjects creates a new zone objects WLD file handler.
func NewWldFileZoneObjects(wldData archive.File, zoneName string, wldType WldType, log logger.Logger, settings *Settings, wldToInject WldFile) *WldFileZoneObjects {
	return &WldFileZoneObjects{
		BaseWldFile: NewBaseWldFile(wldData, zoneName, wldType, log, settings, wldToInject),
	}
}

// ExportData exports the zone objects WLD data.
func (w *WldFileZoneObjects) ExportData() {
	w.exportObjectInstanceAndVertexColorList()
}

// exportObjectInstanceAndVertexColorList exports the object instance list and vertex colors.
func (w *WldFileZoneObjects) exportObjectInstanceAndVertexColorList() {
	instanceList := GetFragmentsByType[*fragments.ObjectInstance](w)

	if len(instanceList) == 0 {
		w.Logger.LogWarning("Cannot export object instance list. No object instances found.")
		return
	}

	// Process instances from this WLD
	for _, instance := range instanceList {
		// Export instance data
		_ = instance

		if instance.Colors == nil {
			continue
		}

		// Export vertex colors
		_ = instance.Colors
	}

	// Process instances from injected WLD if present
	if w.WldToInject != nil {
		injectedInstances := GetFragmentsByType[*fragments.ObjectInstance](w.WldToInject)

		for _, instance := range injectedInstances {
			// Export instance data
			_ = instance

			if instance.Colors == nil {
				continue
			}

			// Export vertex colors
			_ = instance.Colors
		}
	}

	// The actual file writing would be done here
	// In the original C#:
	// - ObjectInstanceWriter writes to object_instances.txt
	// - VertexColorsWriter writes individual vertex color files
}

// GetObjectInstances returns all object instances including from injected WLD.
func (w *WldFileZoneObjects) GetObjectInstances() []*fragments.ObjectInstance {
	instances := GetFragmentsByType[*fragments.ObjectInstance](w)

	if w.WldToInject != nil {
		injectedInstances := GetFragmentsByType[*fragments.ObjectInstance](w.WldToInject)
		instances = append(instances, injectedInstances...)
	}

	return instances
}

// Ensure WldFileZoneObjects implements WldFile interface.
var _ WldFile = (*WldFileZoneObjects)(nil)
