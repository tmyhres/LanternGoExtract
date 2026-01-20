package wld

import (
	"strings"

	"github.com/tmyhres/LanternGoExtract/pkg/archive"
	"github.com/tmyhres/LanternGoExtract/pkg/infrastructure/logger"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
)

// WldFileEquipment represents an equipment WLD file containing general models.
type WldFileEquipment struct {
	*BaseWldFile
}

// NewWldFileEquipment creates a new equipment WLD file handler.
func NewWldFileEquipment(wldData archive.File, zoneName string, wldType WldType, log logger.Logger, settings *Settings, wldToInject WldFile) *WldFileEquipment {
	return &WldFileEquipment{
		BaseWldFile: NewBaseWldFile(wldData, zoneName, wldType, log, settings, wldToInject),
	}
}

// ProcessData processes the equipment WLD data.
func (w *WldFileEquipment) ProcessData() {
	w.BaseWldFile.ProcessData()
	w.findUnhandledSkeletons()
	w.findAdditionalAnimations()
}

// ExportData exports the equipment WLD data.
func (w *WldFileEquipment) ExportData() {
	w.BaseWldFile.ExportData()
	w.exportParticleSystems()
}

// exportParticleSystems exports particle system data.
func (w *WldFileEquipment) exportParticleSystems() {
	particles := GetFragmentsByType[*fragments.ParticleCloud](w)

	for _, particle := range particles {
		// Export particle data
		// In the original C#, this creates a ParticleSystemWriter
		_ = particle
	}
}

// findUnhandledSkeletons finds and assigns unhandled skeletons to actors.
func (w *WldFileEquipment) findUnhandledSkeletons() {
	skeletons := GetFragmentsByType[*fragments.SkeletonHierarchy](w)

	if skeletons == nil {
		return
	}

	for _, skeleton := range skeletons {
		if skeleton.IsAssigned {
			continue
		}

		cleanedName := cleanFragmentNameKeepCase(skeleton.GetName())
		actorName := cleanedName + "_ACTORDEF"

		frag := w.GetFragmentByName(actorName)
		if frag == nil {
			continue
		}

		if actor, ok := frag.(*fragments.Actor); ok && actor != nil {
			actor.AssignSkeletonReference(skeleton)
		}
	}
}

// cleanFragmentNameKeepCase cleans a fragment name but preserves case.
func cleanFragmentNameKeepCase(name string) string {
	if name == "" {
		return ""
	}

	// Find the last underscore and remove the type suffix
	lastUnderscore := strings.LastIndex(name, "_")
	if lastUnderscore > 0 {
		return name[:lastUnderscore]
	}

	return name
}

// findAdditionalAnimations finds and assigns additional animations to skeletons.
func (w *WldFileEquipment) findAdditionalAnimations() {
	animations := GetFragmentsByType[*fragments.TrackFragment](w)
	skeletons := GetFragmentsByType[*fragments.SkeletonHierarchy](w)

	for _, track := range animations {
		if track == nil {
			continue
		}

		if track.IsPoseAnimation {
			continue
		}

		if track.IsProcessed {
			continue
		}

		for _, skeleton := range skeletons {
			cleanedTrackName := cleanFragmentName(track.GetName())

			valid, boneName := skeleton.IsValidSkeleton(cleanedTrackName)
			if valid {
				w.Logger.LogError("Assigning " + track.GetName() + " to " + skeleton.GetName())
				track.IsProcessed = true
				skeleton.AddTrackEquipment(track, strings.ToLower(boneName))
			}
		}
	}

	// Log unassigned tracks
	for _, track := range animations {
		if track.IsPoseAnimation || track.IsProcessed {
			continue
		}
		w.Logger.LogError("WldFileCharacters: Track not assigned: " + track.GetName())
	}
}

// Ensure WldFileEquipment implements WldFile interface.
var _ WldFile = (*WldFileEquipment)(nil)
