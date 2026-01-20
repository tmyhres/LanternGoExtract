// Package eq provides EverQuest file handling utilities for the Lantern extractor.
package eq

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GetValidEqFilePaths returns a sorted list of valid EQ archive file paths
// based on the archive name pattern. Supports special keywords like "all",
// "zones", "characters", "equipment", and "sounds".
func GetValidEqFilePaths(directory, archiveName string) []string {
	archiveName = strings.ToLower(archiveName)

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return []string{}
	}

	eqFiles := getAllFiles(directory)
	var validFiles []string

	switch archiveName {
	case "all":
		validFiles = getAllValidFiles(eqFiles)
	case "zones":
		validFiles = getValidZoneFiles(eqFiles)
	case "characters":
		validFiles = getValidCharacterFiles(eqFiles)
	case "equipment":
		validFiles = getValidEquipmentFiles(eqFiles)
	case "sounds":
		validFiles = getValidSoundFiles(eqFiles)
	default:
		validFiles = getValidFilesForShortname(archiveName, directory)
	}

	sort.Strings(validFiles)
	return validFiles
}

// getAllFiles recursively finds all files in the given directory.
func getAllFiles(directory string) []string {
	var files []string
	filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files
}

// getValidEquipmentFiles filters for equipment archive files.
func getValidEquipmentFiles(eqFiles []string) []string {
	var result []string
	for _, f := range eqFiles {
		if IsEquipmentArchive(filepath.Base(f)) {
			result = append(result, f)
		}
	}
	return result
}

// getAllValidFiles filters for all valid archive files.
func getAllValidFiles(eqFiles []string) []string {
	var result []string
	for _, f := range eqFiles {
		if IsValidArchive(filepath.Base(f)) {
			result = append(result, f)
		}
	}
	return result
}

// getValidZoneFiles filters for zone archive files.
func getValidZoneFiles(eqFiles []string) []string {
	var result []string
	for _, f := range eqFiles {
		if isZoneArchive(filepath.Base(f)) {
			result = append(result, f)
		}
	}
	return result
}

// getValidCharacterFiles filters for character archive files.
func getValidCharacterFiles(eqFiles []string) []string {
	var result []string
	for _, f := range eqFiles {
		if IsCharacterArchive(filepath.Base(f)) {
			result = append(result, f)
		}
	}
	return result
}

// getValidSoundFiles filters for sound archive files.
func getValidSoundFiles(eqFiles []string) []string {
	var result []string
	for _, f := range eqFiles {
		if IsSoundArchive(filepath.Base(f)) {
			result = append(result, f)
		}
	}
	return result
}

// getValidFilesForShortname finds all archive files matching a specific archive name or shortname.
func getValidFilesForShortname(archiveName, directory string) []string {
	var validFiles []string

	// If the archive name already has an extension, try to find it directly
	if strings.HasSuffix(archiveName, ".s3d") ||
		strings.HasSuffix(archiveName, ".pfs") ||
		strings.HasSuffix(archiveName, ".t3d") {
		archivePath := filepath.Join(directory, archiveName)
		if _, err := os.Stat(archivePath); err == nil {
			validFiles = append(validFiles, archivePath)
		}
		return validFiles
	}

	// Try PFS format first
	archivePath := filepath.Join(directory, archiveName+PfsFormatExtension)
	if _, err := os.Stat(archivePath); err == nil {
		validFiles = append(validFiles, archivePath)
		return validFiles
	}

	// Determine archive extension (S3D or T3D)
	archiveExtension := S3dFormatExtension
	if hasT3dFiles(directory) {
		archiveExtension = T3dFormatExtension
	}

	// Try to find all associated files with the shortname
	mainArchivePath := filepath.Join(directory, archiveName+archiveExtension)
	if _, err := os.Stat(mainArchivePath); err == nil {
		validFiles = append(validFiles, mainArchivePath)
	}

	// Some zones have additional object archives for things added past their initial release
	extensionObjectsArchivePath := filepath.Join(directory, archiveName+"_2_obj"+archiveExtension)
	if _, err := os.Stat(extensionObjectsArchivePath); err == nil {
		validFiles = append(validFiles, extensionObjectsArchivePath)
	}

	objectsArchivePath := filepath.Join(directory, archiveName+"_obj"+archiveExtension)
	if _, err := os.Stat(objectsArchivePath); err == nil {
		validFiles = append(validFiles, objectsArchivePath)
	}

	charactersArchivePath := filepath.Join(directory, archiveName+"_chr"+archiveExtension)
	if _, err := os.Stat(charactersArchivePath); err == nil {
		validFiles = append(validFiles, charactersArchivePath)
	}

	// Some zones have additional character archives for things added past their initial release
	// "qeynos" must be excluded because both qeynos and qeynos2 are used as shortnames
	extensionCharactersArchivePath := filepath.Join(directory, archiveName+"2_chr"+archiveExtension)
	if _, err := os.Stat(extensionCharactersArchivePath); err == nil && archiveName != "qeynos" {
		validFiles = append(validFiles, extensionCharactersArchivePath)
	}

	return validFiles
}

// hasT3dFiles checks if any T3D files exist in the directory.
func hasT3dFiles(directory string) bool {
	files, err := filepath.Glob(filepath.Join(directory, "*"+T3dFormatExtension))
	if err != nil {
		return false
	}
	return len(files) > 0
}

// ObjArchivePath returns the object archive path corresponding to a given archive path.
func ObjArchivePath(archivePath string) string {
	ext := filepath.Ext(archivePath)
	base := strings.TrimSuffix(archivePath, ext)
	return base + "_obj" + ext
}

// isZoneArchive checks if the archive is a zone archive (local helper).
func isZoneArchive(archiveName string) bool {
	return IsValidArchive(archiveName) &&
		!IsEquipmentArchive(archiveName) &&
		!IsSkyArchive(archiveName) &&
		!IsBitmapArchive(archiveName) &&
		!IsCharacterArchive(archiveName)
}

// IsValidArchive checks if the archive name indicates a valid archive file.
func IsValidArchive(archiveName string) bool {
	name := strings.ToLower(archiveName)

	// chequip contains broken/conflicting data.
	// _lit archives get injected later during archive extraction
	if strings.Contains(name, "chequip") || strings.HasSuffix(name, "_lit.s3d") {
		return false
	}

	return strings.HasSuffix(name, ".s3d") ||
		strings.HasSuffix(name, ".t3d") ||
		strings.HasSuffix(name, ".pfs")
}

// IsEquipmentArchive checks if the archive name indicates an equipment archive.
func IsEquipmentArchive(archiveName string) bool {
	return strings.HasPrefix(strings.ToLower(archiveName), "gequip")
}

// IsCharacterArchive checks if the archive name indicates a character archive.
func IsCharacterArchive(archiveName string) bool {
	name := strings.ToLower(archiveName)

	// chequip contains broken/conflicting data
	if strings.Contains(name, "chequip") {
		return false
	}

	return strings.Contains(name, "_chr") ||
		strings.HasPrefix(name, "chequip") ||
		strings.Contains(name, "_amr")
}

// IsObjectsArchive checks if the archive name indicates an objects archive.
func IsObjectsArchive(archiveName string) bool {
	return strings.Contains(strings.ToLower(archiveName), "_obj")
}

// IsSkyArchive checks if the archive name indicates a sky archive.
func IsSkyArchive(archiveName string) bool {
	return strings.ToLower(archiveName) == "sky"
}

// IsBitmapArchive checks if the archive name indicates a bitmap-only archive.
func IsBitmapArchive(archiveName string) bool {
	return strings.HasPrefix(strings.ToLower(archiveName), "bmpwad")
}

// IsSoundArchive checks if the archive name indicates a sound archive.
func IsSoundArchive(archiveName string) bool {
	return strings.HasPrefix(strings.ToLower(archiveName), "snd")
}

// IsClientDataFile checks if the archive name indicates client data.
func IsClientDataFile(archiveName string) bool {
	return strings.ToLower(archiveName) == "clientdata"
}

// IsSpecialCaseExtraction checks if the archive requires special handling.
func IsSpecialCaseExtraction(archiveName string) bool {
	name := strings.ToLower(archiveName)
	return name == "clientdata" || name == "music"
}

// IsUsedSoundArchive checks if the sound archive is used by the Trilogy client.
// Archives higher than snd9 are not used.
func IsUsedSoundArchive(archiveName string) bool {
	if !IsSoundArchive(archiveName) {
		return false
	}

	// Trilogy client does not use archives higher than snd9
	name := strings.ToLower(archiveName)
	if len(name) >= 5 {
		suffix := name[len(name)-2:]
		// Check if the last two characters are both digits (snd10, snd11, etc.)
		if len(suffix) == 2 && isDigit(suffix[0]) && isDigit(suffix[1]) {
			return false
		}
	}

	return true
}

// isDigit checks if a byte is a decimal digit.
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
