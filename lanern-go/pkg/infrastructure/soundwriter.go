package infrastructure

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/infrastructure/logger"
)

// WriteSoundAsWav writes raw audio bytes to a WAV file.
// If the file already exists, it will be overwritten and a warning will be logged.
func WriteSoundAsWav(data []byte, filePath, fileName string, log logger.Logger) error {
	if err := os.MkdirAll(filePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filePath, err)
	}

	fullPath := filepath.Join(filePath, fileName)

	if FileExists(fullPath) {
		log.LogInfo(fmt.Sprintf("SoundWriter: overwriting %s", fileName))
	}

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write sound file %s: %w", fullPath, err)
	}

	return nil
}
