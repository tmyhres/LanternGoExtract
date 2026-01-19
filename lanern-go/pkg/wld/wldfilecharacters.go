package wld

import (
	"bufio"
	"os"
	"strings"
	"unicode"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/archive"
	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/infrastructure/logger"
	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/fragments"
)

// WldFileCharacters represents a characters WLD file containing character models and animations.
type WldFileCharacters struct {
	*BaseWldFile

	// animationSources maps model names to their animation source models.
	animationSources map[string]string
}

// NewWldFileCharacters creates a new characters WLD file handler.
func NewWldFileCharacters(wldData archive.File, zoneName string, wldType WldType, log logger.Logger, settings *Settings, wldToInject WldFile) *WldFileCharacters {
	w := &WldFileCharacters{
		BaseWldFile:      NewBaseWldFile(wldData, zoneName, wldType, log, settings, wldToInject),
		animationSources: make(map[string]string),
	}
	w.parseAnimationSources()
	return w
}

// parseAnimationSources loads the animation sources mapping from file.
func (w *WldFileCharacters) parseAnimationSources() {
	filename := "ClientData/animationsources.txt"

	file, err := os.Open(filename)
	if err != nil {
		w.Logger.LogError("WldFileCharacters: No animationsources.txt file found.")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(strings.ToLower(parts[1]))
		w.animationSources[key] = value
	}
}

// getAnimationModelLink returns the animation source model for a given model name.
func (w *WldFileCharacters) getAnimationModelLink(modelName string) string {
	if source, ok := w.animationSources[modelName]; ok {
		return source
	}
	return modelName
}

// ProcessData processes the characters WLD data.
func (w *WldFileCharacters) ProcessData() {
	w.BaseWldFile.ProcessData()
	w.findAdditionalAnimationsAndMeshes()
	w.buildSlotMapping()
	w.findMaterialVariants()

	if w.Settings != nil && w.Settings.ExportCharactersToSingleFolder {
		// Character fixer logic would go here
	}

	// Build skeleton data
	for _, skeleton := range GetFragmentsByType[*fragments.SkeletonHierarchy](w) {
		skeleton.BuildSkeletonData(w.WldType == WldTypeCharacters)
	}
}

// BuildSkeletonData builds skeleton data for all skeletons (called externally).
func (w *WldFileCharacters) BuildSkeletonData() {
	for _, skeleton := range GetFragmentsByType[*fragments.SkeletonHierarchy](w) {
		skeleton.BuildSkeletonData(true)
	}
}

// buildSlotMapping builds slot mappings for all material lists.
func (w *WldFileCharacters) buildSlotMapping() {
	materialLists := GetFragmentsByType[*fragments.MaterialList](w)

	for _, list := range materialLists {
		list.BuildSlotMapping()
	}
}

// findMaterialVariants finds and assigns material variants to material lists.
func (w *WldFileCharacters) findMaterialVariants() {
	materialLists := GetFragmentsByType[*fragments.MaterialList](w)
	materials := GetFragmentsByType[*fragments.Material](w)

	for _, list := range materialLists {
		materialListModelName := cleanFragmentName(list.GetName())

		for _, material := range materials {
			if material.IsHandled {
				continue
			}

			materialName := cleanFragmentName(material.GetName())

			if strings.HasPrefix(materialName, materialListModelName) {
				list.AddVariant(material)
			}
		}
	}

	// Log any unassigned materials
	for _, material := range materials {
		if material.IsHandled {
			continue
		}
		w.Logger.LogWarning("WldFileCharacters: Material not assigned: " + material.GetName())
	}
}

// cleanFragmentName cleans a fragment name by removing the type suffix.
func cleanFragmentName(name string) string {
	if name == "" {
		return ""
	}

	// Find the last underscore and check if what follows is a type identifier
	lastUnderscore := strings.LastIndex(name, "_")
	if lastUnderscore > 0 {
		return strings.ToLower(name[:lastUnderscore])
	}

	return strings.ToLower(name)
}

// findAdditionalAnimationsAndMeshes finds and assigns additional animations and meshes to skeletons.
func (w *WldFileCharacters) findAdditionalAnimationsAndMeshes() {
	tracks := GetFragmentsByType[*fragments.TrackFragment](w)
	if len(tracks) == 0 {
		return
	}

	skeletons := GetFragmentsByType[*fragments.SkeletonHierarchy](w)

	if len(skeletons) == 0 {
		if w.WldToInject == nil {
			return
		}
		skeletons = GetFragmentsByType[*fragments.SkeletonHierarchy](w.WldToInject)
	}

	if len(skeletons) == 0 {
		return
	}

	meshes := GetFragmentsByType[*fragments.Mesh](w)

	for _, skeleton := range skeletons {
		modelBase := skeleton.ModelBase
		alternateModel := w.getAnimationModelLink(modelBase)

		// Find and assign tracks to skeleton
		for _, track := range tracks {
			if track.IsPoseAnimation {
				continue
			}

			if !track.IsNameParsed {
				track.ParseTrackData()
			}

			trackModelBase := track.ModelName

			if trackModelBase != modelBase && alternateModel != trackModelBase {
				continue
			}

			// Add track to skeleton's animations
			skeleton.AddTrack(track)
		}

		// Find and assign additional meshes
		if len(meshes) > 0 {
			for _, mesh := range meshes {
				if mesh.IsHandled {
					continue
				}

				cleanedName := cleanFragmentName(mesh.GetName())
				basename := cleanedName

				// Check if name ends with a number (variant mesh)
				if len(cleanedName) > 0 && isDigit(cleanedName[len(cleanedName)-1]) {
					// Extract base name without variant number
					if len(cleanedName) >= 2 {
						cleanedName = cleanedName[:len(cleanedName)-2]
						if len(cleanedName) > 3 {
							cleanedName = cleanedName[:len(cleanedName)-2]
						}
						basename = cleanedName
					}
				}

				if basename == modelBase {
					// Add mesh to skeleton's secondary meshes
					skeleton.AddSecondaryMesh(mesh)
				}
			}
		}
	}

	// Log unassigned tracks
	for _, track := range tracks {
		if track.IsPoseAnimation || track.IsProcessed {
			continue
		}
		w.Logger.LogWarning("WldFileCharacters: Track not assigned: " + track.GetName())
	}
}

// isDigit checks if a byte is a decimal digit.
func isDigit(b byte) bool {
	return unicode.IsDigit(rune(b))
}

// Ensure WldFileCharacters implements WldFile interface.
var _ WldFile = (*WldFileCharacters)(nil)
