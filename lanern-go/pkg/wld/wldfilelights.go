package wld

import (
	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/archive"
	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/infrastructure/logger"
	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/fragments"
)

// WldFileLights represents a lights WLD file containing light instances.
type WldFileLights struct {
	*BaseWldFile
}

// NewWldFileLights creates a new lights WLD file handler.
func NewWldFileLights(wldData archive.File, zoneName string, wldType WldType, log logger.Logger, settings *Settings, wldToInject WldFile) *WldFileLights {
	return &WldFileLights{
		BaseWldFile: NewBaseWldFile(wldData, zoneName, wldType, log, settings, wldToInject),
	}
}

// ExportData exports the lights WLD data.
func (w *WldFileLights) ExportData() {
	w.exportLightInstanceList()
}

// exportLightInstanceList exports the list of light instances.
func (w *WldFileLights) exportLightInstanceList() {
	lightInstances := GetFragmentsByType[*fragments.LightInstance](w)

	if len(lightInstances) == 0 {
		w.Logger.LogWarning("Unable to export light instance list. No instances found.")
		return
	}

	// Export each light instance
	// In the original C#, this creates a LightInstancesWriter
	for _, light := range lightInstances {
		// Export light data (position, colors, radius)
		_ = light
	}

	// The actual file writing would be done here
	// Output file: light_instances.txt
}

// GetLightInstances returns all light instances.
func (w *WldFileLights) GetLightInstances() []*fragments.LightInstance {
	return GetFragmentsByType[*fragments.LightInstance](w)
}

// Ensure WldFileLights implements WldFile interface.
var _ WldFile = (*WldFileLights)(nil)
