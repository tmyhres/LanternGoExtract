package sound

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// ErrInvalidEALFile is returned when the EAL file format is invalid.
var ErrInvalidEALFile = errors.New("invalid EAL file format")

// EAXAttributes contains EAX audio effect attributes.
type EAXAttributes struct {
	DirectPathLevel int32
	// Additional EAX attributes can be added here as needed
}

// SourceAttributes contains the attributes for an audio source.
type SourceAttributes struct {
	EaxAttributes EAXAttributes
}

// SourceModel represents an audio source model from the EAL file.
type SourceModel struct {
	Name             string
	SourceAttributes SourceAttributes
}

// EalData contains the parsed data from an EAL file.
type EalData struct {
	SourceModels []SourceModel
}

// EnvAudio handles parsing of EAL/EAX audio environment files.
// EAL files are binary files containing audio source definitions with volume levels.
type EnvAudio struct {
	data         *EalData
	loaded       bool
	ealFilePath  string
	sourceLevels map[string]int32
}

// NewEnvAudio creates a new EnvAudio instance.
func NewEnvAudio() *EnvAudio {
	return &EnvAudio{
		sourceLevels: make(map[string]int32),
	}
}

// Load reads and parses the EAL file at the given path.
// Returns true if loading succeeded, false otherwise.
func (e *EnvAudio) Load(ealFilePath string) (bool, error) {
	if e.ealFilePath == ealFilePath {
		return e.loaded, nil
	}

	if ealFilePath == "" {
		return false, nil
	}

	if _, err := os.Stat(ealFilePath); os.IsNotExist(err) {
		return false, nil
	}

	e.ealFilePath = ealFilePath

	file, err := os.Open(ealFilePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	data, err := e.parseEALFile(file)
	if err != nil {
		return false, err
	}

	e.data = data
	if e.data == nil {
		return false, nil
	}

	// Build the source levels map
	e.sourceLevels = make(map[string]int32)
	for _, source := range e.data.SourceModels {
		// Get filename without extension and convert to lowercase
		baseName := filepath.Base(source.Name)
		nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		key := strings.ToLower(nameWithoutExt)
		e.sourceLevels[key] = source.SourceAttributes.EaxAttributes.DirectPathLevel
	}

	e.loaded = true
	return e.loaded, nil
}

// parseEALFile parses an EAL binary file.
// EAL files have a specific binary format containing source models with audio attributes.
func (e *EnvAudio) parseEALFile(reader io.Reader) (*EalData, error) {
	data := &EalData{
		SourceModels: make([]SourceModel, 0),
	}

	// Read full file into memory for easier parsing
	fileData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if len(fileData) < 4 {
		return nil, ErrInvalidEALFile
	}

	// EAL file format (reverse engineered):
	// The file contains a header followed by source model entries.
	// Each source model has a name string and EAX attributes.

	offset := 0

	// Read magic/version header (4 bytes)
	if offset+4 > len(fileData) {
		return nil, ErrInvalidEALFile
	}
	magic := binary.LittleEndian.Uint32(fileData[offset:])
	offset += 4

	// Check for known EAL magic numbers
	// EAL files typically start with a version or magic number
	if magic == 0 {
		return data, nil // Empty file
	}

	// Try to parse as a simple format with count + entries
	// Reset and try alternative parsing
	offset = 0

	// Attempt to find source models in the file
	// The format varies, so we use heuristics to locate sound name entries

	// Look for patterns that indicate sound file names (.wav references)
	sources, err := e.extractSourceModels(fileData)
	if err != nil {
		return nil, err
	}

	data.SourceModels = sources
	return data, nil
}

// extractSourceModels extracts source model data from raw EAL file bytes.
// This uses pattern matching to find sound definitions in the binary data.
func (e *EnvAudio) extractSourceModels(data []byte) ([]SourceModel, error) {
	sources := make([]SourceModel, 0)

	// EAL files contain source definitions with:
	// - A length-prefixed or null-terminated string for the sound name
	// - EAX attributes including direct path level (volume)

	// Scan for string patterns that look like sound file names
	i := 0
	for i < len(data)-8 {
		// Look for potential string length prefix (reasonable length 1-255)
		potentialLen := int(data[i])
		if potentialLen > 0 && potentialLen < 256 && i+1+potentialLen+4 <= len(data) {
			// Check if the following bytes look like a filename
			nameBytes := data[i+1 : i+1+potentialLen]
			if isValidSoundName(nameBytes) {
				name := string(nameBytes)

				// Try to find the DirectPathLevel after the name
				// It's typically stored as a 32-bit integer in the attributes section
				attrOffset := i + 1 + potentialLen

				// Look for the volume level in the following bytes
				// The exact offset depends on the EAL version/format
				directLevel := int32(0)
				if attrOffset+4 <= len(data) {
					// Read potential volume value
					val := int32(binary.LittleEndian.Uint32(data[attrOffset:]))
					// Volume levels are typically negative (attenuation in millibels)
					// Valid range is roughly -10000 to 0
					if val >= -10000 && val <= 0 {
						directLevel = val
					}
				}

				sources = append(sources, SourceModel{
					Name: name,
					SourceAttributes: SourceAttributes{
						EaxAttributes: EAXAttributes{
							DirectPathLevel: directLevel,
						},
					},
				})

				i = attrOffset + 4
				continue
			}
		}

		// Also check for null-terminated strings
		if data[i] != 0 && isASCIIPrintable(data[i]) {
			// Scan for null terminator
			end := i
			for end < len(data) && data[end] != 0 && end-i < 256 {
				end++
			}
			if end > i && end < len(data) {
				nameBytes := data[i:end]
				if isValidSoundName(nameBytes) {
					name := string(nameBytes)

					// Try to find volume after the null terminator
					attrOffset := end + 1
					directLevel := int32(0)
					if attrOffset+4 <= len(data) {
						val := int32(binary.LittleEndian.Uint32(data[attrOffset:]))
						if val >= -10000 && val <= 0 {
							directLevel = val
						}
					}

					sources = append(sources, SourceModel{
						Name: name,
						SourceAttributes: SourceAttributes{
							EaxAttributes: EAXAttributes{
								DirectPathLevel: directLevel,
							},
						},
					})

					i = end + 1
					continue
				}
			}
		}

		i++
	}

	return sources, nil
}

// isValidSoundName checks if the bytes represent a valid sound file name.
func isValidSoundName(b []byte) bool {
	if len(b) < 3 {
		return false
	}

	// Check for common sound file extensions
	s := strings.ToLower(string(b))
	hasExt := strings.HasSuffix(s, ".wav") ||
		strings.HasSuffix(s, ".mp3") ||
		strings.HasSuffix(s, ".ogg")

	if !hasExt {
		return false
	}

	// All characters should be printable ASCII
	for _, c := range b {
		if !isASCIIPrintable(c) {
			return false
		}
	}

	return true
}

// isASCIIPrintable returns true if the byte is a printable ASCII character.
func isASCIIPrintable(b byte) bool {
	return b >= 32 && b < 127
}

// GetVolumeEQ returns the EQ volume level for the given sound file.
// Returns 0 if the sound is not found.
func (e *EnvAudio) GetVolumeEQ(soundFile string) int32 {
	if e.sourceLevels == nil {
		return 0
	}
	volume, _ := e.sourceLevels[strings.ToLower(soundFile)]
	return volume
}

// GetVolumeLinear converts an EQ volume level to a linear 0-1 scale.
func (e *EnvAudio) GetVolumeLinear(soundFile string) float32 {
	volumeEQ := e.GetVolumeEQ(soundFile)
	return e.GetVolumeLinearFromLevel(volumeEQ)
}

// GetVolumeLinearFromLevel converts a direct audio level (in millibels) to linear scale.
// The formula is: linear = 10^(level/2000)
// The result is clamped to [0, 1].
func (e *EnvAudio) GetVolumeLinearFromLevel(directAudioLevel int32) float32 {
	linear := float32(math.Pow(10.0, float64(directAudioLevel)/2000.0))
	if linear < 0 {
		return 0
	}
	if linear > 1 {
		return 1
	}
	return linear
}

// Data returns the parsed EAL data.
func (e *EnvAudio) Data() *EalData {
	return e.data
}

// IsLoaded returns true if an EAL file has been successfully loaded.
func (e *EnvAudio) IsLoaded() bool {
	return e.loaded
}

// SourceCount returns the number of source models loaded.
func (e *EnvAudio) SourceCount() int {
	if e.data == nil {
		return 0
	}
	return len(e.data.SourceModels)
}

// String returns a string representation of the EnvAudio state.
func (e *EnvAudio) String() string {
	if !e.loaded {
		return "EnvAudio: not loaded"
	}
	return fmt.Sprintf("EnvAudio: loaded %d sources from %s", e.SourceCount(), e.ealFilePath)
}
