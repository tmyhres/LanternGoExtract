package archive

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/lanterneq/lanern-go/pkg/infrastructure/logger"
)

const (
	// WldExtension is the file extension for WLD files.
	WldExtension = ".wld"

	// pfsFilenameCrc is the CRC for the filename directory entry.
	pfsFilenameCrc = 0x61580AC9

	// eqZipFilenameCrc is used by EQZip for filename directory entries.
	eqZipFilenameCrc = 0xFFFFFFFF

	// pfsVersion1 indicates PFS version 1 format.
	pfsVersion1 = 0x10000

	// pfsVersion2 indicates PFS version 2 format.
	pfsVersion2 = 0x20000
)

var (
	// ErrCorruptedLength is returned when a corrupted length is detected.
	ErrCorruptedLength = errors.New("corrupted archive length detected")

	// ErrInflate is returned when decompression fails.
	ErrInflate = errors.New("error inflating compressed data")

	// ErrFileNotFound is returned when the archive file does not exist.
	ErrFileNotFound = errors.New("archive file does not exist")

	// ErrUnexpectedVersion is returned for unexpected PFS versions.
	ErrUnexpectedVersion = errors.New("unexpected pfs version")
)

// PfsArchive represents a PFS/S3D archive.
type PfsArchive struct {
	*BaseArchive
}

// NewPfsArchive creates a new PfsArchive.
func NewPfsArchive(filePath string, log logger.Logger) *PfsArchive {
	return &PfsArchive{
		BaseArchive: NewBaseArchive(filePath, log),
	}
}

// Initialize reads and parses the PFS archive.
func (p *PfsArchive) Initialize() error {
	p.Logger.LogInfo("PfsArchive: Started initialization of archive: " + p.FileName)

	file, err := os.Open(p.FilePath)
	if err != nil {
		p.Logger.LogError("PfsArchive: File does not exist at: " + p.FilePath)
		return ErrFileNotFound
	}
	defer file.Close()

	reader := newBinaryReader(file)

	// Read header
	directoryOffset, err := reader.readInt32()
	if err != nil {
		return err
	}

	_, err = reader.readUint32() // pfsMagic
	if err != nil {
		return err
	}

	pfsVersion, err := reader.readInt32()
	if err != nil {
		return err
	}

	// Seek to directory
	if _, err := file.Seek(int64(directoryOffset), io.SeekStart); err != nil {
		return err
	}

	fileCount, err := reader.readInt32()
	if err != nil {
		return err
	}

	fileNames := make([]string, 0)
	streamLen, _ := file.Seek(0, io.SeekEnd)
	file.Seek(int64(directoryOffset)+4, io.SeekStart) // Reset to after fileCount

	for i := int32(0); i < fileCount; i++ {
		crc, err := reader.readUint32()
		if err != nil {
			return err
		}

		offset, err := reader.readUint32()
		if err != nil {
			return err
		}

		size, err := reader.readUint32()
		if err != nil {
			return err
		}

		if int64(offset) > streamLen {
			p.Logger.LogError("PfsArchive: Corrupted PFS length detected!")
			return ErrCorruptedLength
		}

		// Save current position
		cachedOffset, _ := file.Seek(0, io.SeekCurrent)

		// Read file data
		fileBytes := make([]byte, size)
		if _, err := file.Seek(int64(offset), io.SeekStart); err != nil {
			return err
		}

		var inflatedSize uint32

		for inflatedSize != size {
			deflatedLength, err := reader.readUint32()
			if err != nil {
				return err
			}

			inflatedLength, err := reader.readUint32()
			if err != nil {
				return err
			}

			if int64(deflatedLength) >= streamLen {
				p.Logger.LogError("PfsArchive: Corrupted file length detected!")
				return ErrCorruptedLength
			}

			compressedBytes := make([]byte, deflatedLength)
			if _, err := io.ReadFull(file, compressedBytes); err != nil {
				return err
			}

			inflatedBytes, err := inflateBlock(compressedBytes, int(inflatedLength))
			if err != nil {
				p.Logger.LogError("PfsArchive: Error occured inflating data: " + err.Error())
				return ErrInflate
			}

			copy(fileBytes[inflatedSize:], inflatedBytes)
			inflatedSize += inflatedLength
		}

		// Check if this is a filename directory
		// EQZip saved archives use 0xFFFFFFFF for filenames
		if crc == pfsFilenameCrc || (crc == eqZipFilenameCrc && len(fileNames) == 0) {
			dictReader := bytes.NewReader(fileBytes)
			var filenameCount uint32
			if err := binary.Read(dictReader, binary.LittleEndian, &filenameCount); err != nil {
				return err
			}

			for j := uint32(0); j < filenameCount; j++ {
				var fileNameLength uint32
				if err := binary.Read(dictReader, binary.LittleEndian, &fileNameLength); err != nil {
					return err
				}

				fileNameBytes := make([]byte, fileNameLength)
				if _, err := io.ReadFull(dictReader, fileNameBytes); err != nil {
					return err
				}

				// Remove null terminator
				fileName := string(fileNameBytes[:len(fileNameBytes)-1])
				fileNames = append(fileNames, fileName)
			}

			// Restore position and continue
			if _, err := file.Seek(cachedOffset, io.SeekStart); err != nil {
				return err
			}
			continue
		}

		p.Files = append(p.Files, NewPfsFile(crc, size, offset, fileBytes))

		// Restore position
		if _, err := file.Seek(cachedOffset, io.SeekStart); err != nil {
			return err
		}
	}

	// Sort files by offset so we can assign names
	sort.Slice(p.Files, func(i, j int) bool {
		return p.Files[i].GetOffset() < p.Files[j].GetOffset()
	})

	// Assign file names
	for i, f := range p.Files {
		switch pfsVersion {
		case pfsVersion1:
			// PFS version 1 files do not appear to contain the filenames
			if pfsFile, ok := f.(*PfsFile); ok {
				pfsFile.SetName(fmt.Sprintf("%08X.bin", pfsFile.Crc))
			}
		case pfsVersion2:
			if i < len(fileNames) {
				f.SetName(fileNames[i])
				p.FileNameRef[fileNames[i]] = f

				if !p.IsWld && strings.HasSuffix(strings.ToLower(fileNames[i]), WldExtension) {
					p.IsWld = true
				}
			}
		default:
			p.Logger.LogError("PfsArchive: Unexpected pfs version: " + p.FileName)
			return ErrUnexpectedVersion
		}
	}

	p.Logger.LogInfo("PfsArchive: Finished initialization of archive: " + p.FileName)
	return nil
}

// inflateBlock decompresses a zlib-compressed block.
func inflateBlock(deflatedBytes []byte, inflatedSize int) ([]byte, error) {
	reader := bytes.NewReader(deflatedBytes)
	zlibReader, err := zlib.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer zlibReader.Close()

	output := make([]byte, inflatedSize)
	n, err := io.ReadFull(zlibReader, output)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("failed to read inflated data: %w", err)
	}

	return output[:n], nil
}

// binaryReader wraps an io.Reader for little-endian binary reading.
type binaryReader struct {
	reader io.Reader
}

func newBinaryReader(r io.Reader) *binaryReader {
	return &binaryReader{reader: r}
}

func (b *binaryReader) readInt32() (int32, error) {
	var val int32
	err := binary.Read(b.reader, binary.LittleEndian, &val)
	return val, err
}

func (b *binaryReader) readUint32() (uint32, error) {
	var val uint32
	err := binary.Read(b.reader, binary.LittleEndian, &val)
	return val, err
}
