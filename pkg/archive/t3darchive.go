package archive

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/tmyhres/LanternGoExtract/pkg/infrastructure/logger"
)

var (
	// T3dMagic is the magic bytes for T3D archives.
	T3dMagic = []byte{0x02, 0x3D, 0xFF, 0xFF}

	// T3dVersion is the expected version bytes for T3D archives.
	T3dVersion = []byte{0x00, 0x57, 0x01, 0x00}

	// ErrIncorrectMagic is returned when the file magic is incorrect.
	ErrIncorrectMagic = errors.New("incorrect file magic")

	// ErrIncorrectVersion is returned when the file version is incorrect.
	ErrIncorrectVersion = errors.New("incorrect file version")
)

// T3dArchive represents a T3D archive.
type T3dArchive struct {
	*BaseArchive
}

// NewT3dArchive creates a new T3dArchive.
func NewT3dArchive(filePath string, log logger.Logger) *T3dArchive {
	return &T3dArchive{
		BaseArchive: NewBaseArchive(filePath, log),
	}
}

// Initialize reads and parses the T3D archive.
func (t *T3dArchive) Initialize() error {
	t.Logger.LogInfo("T3dArchive: Started initialization of archive: " + t.FileName)

	file, err := os.Open(t.FilePath)
	if err != nil {
		t.Logger.LogError("T3dArchive: File does not exist at: " + t.FilePath)
		return ErrFileNotFound
	}
	defer file.Close()

	// Read and verify magic
	magic := make([]byte, 4)
	if _, err := io.ReadFull(file, magic); err != nil {
		return err
	}

	if !bytes.Equal(magic, T3dMagic) {
		t.Logger.LogError("T3dArchive: Incorrect file magic")
		return ErrIncorrectMagic
	}

	// Read and verify version
	version := make([]byte, 4)
	if _, err := io.ReadFull(file, version); err != nil {
		return err
	}

	if !bytes.Equal(version, T3dVersion) {
		t.Logger.LogError("T3dArchive: Incorrect file version")
		return ErrIncorrectVersion
	}

	var fileCount uint32
	if err := binary.Read(file, binary.LittleEndian, &fileCount); err != nil {
		return err
	}

	var filenamesLength uint32
	if err := binary.Read(file, binary.LittleEndian, &filenamesLength); err != nil {
		return err
	}

	// Read offset pairs
	type offsetPair struct {
		FileOffset     uint32
		FileNameOffset uint32
	}

	offsetPairs := make([]offsetPair, fileCount)
	for i := uint32(0); i < fileCount; i++ {
		currentPos, _ := file.Seek(0, io.SeekCurrent)
		fileNameBaseOffset := uint32(currentPos)

		var fileOffset, fileNameOffset uint32
		if err := binary.Read(file, binary.LittleEndian, &fileOffset); err != nil {
			return err
		}
		if err := binary.Read(file, binary.LittleEndian, &fileNameOffset); err != nil {
			return err
		}

		offsetPairs[i] = offsetPair{
			FileOffset:     fileOffset,
			FileNameOffset: fileNameOffset + fileNameBaseOffset,
		}
	}

	var totalFilesize uint64
	if err := binary.Read(file, binary.LittleEndian, &totalFilesize); err != nil {
		return err
	}

	// Process files (note: last entry is typically the directory, so fileCount-1)
	for i := uint32(0); i < fileCount-1; i++ {
		fileOffset := offsetPairs[i].FileOffset
		fileNameOffset := offsetPairs[i].FileNameOffset

		var nextFileOffset uint64
		if i == fileCount-2 {
			nextFileOffset = totalFilesize
		} else {
			nextFileOffset = uint64(offsetPairs[i+1].FileOffset)
		}

		fileSize := uint32(nextFileOffset - uint64(fileOffset))
		fileBytes := make([]byte, fileSize)

		if _, err := file.Seek(int64(fileOffset), io.SeekStart); err != nil {
			return err
		}
		if _, err := io.ReadFull(file, fileBytes); err != nil {
			return err
		}

		t3dFile := NewT3dFile(fileSize, fileOffset, fileBytes)

		// Read filename (null-terminated string)
		if _, err := file.Seek(int64(fileNameOffset), io.SeekStart); err != nil {
			return err
		}

		fileName, err := readNullTerminatedString(file)
		if err != nil {
			return err
		}

		t3dFile.SetName(strings.ToLower(fileName))

		if !t.IsWld && strings.HasSuffix(strings.ToLower(fileName), WldExtension) {
			t.IsWld = true
		}

		t.Files = append(t.Files, t3dFile)
		t.FileNameRef[t3dFile.GetName()] = t3dFile
	}

	t.Logger.LogInfo("T3dArchive: Finished initialization of archive: " + t.FileName)
	return nil
}

// readNullTerminatedString reads a null-terminated string from the reader.
func readNullTerminatedString(r io.Reader) (string, error) {
	var result []byte
	buf := make([]byte, 1)

	for {
		if _, err := io.ReadFull(r, buf); err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		if buf[0] == 0 {
			break
		}

		result = append(result, buf[0])
	}

	return string(result), nil
}
