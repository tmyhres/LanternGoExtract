// Package config provides configuration management for the Lantern extractor.
package config

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/tmyhres/LanternGoExtract/pkg/infrastructure"
	"github.com/tmyhres/LanternGoExtract/pkg/infrastructure/logger"
	"github.com/tmyhres/LanternGoExtract/pkg/wld"
)

// Settings holds the configuration options for the Lantern extractor.
type Settings struct {
	// settingsFilePath is the OS path to the settings file.
	settingsFilePath string

	// logger is the logger reference for debug output.
	logger logger.Logger

	// EverQuestDirectory is the OS path to the EverQuest directory.
	EverQuestDirectory string

	// RawS3dExtract determines whether to extract data from the WLD file.
	// If false, only S3D contents are extracted.
	RawS3dExtract bool

	// ExportZoneMeshGroups adds group separation in the zone mesh export.
	ExportZoneMeshGroups bool

	// ExportHiddenGeometry exports hidden geometry like zone boundaries.
	ExportHiddenGeometry bool

	// ModelExportFormat sets the desired model export format.
	ModelExportFormat wld.ModelExportFormat

	// ExportCharactersToSingleFolder exports all characters to one folder.
	ExportCharactersToSingleFolder bool

	// ExportEquipmentToSingleFolder exports all equipment to one folder.
	ExportEquipmentToSingleFolder bool

	// ExportSoundsToSingleFolder exports all sounds to one folder.
	ExportSoundsToSingleFolder bool

	// ExportAllAnimationFrames exports all OBJ frames for all animations.
	ExportAllAnimationFrames bool

	// ExportZoneWithObjects exports zones with their objects.
	ExportZoneWithObjects bool

	// ExportGltfVertexColors exports vertex colors with glTF models.
	// Default behavior of glTF renderers is to mix the vertex color with the
	// base color, which will not look right. Only enable this if you intend
	// to do post-processing that requires vertex colors.
	ExportGltfVertexColors bool

	// ExportGltfInGlbFormat exports glTF models in GLB file format.
	// GLB packages the glTF JSON, the associated .bin, and all textures
	// into one file. This takes more space but makes models more portable.
	ExportGltfInGlbFormat bool

	// ClientDataToCopy specifies additional files to copy when extracting
	// with "all" or "clientdata".
	ClientDataToCopy string

	// CopyMusic enables copying XMI files to the 'Exports/Music' folder.
	CopyMusic bool

	// LoggerVerbosity sets the verbosity level of the logger.
	LoggerVerbosity int
}

// NewSettings creates a new Settings instance with default values.
func NewSettings(settingsFilePath string, log logger.Logger) *Settings {
	return &Settings{
		settingsFilePath:     settingsFilePath,
		logger:               log,
		EverQuestDirectory:   "/opt/EverQuest/",
		RawS3dExtract:        false,
		ExportZoneMeshGroups: false,
		ExportHiddenGeometry: false,
		LoggerVerbosity:      0,
	}
}

// Initialize loads settings from the settings file.
func (s *Settings) Initialize() error {
	data, err := os.ReadFile(s.settingsFilePath)
	if err != nil {
		s.logger.LogError("Error loading settings file: " + err.Error())
		return err
	}

	settingsText := string(data)
	parsedSettings := infrastructure.ParseTextToDictionary(settingsText, '=', '#')
	if parsedSettings == nil {
		return nil
	}

	if val, ok := parsedSettings["EverQuestDirectory"]; ok {
		s.EverQuestDirectory = val
		// Ensure the path ends with a separator
		s.EverQuestDirectory = filepath.Clean(s.EverQuestDirectory) + string(filepath.Separator)
	}

	if val, ok := parsedSettings["RawS3DExtract"]; ok {
		s.RawS3dExtract = parseBool(val)
	}

	if val, ok := parsedSettings["ExportZoneMeshGroups"]; ok {
		s.ExportZoneMeshGroups = parseBool(val)
	}

	if val, ok := parsedSettings["ExportHiddenGeometry"]; ok {
		s.ExportHiddenGeometry = parseBool(val)
	}

	if val, ok := parsedSettings["ExportZoneWithObjects"]; ok {
		s.ExportZoneWithObjects = parseBool(val)
	}

	if val, ok := parsedSettings["ModelExportFormat"]; ok {
		if intVal, err := strconv.Atoi(val); err == nil {
			s.ModelExportFormat = wld.ModelExportFormat(intVal)
		}
	}

	if val, ok := parsedSettings["ExportCharacterToSingleFolder"]; ok {
		s.ExportCharactersToSingleFolder = parseBool(val)
	}

	if val, ok := parsedSettings["ExportEquipmentToSingleFolder"]; ok {
		s.ExportEquipmentToSingleFolder = parseBool(val)
	}

	if val, ok := parsedSettings["ExportSoundsToSingleFolder"]; ok {
		s.ExportSoundsToSingleFolder = parseBool(val)
	}

	if val, ok := parsedSettings["ExportAllAnimationFrames"]; ok {
		s.ExportAllAnimationFrames = parseBool(val)
	}

	if val, ok := parsedSettings["ExportGltfVertexColors"]; ok {
		s.ExportGltfVertexColors = parseBool(val)
	}

	if val, ok := parsedSettings["ExportGltfInGlbFormat"]; ok {
		s.ExportGltfInGlbFormat = parseBool(val)
	}

	if val, ok := parsedSettings["ClientDataToCopy"]; ok {
		s.ClientDataToCopy = val
	}

	if val, ok := parsedSettings["CopyMusic"]; ok {
		s.CopyMusic = parseBool(val)
	}

	if val, ok := parsedSettings["LoggerVerbosity"]; ok {
		if intVal, err := strconv.Atoi(val); err == nil {
			s.LoggerVerbosity = intVal
		}
	}

	return nil
}

// parseBool converts a string to a boolean value.
// Accepts "true", "True", "TRUE", "1" as true values.
func parseBool(s string) bool {
	switch s {
	case "true", "True", "TRUE", "1":
		return true
	default:
		return false
	}
}
