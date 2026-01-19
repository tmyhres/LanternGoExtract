package eq

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lanterneq/lanern-go/pkg/infrastructure/logger"
	"github.com/lanterneq/lanern-go/pkg/wld"
)

const (
	// ClientDataDirectory is the name of the directory where client data files are copied.
	ClientDataDirectory = "clientdata"
)

// ClientDataCopierSettings contains the configuration needed for client data copying.
type ClientDataCopierSettings struct {
	// ClientDataToCopy is a comma-separated list of files to copy from the EverQuest directory.
	ClientDataToCopy string
	// ModelExportFormat determines the export format for models.
	ModelExportFormat wld.ModelExportFormat
	// EverQuestDirectory is the path to the EverQuest installation directory.
	EverQuestDirectory string
}

// CopyClientData copies client data files from the EverQuest directory to the export folder.
// It only copies when the fileName is "clientdata" or "all" and the export format is Intermediate.
//
// Parameters:
//   - fileName: The archive name being extracted ("clientdata" or "all" to trigger copying)
//   - rootFolder: The root export folder
//   - log: Logger for status messages
//   - settings: Configuration containing file list and paths
//
// Returns an error if copying fails.
func CopyClientData(fileName, rootFolder string, log logger.Logger, settings *ClientDataCopierSettings) error {
	if settings == nil {
		return nil
	}

	if settings.ClientDataToCopy == "" {
		return nil
	}

	if settings.ModelExportFormat != wld.ModelExportFormatIntermediate {
		return nil
	}

	if !isValidClientDataName(fileName) {
		return nil
	}

	return writeAllClientDataFiles(rootFolder, log, settings)
}

// writeAllClientDataFiles copies all configured client data files to the destination.
func writeAllClientDataFiles(rootFolder string, log logger.Logger, settings *ClientDataCopierSettings) error {
	destDir := filepath.Join(rootFolder, ClientDataDirectory)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create client data directory: %w", err)
	}

	filePaths := getClientDataFilePaths(settings)

	for _, srcPath := range filePaths {
		destPath := getDestinationPath(rootFolder, srcPath)
		log.LogInfo(fmt.Sprintf("Copying %s to %s", srcPath, destPath))

		if err := copyFile(srcPath, destPath); err != nil {
			log.LogError(fmt.Sprintf("Failed to copy %s: %v", srcPath, err))
			// Continue with other files even if one fails
		}
	}

	return nil
}

// getClientDataFilePaths returns a list of valid file paths from the settings.
func getClientDataFilePaths(settings *ClientDataCopierSettings) []string {
	var paths []string

	files := strings.Split(settings.ClientDataToCopy, ",")
	for _, file := range files {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}

		clientDataFilePath := filepath.Join(settings.EverQuestDirectory, file)
		if _, err := os.Stat(clientDataFilePath); err == nil {
			paths = append(paths, clientDataFilePath)
		}
	}

	return paths
}

// getDestinationPath constructs the destination file path.
func getDestinationPath(rootFolder, sourceFilePath string) string {
	sourceFileName := filepath.Base(sourceFilePath)
	return filepath.Join(rootFolder, ClientDataDirectory, sourceFileName)
}

// isValidClientDataName checks if the filename indicates client data should be copied.
func isValidClientDataName(fileName string) bool {
	return IsClientDataFile(fileName) || fileName == "all"
}

// copyFile copies a file from source to destination, overwriting if it exists.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}
