package archive

import (
	"os"
	"path/filepath"

	"github.com/tmyhres/LanternGoExtract/pkg/infrastructure/logger"
)

// Archive defines the interface for reading archive files.
type Archive interface {
	// Initialize reads and parses the archive file.
	// Returns an error if initialization fails.
	Initialize() error

	// GetFilePath returns the full path to the archive file.
	GetFilePath() string

	// GetFileName returns just the filename of the archive.
	GetFileName() string

	// GetFile returns a file by name, or nil if not found.
	GetFile(name string) File

	// GetFileByIndex returns a file by index, or nil if out of range.
	GetFileByIndex(index int) File

	// GetAllFiles returns all files in the archive.
	GetAllFiles() []File

	// WriteAllFiles writes all archive files to the specified folder.
	WriteAllFiles(folder string) error

	// RenameFile renames a file within the archive.
	RenameFile(originalName, newName string)

	// IsWldArchive returns true if this archive contains WLD files.
	IsWldArchive() bool

	// SetIsWldArchive sets whether this archive contains WLD files.
	SetIsWldArchive(isWld bool)

	// GetFilenameChanges returns a map of filename changes.
	GetFilenameChanges() map[string]string
}

// BaseArchive provides common functionality for archive implementations.
type BaseArchive struct {
	FilePath        string
	FileName        string
	Files           []File
	FileNameRef     map[string]File
	Logger          logger.Logger
	IsWld           bool
	FilenameChanges map[string]string
}

// NewBaseArchive creates a new BaseArchive with the given parameters.
func NewBaseArchive(filePath string, log logger.Logger) *BaseArchive {
	return &BaseArchive{
		FilePath:        filePath,
		FileName:        filepath.Base(filePath),
		Files:           make([]File, 0),
		FileNameRef:     make(map[string]File),
		Logger:          log,
		FilenameChanges: make(map[string]string),
	}
}

// GetFilePath returns the full path to the archive file.
func (b *BaseArchive) GetFilePath() string {
	return b.FilePath
}

// GetFileName returns just the filename of the archive.
func (b *BaseArchive) GetFileName() string {
	return b.FileName
}

// GetFile returns a file by name, or nil if not found.
func (b *BaseArchive) GetFile(name string) File {
	if f, ok := b.FileNameRef[name]; ok {
		return f
	}
	return nil
}

// GetFileByIndex returns a file by index, or nil if out of range.
func (b *BaseArchive) GetFileByIndex(index int) File {
	if index < 0 || index >= len(b.Files) {
		return nil
	}
	return b.Files[index]
}

// GetAllFiles returns all files in the archive.
func (b *BaseArchive) GetAllFiles() []File {
	return b.Files
}

// WriteAllFiles writes all archive files to the specified folder.
func (b *BaseArchive) WriteAllFiles(folder string) error {
	for _, f := range b.Files {
		filePath := filepath.Join(folder, f.GetName())
		dir := filepath.Dir(filePath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		if err := os.WriteFile(filePath, f.GetBytes(), 0644); err != nil {
			return err
		}
	}
	return nil
}

// RenameFile renames a file within the archive.
func (b *BaseArchive) RenameFile(originalName, newName string) {
	f, ok := b.FileNameRef[originalName]
	if !ok {
		return
	}

	delete(b.FileNameRef, originalName)
	f.SetName(newName)
	b.FileNameRef[newName] = f
}

// IsWldArchive returns true if this archive contains WLD files.
func (b *BaseArchive) IsWldArchive() bool {
	return b.IsWld
}

// SetIsWldArchive sets whether this archive contains WLD files.
func (b *BaseArchive) SetIsWldArchive(isWld bool) {
	b.IsWld = isWld
}

// GetFilenameChanges returns a map of filename changes.
func (b *BaseArchive) GetFilenameChanges() map[string]string {
	return b.FilenameChanges
}
