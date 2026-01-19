package archive

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/infrastructure/logger"
)

func TestArchiveTypeString(t *testing.T) {
	tests := []struct {
		archiveType Type
		expected    string
	}{
		{TypeUnknown, "Unknown"},
		{TypePfs, "PFS"},
		{TypeT3d, "T3D"},
	}

	for _, test := range tests {
		result := test.archiveType.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestBaseFile(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}
	f := NewBaseFile(4, 100, data)

	if f.GetSize() != 4 {
		t.Errorf("Expected size 4, got %d", f.GetSize())
	}

	if f.GetOffset() != 100 {
		t.Errorf("Expected offset 100, got %d", f.GetOffset())
	}

	if !bytes.Equal(f.GetBytes(), data) {
		t.Error("Bytes mismatch")
	}

	f.SetName("test.txt")
	if f.GetName() != "test.txt" {
		t.Errorf("Expected name test.txt, got %s", f.GetName())
	}
}

func TestPfsFile(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}
	f := NewPfsFile(0x12345678, 4, 100, data)

	if f.GetCrc() != 0x12345678 {
		t.Errorf("Expected CRC 0x12345678, got %x", f.GetCrc())
	}

	if f.GetSize() != 4 {
		t.Errorf("Expected size 4, got %d", f.GetSize())
	}
}

func TestT3dFile(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}
	f := NewT3dFile(4, 100, data)

	if f.GetSize() != 4 {
		t.Errorf("Expected size 4, got %d", f.GetSize())
	}
}

func TestNullArchive(t *testing.T) {
	log := logger.NewNullLogger()
	archive := NewNullArchive("/nonexistent/path.s3d", log)

	err := archive.Initialize()
	if err != ErrNullArchive {
		t.Errorf("Expected ErrNullArchive, got %v", err)
	}

	if archive.GetFileName() != "path.s3d" {
		t.Errorf("Expected filename path.s3d, got %s", archive.GetFileName())
	}
}

func TestBaseArchive(t *testing.T) {
	log := logger.NewNullLogger()
	archive := NewBaseArchive("/test/path/archive.s3d", log)

	if archive.GetFilePath() != "/test/path/archive.s3d" {
		t.Errorf("Expected file path /test/path/archive.s3d, got %s", archive.GetFilePath())
	}

	if archive.GetFileName() != "archive.s3d" {
		t.Errorf("Expected file name archive.s3d, got %s", archive.GetFileName())
	}

	// Test file operations
	file1 := &BaseFile{Name: "test1.txt", Size: 10, Offset: 0, Bytes: []byte("test data1")}
	file2 := &BaseFile{Name: "test2.txt", Size: 10, Offset: 10, Bytes: []byte("test data2")}

	archive.Files = append(archive.Files, file1, file2)
	archive.FileNameRef["test1.txt"] = file1
	archive.FileNameRef["test2.txt"] = file2

	// Test GetFile by name
	result := archive.GetFile("test1.txt")
	if result != file1 {
		t.Error("GetFile by name failed")
	}

	result = archive.GetFile("nonexistent.txt")
	if result != nil {
		t.Error("Expected nil for nonexistent file")
	}

	// Test GetFileByIndex
	result = archive.GetFileByIndex(0)
	if result != file1 {
		t.Error("GetFileByIndex failed")
	}

	result = archive.GetFileByIndex(99)
	if result != nil {
		t.Error("Expected nil for out of range index")
	}

	// Test GetAllFiles
	allFiles := archive.GetAllFiles()
	if len(allFiles) != 2 {
		t.Errorf("Expected 2 files, got %d", len(allFiles))
	}

	// Test RenameFile
	archive.RenameFile("test1.txt", "renamed.txt")
	if archive.GetFile("test1.txt") != nil {
		t.Error("Old name should not exist after rename")
	}
	if archive.GetFile("renamed.txt") == nil {
		t.Error("New name should exist after rename")
	}

	// Test IsWldArchive
	if archive.IsWldArchive() {
		t.Error("IsWldArchive should be false by default")
	}
	archive.SetIsWldArchive(true)
	if !archive.IsWldArchive() {
		t.Error("IsWldArchive should be true after setting")
	}
}

func TestInflateBlock(t *testing.T) {
	// Create test data and compress it
	originalData := []byte("Hello, this is test data for zlib compression!")

	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write(originalData)
	w.Close()

	// Test decompression
	decompressed, err := inflateBlock(compressed.Bytes(), len(originalData))
	if err != nil {
		t.Errorf("inflateBlock failed: %v", err)
	}

	if !bytes.Equal(decompressed, originalData) {
		t.Errorf("Decompressed data doesn't match original. Got %s, expected %s", decompressed, originalData)
	}
}

func TestGetArchiveTypeFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		expected Type
	}{
		{"test.s3d", TypePfs},
		{"test.S3D", TypePfs},
		{"test.pfs", TypePfs},
		{"test.pak", TypePfs},
		{"test.t3d", TypeT3d},
		{"test.T3D", TypeT3d},
		{"test.unknown", TypeUnknown},
		{"test.txt", TypeUnknown},
	}

	for _, test := range tests {
		result := getArchiveTypeFromFilename(test.filename)
		if result != test.expected {
			t.Errorf("For %s: expected %v, got %v", test.filename, test.expected, result)
		}
	}
}

func TestGetArchive(t *testing.T) {
	log := logger.NewNullLogger()

	// Test with non-existent file
	archive, err := GetArchive("/nonexistent/file.s3d", log)
	if err != nil {
		t.Errorf("GetArchive should not error for non-existent file: %v", err)
	}

	_, ok := archive.(*NullArchive)
	if !ok {
		t.Error("Expected NullArchive for non-existent file")
	}
}

func TestCreateMinimalPfsArchive(t *testing.T) {
	// Create a minimal valid PFS archive for testing
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.s3d")

	// Create minimal PFS structure
	var buf bytes.Buffer

	// PFS header: directory offset (4 bytes) + magic (4 bytes) + version (4 bytes)
	// We'll create a simple archive with one compressed file

	testData := []byte("test file content")

	// Compress the test data
	var compressedBuf bytes.Buffer
	w := zlib.NewWriter(&compressedBuf)
	w.Write(testData)
	w.Close()
	compressedData := compressedBuf.Bytes()

	// File block: deflated length (4) + inflated length (4) + compressed data
	fileBlockStart := uint32(12) // After header
	var fileBlock bytes.Buffer
	binary.Write(&fileBlock, binary.LittleEndian, uint32(len(compressedData)))
	binary.Write(&fileBlock, binary.LittleEndian, uint32(len(testData)))
	fileBlock.Write(compressedData)

	// Filename dictionary
	filename := "testfile.txt\x00"
	var dictData bytes.Buffer
	binary.Write(&dictData, binary.LittleEndian, uint32(1)) // 1 filename
	binary.Write(&dictData, binary.LittleEndian, uint32(len(filename)))
	dictData.WriteString(filename)

	// Compress the dictionary
	var compressedDict bytes.Buffer
	dictWriter := zlib.NewWriter(&compressedDict)
	dictWriter.Write(dictData.Bytes())
	dictWriter.Close()

	dictBlockStart := fileBlockStart + uint32(fileBlock.Len())
	var dictBlock bytes.Buffer
	binary.Write(&dictBlock, binary.LittleEndian, uint32(len(compressedDict.Bytes())))
	binary.Write(&dictBlock, binary.LittleEndian, uint32(dictData.Len()))
	dictBlock.Write(compressedDict.Bytes())

	directoryOffset := dictBlockStart + uint32(dictBlock.Len())

	// Write header
	binary.Write(&buf, binary.LittleEndian, directoryOffset) // directory offset
	binary.Write(&buf, binary.LittleEndian, PfsMagicValue)   // magic
	binary.Write(&buf, binary.LittleEndian, int32(0x20000))  // version 2

	// Write file block
	buf.Write(fileBlock.Bytes())

	// Write dictionary block
	buf.Write(dictBlock.Bytes())

	// Directory: file count (4) + entries
	// Each entry: CRC (4) + offset (4) + size (4)
	binary.Write(&buf, binary.LittleEndian, int32(2)) // 2 entries (file + dictionary)

	// File entry
	binary.Write(&buf, binary.LittleEndian, uint32(0x12345678))        // CRC (arbitrary)
	binary.Write(&buf, binary.LittleEndian, fileBlockStart)            // offset
	binary.Write(&buf, binary.LittleEndian, uint32(len(testData)))     // inflated size

	// Dictionary entry (special CRC)
	binary.Write(&buf, binary.LittleEndian, uint32(0x61580AC9))        // Dictionary CRC
	binary.Write(&buf, binary.LittleEndian, dictBlockStart)            // offset
	binary.Write(&buf, binary.LittleEndian, uint32(dictData.Len()))    // inflated size

	// Write the archive
	if err := os.WriteFile(archivePath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("Failed to write test archive: %v", err)
	}

	// Now test reading it
	log := logger.NewNullLogger()
	archive, err := GetArchive(archivePath, log)
	if err != nil {
		t.Fatalf("GetArchive failed: %v", err)
	}

	_, ok := archive.(*PfsArchive)
	if !ok {
		t.Fatal("Expected PfsArchive")
	}

	if err := archive.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	files := archive.GetAllFiles()
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	if len(files) > 0 {
		file := files[0]
		if file.GetName() != "testfile.txt" {
			t.Errorf("Expected filename 'testfile.txt', got '%s'", file.GetName())
		}

		if !bytes.Equal(file.GetBytes(), testData) {
			t.Errorf("File content mismatch. Expected '%s', got '%s'", testData, file.GetBytes())
		}
	}
}
