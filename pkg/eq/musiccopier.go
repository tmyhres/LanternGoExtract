package eq

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tmyhres/LanternGoExtract/pkg/infrastructure/logger"
)

const (
	// MusicDirectory is the name of the directory where music files are copied.
	MusicDirectory = "music"

	// DefaultExportsFolder is the default folder for exports.
	DefaultExportsFolder = "Exports"
)

// MusicCopierSettings contains the configuration needed for music copying.
type MusicCopierSettings struct {
	// CopyMusic indicates whether music files should be copied.
	CopyMusic bool
	// EverQuestDirectory is the path to the EverQuest installation directory.
	EverQuestDirectory string
}

// CopyMusic copies XMI music files from the EverQuest directory to the exports folder.
// It only operates when shortname is "music" or "all" and CopyMusic is enabled.
//
// Parameters:
//   - shortname: The archive shortname ("music" or "all" to trigger copying)
//   - log: Logger for status messages
//   - settings: Configuration containing the EverQuest directory path
//
// Returns an error if copying fails.
func CopyMusic(shortname string, log logger.Logger, settings *MusicCopierSettings) error {
	if settings == nil {
		return nil
	}

	// Only copy music if shortname is "music" or "all"
	if shortname != "music" && shortname != "all" {
		return nil
	}

	// Check if music copying is enabled
	if !settings.CopyMusic {
		return nil
	}

	// Find all XMI files in the EverQuest directory
	xmiFiles, err := findMusicFiles(settings.EverQuestDirectory)
	if err != nil {
		log.LogError("Failed to find music files: " + err.Error())
		return err
	}

	// Create destination folder
	destinationFolder := filepath.Join(DefaultExportsFolder, MusicDirectory)
	if err := os.MkdirAll(destinationFolder, 0755); err != nil {
		log.LogError("Failed to create music directory: " + err.Error())
		return err
	}

	// Copy each music file
	for _, xmiPath := range xmiFiles {
		fileName := filepath.Base(xmiPath)
		destination := filepath.Join(destinationFolder, fileName)

		// Skip if file already exists
		if _, err := os.Stat(destination); err == nil {
			continue
		}

		if err := copyMusicFile(xmiPath, destination); err != nil {
			log.LogWarning("Failed to copy music file " + fileName + ": " + err.Error())
			// Continue with other files
		}
	}

	return nil
}

// CopyMusicToFolder copies XMI music files to a specific folder.
// This variant allows specifying a custom export folder instead of the default.
//
// Parameters:
//   - shortname: The archive shortname ("music" or "all" to trigger copying)
//   - exportFolder: The folder to copy music files to
//   - log: Logger for status messages
//   - settings: Configuration containing the EverQuest directory path
//
// Returns an error if copying fails.
func CopyMusicToFolder(shortname, exportFolder string, log logger.Logger, settings *MusicCopierSettings) error {
	if settings == nil {
		return nil
	}

	if shortname != "music" && shortname != "all" {
		return nil
	}

	if !settings.CopyMusic {
		return nil
	}

	xmiFiles, err := findMusicFiles(settings.EverQuestDirectory)
	if err != nil {
		log.LogError("Failed to find music files: " + err.Error())
		return err
	}

	destinationFolder := filepath.Join(exportFolder, MusicDirectory)
	if err := os.MkdirAll(destinationFolder, 0755); err != nil {
		log.LogError("Failed to create music directory: " + err.Error())
		return err
	}

	for _, xmiPath := range xmiFiles {
		fileName := filepath.Base(xmiPath)
		destination := filepath.Join(destinationFolder, fileName)

		if _, err := os.Stat(destination); err == nil {
			continue
		}

		if err := copyMusicFile(xmiPath, destination); err != nil {
			log.LogWarning("Failed to copy music file " + fileName + ": " + err.Error())
		}
	}

	return nil
}

// findMusicFiles recursively searches for XMI music files in the given directory.
func findMusicFiles(directory string) ([]string, error) {
	var musicFiles []string

	err := filepath.WalkDir(directory, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Continue walking on errors
		}

		if d.IsDir() {
			return nil
		}

		if IsMusicFile(path) {
			musicFiles = append(musicFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return musicFiles, nil
}

// copyMusicFile copies a single music file from source to destination.
func copyMusicFile(source, destination string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	return os.WriteFile(destination, data, 0644)
}

// IsMusicFile checks if the given filename is a music file (XMI format).
func IsMusicFile(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".xmi")
}
