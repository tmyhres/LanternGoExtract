package sound

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"os"
)

// SoundTest provides helper functions for sound isolation and testing.
type SoundTest struct{}

// OutputSingleInstance extracts a single sound instance from the data and writes it to a file.
// This is useful for debugging and testing individual sound entries.
func OutputSingleInstance(data []byte, index int, fileName string) error {
	if index < 0 {
		return nil
	}

	startOffset := index * EntryLengthInBytes
	endOffset := startOffset + EntryLengthInBytes

	if endOffset > len(data) {
		return nil
	}

	instanceBytes := data[startOffset:endOffset]
	return os.WriteFile(fileName, instanceBytes, 0644)
}

// OutputSingleInstanceFromReader extracts a single sound instance from a reader and writes it to a file.
func OutputSingleInstanceFromReader(reader io.ReadSeeker, index int, fileName string) error {
	// Seek to beginning
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Read all data
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	return OutputSingleInstance(data, index, fileName)
}

// ModifyInstance modifies a sound instance's position and writes the result to a file.
// This is useful for testing sound positioning.
func ModifyInstance(data []byte, fileName string) error {
	if len(data) < EntryLengthInBytes {
		return nil
	}

	// Create a copy to avoid modifying the original
	modified := make([]byte, len(data))
	copy(modified, data)

	// Modify positions at offset 16
	// Position X = 0
	binary.LittleEndian.PutUint32(modified[16:], math.Float32bits(0))
	// Position Y = 0
	binary.LittleEndian.PutUint32(modified[20:], math.Float32bits(0))
	// Position Z = 50
	binary.LittleEndian.PutUint32(modified[24:], math.Float32bits(50))

	return os.WriteFile(fileName, modified, 0644)
}

// ModifyInstanceFromReader modifies a sound instance from a reader and writes it to a file.
func ModifyInstanceFromReader(reader io.ReadSeeker, fileName string) error {
	// Seek to beginning
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Read all data
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	return ModifyInstance(data, fileName)
}

// CreateTestEntry creates a test sound entry with the specified parameters.
// This is useful for creating test fixtures.
func CreateTestEntry(
	posX, posY, posZ, radius float32,
	audioType AudioType,
	soundID1, soundID2 int32,
) []byte {
	entry := make([]byte, EntryLengthInBytes)

	// Write positions at offset 16
	binary.LittleEndian.PutUint32(entry[16:], math.Float32bits(posX))
	binary.LittleEndian.PutUint32(entry[20:], math.Float32bits(posY))
	binary.LittleEndian.PutUint32(entry[24:], math.Float32bits(posZ))
	binary.LittleEndian.PutUint32(entry[28:], math.Float32bits(radius))

	// Write sound IDs at offset 48
	binary.LittleEndian.PutUint32(entry[48:], uint32(soundID1))
	binary.LittleEndian.PutUint32(entry[52:], uint32(soundID2))

	// Write audio type at offset 56
	entry[56] = byte(audioType)

	return entry
}

// ParseTestEntry parses a single entry from raw bytes for testing purposes.
func ParseTestEntry(data []byte) (*TestEntryData, error) {
	if len(data) < EntryLengthInBytes {
		return nil, ErrInvalidEffFile
	}

	reader := bytes.NewReader(data)

	entry := &TestEntryData{}

	// Skip first 16 bytes
	reader.Seek(16, io.SeekStart)

	// Read positions
	if err := binary.Read(reader, binary.LittleEndian, &entry.PosX); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &entry.PosY); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &entry.PosZ); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &entry.Radius); err != nil {
		return nil, err
	}

	// Read cooldowns at offset 32
	reader.Seek(32, io.SeekStart)
	if err := binary.Read(reader, binary.LittleEndian, &entry.Cooldown1); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &entry.Cooldown2); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &entry.CooldownRandom); err != nil {
		return nil, err
	}

	// Read sound IDs at offset 48
	reader.Seek(48, io.SeekStart)
	if err := binary.Read(reader, binary.LittleEndian, &entry.SoundID1); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &entry.SoundID2); err != nil {
		return nil, err
	}

	// Read type at offset 56
	reader.Seek(56, io.SeekStart)
	var typeByte byte
	if err := binary.Read(reader, binary.LittleEndian, &typeByte); err != nil {
		return nil, err
	}
	entry.AudioType = AudioType(typeByte)

	// Read volumes at offset 60
	reader.Seek(60, io.SeekStart)
	if err := binary.Read(reader, binary.LittleEndian, &entry.Volume1); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &entry.Volume2); err != nil {
		return nil, err
	}

	// Read multiplier at offset 72
	reader.Seek(72, io.SeekStart)
	if err := binary.Read(reader, binary.LittleEndian, &entry.Multiplier); err != nil {
		return nil, err
	}

	return entry, nil
}

// TestEntryData holds the parsed data from a single sound entry for testing.
type TestEntryData struct {
	PosX           float32
	PosY           float32
	PosZ           float32
	Radius         float32
	AudioType      AudioType
	SoundID1       int32
	SoundID2       int32
	Cooldown1      int32
	Cooldown2      int32
	CooldownRandom int32
	Volume1        int32
	Volume2        int32
	Multiplier     int32
}
