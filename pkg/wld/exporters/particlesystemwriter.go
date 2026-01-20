package exporters

import (
	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/helpers"
)

// ParticleSystemWriter exports particle system data to a text format.
type ParticleSystemWriter struct {
	TextAssetWriter
}

// NewParticleSystemWriter creates a new ParticleSystemWriter.
func NewParticleSystemWriter() *ParticleSystemWriter {
	w := &ParticleSystemWriter{}
	w.AppendLine(ExportHeaderTitle + "Particle System")
	return w
}

// AddFragmentData adds particle cloud fragment data to the export buffer.
func (w *ParticleSystemWriter) AddFragmentData(data fragments.Fragment) {
	particleCloud, ok := data.(*fragments.ParticleCloud)
	if !ok || particleCloud == nil {
		return
	}

	w.AppendLine(helpers.CleanName(particleCloud.GetName(), "ParticleCloud", true))
}
