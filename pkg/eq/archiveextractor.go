package eq

import (
	"path/filepath"
	"strings"

	"github.com/tmyhres/LanternGoExtract/pkg/archive"
	"github.com/tmyhres/LanternGoExtract/pkg/infrastructure"
	"github.com/tmyhres/LanternGoExtract/pkg/infrastructure/logger"
	"github.com/tmyhres/LanternGoExtract/pkg/sound"
	"github.com/tmyhres/LanternGoExtract/pkg/wld"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/exporters"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/helpers"
)

// Settings contains configuration options for archive extraction.
type Settings struct {
	// EverQuestDirectory is the path to the EverQuest installation.
	EverQuestDirectory string

	// RawS3dExtract extracts raw S3D contents without WLD processing.
	RawS3dExtract bool

	// ExportSoundsToSingleFolder exports all sounds to a single "sounds" folder.
	ExportSoundsToSingleFolder bool

	// ExportCharactersToSingleFolder exports all characters to a single folder.
	ExportCharactersToSingleFolder bool

	// ExportEquipmentToSingleFolder exports all equipment to a single folder.
	ExportEquipmentToSingleFolder bool

	// ExportZoneWithObjects merges zone and object data during export.
	ExportZoneWithObjects bool

	// ModelExportFormat determines the output format for 3D models.
	ModelExportFormat wld.ModelExportFormat
}

// Extract extracts and processes an archive file.
//
// The extraction process varies based on archive type:
//   - Zone archives: Extract zone geometry, lights, and object placements
//   - Character archives: Extract character models and animations
//   - Equipment archives: Extract equipment and item models
//   - Object archives: Extract zone object models
//   - Sky archives: Extract sky geometry
//   - Sound archives: Extract sound files
//   - Texture archives: Extract textures only
//
// Parameters:
//   - path: Full path to the archive file
//   - rootFolder: Root folder for exports
//   - log: Logger for status messages
//   - settings: Configuration options
//
// Returns an error if extraction fails.
func Extract(path, rootFolder string, log logger.Logger, settings *Settings) error {
	archiveName := getFileNameWithoutExtension(path)
	if archiveName == "" {
		return nil
	}

	arc, err := archive.GetArchive(path, log)
	if err != nil {
		log.LogError("LanternExtractor: Failed to get archive at path: " + path)
		return err
	}

	shortName := strings.Split(archiveName, "_")[0]

	if err := arc.Initialize(); err != nil {
		log.LogError("LanternExtractor: Failed to initialize archive at path: " + path)
		return err
	}

	// Raw S3D extraction mode - just dump all files
	if settings.RawS3dExtract {
		return arc.WriteAllFiles(filepath.Join(rootFolder, archiveName))
	}

	// For non-WLD archives, extract textures and sounds only
	if !arc.IsWldArchive() {
		writeS3dTextures(arc, filepath.Join(rootFolder, shortName), log)

		if IsUsedSoundArchive(archiveName) {
			soundPath := filepath.Join(rootFolder, shortName)
			if settings.ExportSoundsToSingleFolder {
				soundPath = filepath.Join(rootFolder, "sounds")
			}
			writeS3dSounds(arc, soundPath, log)
		}

		return nil
	}

	// Process WLD-containing archives
	wldFileName := archiveName + WldFormatExtension
	wldFileInArchive := arc.GetFile(wldFileName)
	if wldFileInArchive == nil {
		log.LogError("Unable to extract WLD file " + wldFileName + " from archive: " + path)
		return nil
	}

	// Route to appropriate handler based on archive type
	switch {
	case IsEquipmentArchive(archiveName):
		extractArchiveEquipment(rootFolder, log, settings, wldFileInArchive, shortName, arc)

	case IsSkyArchive(archiveName):
		extractArchiveSky(rootFolder, log, settings, wldFileInArchive, shortName, arc)

	case IsCharacterArchive(archiveName):
		extractArchiveCharacters(path, rootFolder, log, settings, archiveName, wldFileInArchive, shortName, arc)

	case IsObjectsArchive(archiveName):
		extractArchiveObjects(path, rootFolder, log, settings, wldFileInArchive, shortName, arc)

	default:
		extractArchiveZone(path, rootFolder, log, settings, shortName, wldFileInArchive, arc)
	}

	// Fix missing textures post-extraction
	helpers.FixMissingTextures(archiveName)

	return nil
}

// extractArchiveZone extracts a zone archive with its associated data.
func extractArchiveZone(path, rootFolder string, log logger.Logger, settings *Settings, shortName string, wldFileInArchive archive.File, arc archive.Archive) {
	// Some Kunark zones have a "_lit" archive for additional lighting data
	litPath := strings.Replace(path, shortName+".s3d", shortName+"_lit.s3d", 1)
	litPath = strings.Replace(litPath, shortName+".t3d", shortName+"_lit.t3d", 1)

	arcLit, _ := archive.GetArchive(litPath, log)
	var wldFileLit *wld.WldFileZone

	if arcLit != nil {
		if err := arcLit.Initialize(); err == nil {
			litWldFileInArchive := arcLit.GetFile(shortName + "_lit.wld")
			if litWldFileInArchive != nil {
				wldSettings := convertToWldSettings(settings)
				wldFileLit = wld.NewWldFileZone(litWldFileInArchive, shortName, wld.WldTypeZone, log, wldSettings, nil)
				wldFileLit.Initialize(rootFolder, false)

				// Also check for lights WLD in the lit archive
				litLightsFile := arcLit.GetFile(shortName + "_lit.wld")
				if litLightsFile != nil {
					lightsWldFile := wld.NewWldFileLights(litLightsFile, shortName, wld.WldTypeLights, log, wldSettings, wldFileLit)
					lightsWldFile.Initialize(rootFolder, true)
				}
			}
		}
	}

	// Create main zone WLD file
	wldSettings := convertToWldSettings(settings)
	wldFile := wld.NewWldFileZone(wldFileInArchive, shortName, wld.WldTypeZone, log, wldSettings, wldFileLit)

	// If merging zone with objects, inject additional data
	if settings.ExportZoneWithObjects {
		wldFile.BasePath = path
		wldFile.BaseS3DArchive = arc
		wldFile.WldFileToInject = wldFileLit
		wldFile.RootFolder = rootFolder
		wldFile.ShortName = shortName
	}

	texturePath := filepath.Join(rootFolder, shortName, "Zone", "Textures")
	initializeWldAndWriteTextures(wldFile, rootFolder, texturePath, arc, settings, log)

	// Process lights WLD
	lightsFileInArchive := arc.GetFile("lights" + WldFormatExtension)
	if lightsFileInArchive != nil {
		lightsWldFile := wld.NewWldFileLights(lightsFileInArchive, shortName, wld.WldTypeLights, log, wldSettings, wldFileLit)
		lightsWldFile.Initialize(rootFolder, true)
	}

	// Process zone objects WLD
	zoneObjectsFileInArchive := arc.GetFile("objects" + WldFormatExtension)
	if zoneObjectsFileInArchive != nil {
		zoneObjectsWldFile := wld.NewWldFileZoneObjects(zoneObjectsFileInArchive, shortName, wld.WldTypeZoneObjects, log, wldSettings, wldFileLit)
		zoneObjectsWldFile.Initialize(rootFolder, true)
	}

	// Extract sound data
	extractSoundData(shortName, rootFolder, log, settings)
}

// extractArchiveObjects extracts an objects archive.
func extractArchiveObjects(path, rootFolder string, log logger.Logger, settings *Settings, wldFileInArchive archive.File, shortName string, arc archive.Archive) {
	// Some zones have a "_2_obj" archive for additional objects
	obj2Path := strings.Replace(path, shortName+"_obj", shortName+"_2_obj", 1)
	arcObj2, _ := archive.GetArchive(obj2Path, log)
	var wldFileObj2 *wld.WldFileZone

	if arcObj2 != nil {
		if err := arcObj2.Initialize(); err == nil {
			obj2WldFileInArchive := arcObj2.GetFile(shortName + "_2_obj.wld")
			if obj2WldFileInArchive != nil {
				wldSettings := convertToWldSettings(settings)
				wldFileObj2 = wld.NewWldFileZone(obj2WldFileInArchive, shortName, wld.WldTypeZone, log, wldSettings, nil)
				wldFileObj2.Initialize(rootFolder, false)
			}
		}
	}

	wldSettings := convertToWldSettings(settings)
	wldFile := wld.NewWldFileZone(wldFileInArchive, shortName, wld.WldTypeObjects, log, wldSettings, wldFileObj2)

	correctShortname := GetCorrectZoneShortname(shortName)
	texturePath := filepath.Join(rootFolder, correctShortname, "Objects", "Textures")
	initializeWldAndWriteTextures(wldFile, rootFolder, texturePath, arc, settings, log)
}

// extractArchiveCharacters extracts a characters archive.
func extractArchiveCharacters(path, rootFolder string, log logger.Logger, settings *Settings, archiveName string, wldFileInArchive archive.File, shortName string, arc archive.Archive) {
	var wldFileToInject wld.WldFile

	// global3_chr contains only animations and needs global_chr data
	if strings.HasPrefix(archiveName, "global3_chr") {
		globalPath := strings.Replace(path, "global3_chr", "global_chr", 1)
		arc2, err := archive.GetArchive(globalPath, log)
		if err != nil {
			log.LogError("Failed to get global_chr archive at path: " + globalPath)
			return
		}

		if err := arc2.Initialize(); err != nil {
			log.LogError("Failed to initialize archive at path: " + globalPath)
			return
		}

		wldFileInArchive2 := arc2.GetFile("global_chr.wld")
		if wldFileInArchive2 != nil {
			wldSettings := convertToWldSettings(settings)
			wldToInject := wld.NewWldFileCharacters(wldFileInArchive2, "global_chr", wld.WldTypeCharacters, log, wldSettings, nil)
			wldToInject.Initialize(rootFolder, false)
			wldFileToInject = wldToInject
		}
	}

	wldSettings := convertToWldSettings(settings)
	wldFile := wld.NewWldFileCharacters(wldFileInArchive, shortName, wld.WldTypeCharacters, log, wldSettings, wldFileToInject)

	// Determine export path
	var exportPath string
	if settings.ExportCharactersToSingleFolder && settings.ModelExportFormat == wld.ModelExportFormatIntermediate {
		exportPath = filepath.Join(rootFolder, "characters", "Textures")
	} else {
		correctShortname := GetCorrectZoneShortname(shortName)
		exportPath = filepath.Join(rootFolder, correctShortname, "Characters", "Textures")
	}

	initializeWldAndWriteTextures(wldFile, rootFolder, exportPath, arc, settings, log)
}

// extractArchiveSky extracts a sky archive.
func extractArchiveSky(rootFolder string, log logger.Logger, settings *Settings, wldFileInArchive archive.File, shortName string, arc archive.Archive) {
	wldSettings := convertToWldSettings(settings)
	wldFile := wld.NewWldFileZone(wldFileInArchive, shortName, wld.WldTypeSky, log, wldSettings, nil)

	texturePath := filepath.Join(rootFolder, shortName, "Textures")
	initializeWldAndWriteTextures(wldFile, rootFolder, texturePath, arc, settings, log)
}

// extractArchiveEquipment extracts an equipment archive.
func extractArchiveEquipment(rootFolder string, log logger.Logger, settings *Settings, wldFileInArchive archive.File, shortName string, arc archive.Archive) {
	wldSettings := convertToWldSettings(settings)
	wldFile := wld.NewWldFileEquipment(wldFileInArchive, shortName, wld.WldTypeEquipment, log, wldSettings, nil)

	// Determine export path
	var exportPath string
	if settings.ExportEquipmentToSingleFolder && settings.ModelExportFormat == wld.ModelExportFormatIntermediate {
		exportPath = filepath.Join(rootFolder, "equipment", "Textures")
	} else {
		exportPath = filepath.Join(rootFolder, shortName, "Textures")
	}

	initializeWldAndWriteTextures(wldFile, rootFolder, exportPath, arc, settings, log)
}

// initializeWldAndWriteTextures initializes a WLD file and writes its textures.
func initializeWldAndWriteTextures(wldFile wld.WldFile, rootFolder, texturePath string, arc archive.Archive, settings *Settings, log logger.Logger) {
	if settings.ModelExportFormat != wld.ModelExportFormatGltf {
		// Standard flow: initialize, then write textures
		wldFile.Initialize(rootFolder, true)
		writeWldTextures(arc, wldFile, texturePath, log)
	} else {
		// glTF export requires textures to be present first
		wldFile.Initialize(rootFolder, false)
		writeWldTextures(arc, wldFile, texturePath, log)
		exportWldToGltf(wldFile, settings, log)
	}
}

// exportWldToGltf exports a WLD file to glTF format.
func exportWldToGltf(wldFile wld.WldFile, settings *Settings, log logger.Logger) {
	actors := wldFile.GetActors()
	meshes := wldFile.GetMeshes()
	materialLists := wldFile.GetMaterialLists()

	exportFolder := wldFile.GetExportFolderForWldType()
	zoneName := wldFile.GetZoneName()
	wldType := wldFile.GetWldType()

	exportSettings := &exporters.ActorExportSettings{
		ExportGltfInGlbFormat:          false,
		ExportGltfVertexColors:         false,
		ExportZoneWithObjects:          settings.ExportZoneWithObjects,
		ExportAllAnimationFrames:       false,
		ExportCharactersToSingleFolder: settings.ExportCharactersToSingleFolder,
	}

	if err := exporters.ExportActorsToGltf(actors, meshes, materialLists, wldType, zoneName, exportFolder, exportSettings); err != nil {
		log.LogError("Failed to export to glTF: " + err.Error())
	}
}

// writeS3dSounds writes sound files from an archive to disk.
func writeS3dSounds(arc archive.Archive, filePath string, log logger.Logger) {
	allFiles := arc.GetAllFiles()

	for _, file := range allFiles {
		if strings.HasSuffix(strings.ToLower(file.GetName()), ".wav") {
			infrastructure.WriteSoundAsWav(file.GetBytes(), filePath, file.GetName(), log)
		}
	}
}

// writeS3dTextures writes texture files from an archive to disk as PNG.
func writeS3dTextures(arc archive.Archive, filePath string, log logger.Logger) {
	allFiles := arc.GetAllFiles()

	for _, file := range allFiles {
		name := strings.ToLower(file.GetName())
		if strings.HasSuffix(name, ".bmp") || strings.HasSuffix(name, ".dds") {
			infrastructure.WriteImageAsPng(file.GetBytes(), filePath, file.GetName(), false, log)
		}
	}
}

// WriteWldTextures writes textures referenced by a WLD file to disk.
// This is exported for use by other packages that need to write WLD textures.
func WriteWldTextures(arc archive.Archive, wldFile wld.WldFile, zoneName string, log logger.Logger) {
	writeWldTextures(arc, wldFile, zoneName, log)
}

// writeWldTextures writes textures referenced by a WLD file to disk.
func writeWldTextures(arc archive.Archive, wldFile wld.WldFile, zoneName string, log logger.Logger) {
	allBitmaps := wldFile.GetAllBitmapNames()
	maskedBitmaps := getMaskedBitmaps(wldFile)

	filenameChanges := wldFile.GetFilenameChanges()

	for _, bitmap := range allBitmaps {
		filename := bitmap

		// Check for filename changes
		if filenameChanges != nil {
			baseName := getFileNameWithoutExtension(bitmap)
			if newName, ok := filenameChanges[baseName]; ok {
				filename = newName + ".bmp"
			}
		}

		pfsFile := arc.GetFile(filename)
		if pfsFile == nil {
			continue
		}

		isMasked := maskedBitmaps != nil && containsString(maskedBitmaps, bitmap)
		infrastructure.WriteImageAsPng(pfsFile.GetBytes(), zoneName, bitmap, isMasked, log)
	}
}

// extractSoundData extracts sound data for a zone.
func extractSoundData(shortName, rootFolder string, log logger.Logger, settings *Settings) {
	envAudio := sound.NewEnvAudio()

	// Try to load defaults.dat first, then defaults.eal
	ealFilePath := filepath.Join(settings.EverQuestDirectory, "defaults.dat")
	loaded, _ := envAudio.Load(ealFilePath)
	if !loaded {
		ealFilePath = filepath.Join(settings.EverQuestDirectory, "defaults.eal")
		envAudio.Load(ealFilePath)
	}

	// Load sound bank
	soundBankPath := settings.EverQuestDirectory + shortName + "_sndbnk" + SoundFormatExtension
	sounds := sound.NewEffSndBnk(soundBankPath)
	sounds.Initialize()

	// Load sound entries
	soundEntriesPath := settings.EverQuestDirectory + shortName + "_sounds" + SoundFormatExtension
	soundEntries := sound.NewEffSounds(soundEntriesPath, sounds, envAudio)
	soundEntries.Initialize(log)
	soundEntries.ExportSoundData(shortName, rootFolder)
}

// convertToWldSettings converts extractor settings to WLD settings.
func convertToWldSettings(settings *Settings) *wld.Settings {
	if settings == nil {
		return nil
	}
	return &wld.Settings{
		ModelExportFormat:              settings.ModelExportFormat,
		ExportCharactersToSingleFolder: settings.ExportCharactersToSingleFolder,
		ExportEquipmentToSingleFolder:  settings.ExportEquipmentToSingleFolder,
	}
}

// Helper functions

// getFileNameWithoutExtension returns the filename without its extension.
func getFileNameWithoutExtension(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// getMaskedBitmaps returns the list of masked bitmap names from a WLD file.
// This is a placeholder - the actual implementation depends on the WLD file interface.
func getMaskedBitmaps(wldFile wld.WldFile) []string {
	// The WldFile interface would need a GetMaskedBitmaps method
	// For now, return nil as a placeholder
	return nil
}

// containsString checks if a string slice contains a specific string.
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
