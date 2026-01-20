package archive

import (
	"errors"

	"github.com/tmyhres/LanternGoExtract/pkg/infrastructure/logger"
)

// ErrNullArchive is returned when attempting to initialize a null archive.
var ErrNullArchive = errors.New("null archive: file does not exist or format unknown")

// NullArchive represents an archive that could not be loaded.
type NullArchive struct {
	*BaseArchive
}

// NewNullArchive creates a new NullArchive.
func NewNullArchive(filePath string, log logger.Logger) *NullArchive {
	return &NullArchive{
		BaseArchive: NewBaseArchive(filePath, log),
	}
}

// Initialize always returns an error for NullArchive.
func (n *NullArchive) Initialize() error {
	return ErrNullArchive
}
