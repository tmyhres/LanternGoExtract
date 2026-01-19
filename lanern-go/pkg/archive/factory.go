package archive

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/infrastructure/logger"
)

const (
	// T3dMagicValue is the magic number for T3D archives.
	T3dMagicValue uint32 = 0xffff3d02

	// PfsMagicValue is the magic number for PFS archives.
	PfsMagicValue uint32 = 0x20534650

	// S3dExtension is the file extension for S3D archives.
	S3dExtension = ".s3d"

	// PfsExtension is the file extension for PFS archives.
	PfsExtension = ".pfs"

	// PakExtension is the file extension for PAK archives.
	PakExtension = ".pak"

	// T3dExtension is the file extension for T3D archives.
	T3dExtension = ".t3d"
)

// ErrUnknownArchiveType is returned when the archive type cannot be determined.
var ErrUnknownArchiveType = errors.New("unknown archive type")

// GetArchive returns the appropriate archive implementation for the given file.
func GetArchive(filePath string, log logger.Logger) (Archive, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Skip detection and let the archive Initialize fail.
		return NewNullArchive(filePath, log), nil
	}

	archiveType := getArchiveTypeFromMagic(filePath)
	if archiveType == TypeUnknown {
		archiveType = getArchiveTypeFromFilename(filePath)
	}

	switch archiveType {
	case TypePfs:
		return NewPfsArchive(filePath, log), nil
	case TypeT3d:
		return NewT3dArchive(filePath, log), nil
	default:
		return nil, ErrUnknownArchiveType
	}
}

// getArchiveTypeFromMagic detects the archive type by reading magic bytes.
func getArchiveTypeFromMagic(filePath string) Type {
	file, err := os.Open(filePath)
	if err != nil {
		return TypeUnknown
	}
	defer file.Close()

	// Read first uint32
	var data uint32
	if err := binary.Read(file, binary.LittleEndian, &data); err != nil {
		return TypeUnknown
	}

	if data == T3dMagicValue {
		return TypeT3d
	}

	// Read second uint32 (PFS magic is at offset 4)
	if err := binary.Read(file, binary.LittleEndian, &data); err != nil {
		return TypeUnknown
	}

	if data == PfsMagicValue {
		return TypePfs
	}

	return TypeUnknown
}

// getArchiveTypeFromFilename detects the archive type from the file extension.
func getArchiveTypeFromFilename(filePath string) Type {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case T3dExtension:
		return TypeT3d
	case S3dExtension, PfsExtension, PakExtension:
		return TypePfs
	default:
		return TypeUnknown
	}
}
