package wld

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/tmyhres/LanternGoExtract/pkg/archive"
	"github.com/tmyhres/LanternGoExtract/pkg/infrastructure/logger"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
)

// Common errors for WLD file processing.
var (
	ErrInvalidWldFile     = errors.New("not a valid WLD file")
	ErrUnrecognizedFormat = errors.New("unrecognized WLD format")
)

// WldFile defines the interface for all WLD file types.
// Different WLD types (zone, objects, characters, etc.) implement this interface
// with their specific processing and export logic.
type WldFile interface {
	// Initialize reads and parses the WLD file data.
	// Returns an error if initialization fails.
	Initialize(rootFolder string, exportData bool) error

	// GetWldType returns the type of this WLD file.
	GetWldType() WldType

	// GetZoneName returns the zone shortname.
	GetZoneName() string

	// GetFragments returns all parsed fragments.
	GetFragments() []fragments.Fragment

	// GetFragmentsOfType returns all fragments of a specific type.
	// The type parameter should be a pointer to the fragment type.
	GetFragmentsOfType(fragmentType interface{}) []fragments.Fragment

	// GetFragmentByName returns a fragment by its name.
	GetFragmentByName(name string) fragments.Fragment

	// ProcessData processes the parsed fragment data.
	ProcessData()

	// ExportData exports the WLD data to files.
	ExportData()

	// GetRootExportFolder returns the root export folder path.
	GetRootExportFolder() string

	// GetExportFolderForWldType returns the export folder for this WLD type.
	GetExportFolderForWldType() string

	// GetFilenameChanges returns a map of filename changes made during processing.
	GetFilenameChanges() map[string]string

	// GetAllBitmapNames returns all bitmap names from the WLD file.
	GetAllBitmapNames() []string

	// GetActors returns all Actor fragments from the WLD file.
	GetActors() []*fragments.Actor

	// GetMeshes returns all Mesh fragments from the WLD file.
	GetMeshes() []*fragments.Mesh

	// GetMaterialLists returns all MaterialList fragments from the WLD file.
	GetMaterialLists() []*fragments.MaterialList
}

// BaseWldFile provides common functionality for all WLD file types.
// It implements the core parsing logic and fragment management.
type BaseWldFile struct {
	// RootExportFolder is the root folder for exporting data.
	RootExportFolder string

	// WldType is the type of this WLD file.
	WldType WldType

	// ZoneName is the shortname of the zone.
	ZoneName string

	// Fragments is a list of all parsed fragments.
	Fragments []fragments.Fragment

	// FragmentTypeDictionary maps fragment types to lists of fragments.
	FragmentTypeDictionary map[reflect.Type][]fragments.Fragment

	// FragmentNameDictionary maps fragment names to fragments.
	FragmentNameDictionary map[string]fragments.Fragment

	// StringHash contains the decoded string hash table.
	StringHash map[int]string

	// Logger for debug output.
	Logger logger.Logger

	// WldData is the raw WLD file data.
	WldData archive.File

	// IsNewWldFormat indicates if this is the new WLD format.
	IsNewWldFormat bool

	// WldToInject is another WLD file to inject data from.
	WldToInject WldFile

	// FilenameChanges tracks any filename changes made.
	FilenameChanges map[string]string

	// Settings contains configuration options.
	Settings *Settings
}

// Settings contains configuration options for WLD file processing.
type Settings struct {
	// ModelExportFormat determines the export format for models.
	ModelExportFormat ModelExportFormat

	// ExportCharactersToSingleFolder exports all characters to one folder.
	ExportCharactersToSingleFolder bool

	// ExportEquipmentToSingleFolder exports all equipment to one folder.
	ExportEquipmentToSingleFolder bool
}

// ModelExportFormat represents the format for exporting models.
type ModelExportFormat int

const (
	// ModelExportFormatIntermediate exports to intermediate text format.
	ModelExportFormatIntermediate ModelExportFormat = iota

	// ModelExportFormatObj exports to OBJ format.
	ModelExportFormatObj

	// ModelExportFormatGltf exports to glTF format.
	ModelExportFormatGltf
)

// NewBaseWldFile creates a new BaseWldFile.
func NewBaseWldFile(wldData archive.File, zoneName string, wldType WldType, log logger.Logger, settings *Settings, wldToInject WldFile) *BaseWldFile {
	return &BaseWldFile{
		WldType:                wldType,
		ZoneName:               strings.ToLower(zoneName),
		Logger:                 log,
		WldData:                wldData,
		Settings:               settings,
		WldToInject:            wldToInject,
		Fragments:              make([]fragments.Fragment, 0),
		FragmentTypeDictionary: make(map[reflect.Type][]fragments.Fragment),
		FragmentNameDictionary: make(map[string]fragments.Fragment),
		StringHash:             make(map[int]string),
		FilenameChanges:        make(map[string]string),
	}
}

// Initialize parses the WLD file data.
func (w *BaseWldFile) Initialize(rootFolder string, exportData bool) error {
	w.RootExportFolder = rootFolder

	if w.WldData == nil {
		return errors.New("WLD data is nil")
	}

	w.Logger.LogInfo(fmt.Sprintf("Extracting WLD archive: %s", w.WldData.GetName()))
	w.Logger.LogInfo("-----------------------------------")
	w.Logger.LogInfo(fmt.Sprintf("WLD type: %s", w.WldType.String()))

	data := w.WldData.GetBytes()
	reader := bytes.NewReader(data)

	// Read and validate file identifier
	var identifier int32
	if err := binary.Read(reader, binary.LittleEndian, &identifier); err != nil {
		return fmt.Errorf("failed to read WLD identifier: %w", err)
	}

	if identifier != WldFileIdentifier {
		w.Logger.LogError("Not a valid WLD file!")
		return ErrInvalidWldFile
	}

	// Read and validate version
	var version int32
	if err := binary.Read(reader, binary.LittleEndian, &version); err != nil {
		return fmt.Errorf("failed to read WLD version: %w", err)
	}

	switch version {
	case WldFormatOldIdentifier:
		w.IsNewWldFormat = false
	case WldFormatNewIdentifier:
		w.IsNewWldFormat = true
		w.Logger.LogWarning("New WLD format not fully supported.")
	default:
		w.Logger.LogError("Unrecognized WLD format.")
		return ErrUnrecognizedFormat
	}

	// Read header fields
	var fragmentCount uint32
	var bspRegionCount uint32
	var unknown int32
	var stringHashSize uint32
	var unknown2 int32

	if err := binary.Read(reader, binary.LittleEndian, &fragmentCount); err != nil {
		return fmt.Errorf("failed to read fragment count: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &bspRegionCount); err != nil {
		return fmt.Errorf("failed to read BSP region count: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &unknown); err != nil {
		return fmt.Errorf("failed to read unknown field: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &stringHashSize); err != nil {
		return fmt.Errorf("failed to read string hash size: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &unknown2); err != nil {
		return fmt.Errorf("failed to read unknown2 field: %w", err)
	}

	// Read and decode string hash
	stringHashBytes := make([]byte, stringHashSize)
	if _, err := reader.Read(stringHashBytes); err != nil {
		return fmt.Errorf("failed to read string hash: %w", err)
	}

	decodedHash := DecodeString(stringHashBytes)
	w.parseStringHash(decodedHash)

	// Parse fragments
	for i := uint32(0); i < fragmentCount; i++ {
		var fragSize uint32
		var fragID int32

		if err := binary.Read(reader, binary.LittleEndian, &fragSize); err != nil {
			return fmt.Errorf("failed to read fragment %d size: %w", i, err)
		}
		if err := binary.Read(reader, binary.LittleEndian, &fragID); err != nil {
			return fmt.Errorf("failed to read fragment %d ID: %w", i, err)
		}

		fragData := make([]byte, fragSize)
		if _, err := reader.Read(fragData); err != nil {
			return fmt.Errorf("failed to read fragment %d data: %w", i, err)
		}

		// Create the appropriate fragment type
		newFragment := CreateFragment(int(fragID))

		if _, isGeneric := newFragment.(*fragments.Generic); isGeneric {
			w.Logger.LogWarning(fmt.Sprintf("WldFile: Unhandled fragment type: %x", fragID))
		}

		// Initialize the fragment
		if err := newFragment.Initialize(int(i), int(fragID), int(fragSize), fragData, w.Fragments, w.StringHash, w.IsNewWldFormat); err != nil {
			w.Logger.LogWarning(fmt.Sprintf("Failed to initialize fragment %d (type %x): %v", i, fragID, err))
		}

		w.Fragments = append(w.Fragments, newFragment)

		// Add to type dictionary
		fragType := reflect.TypeOf(newFragment)
		if _, ok := w.FragmentTypeDictionary[fragType]; !ok {
			w.FragmentTypeDictionary[fragType] = make([]fragments.Fragment, 0)
		}
		w.FragmentTypeDictionary[fragType] = append(w.FragmentTypeDictionary[fragType], newFragment)

		// Add to name dictionary
		if name := newFragment.GetName(); name != "" {
			if _, exists := w.FragmentNameDictionary[name]; !exists {
				w.FragmentNameDictionary[name] = newFragment
			}
		}
	}

	w.Logger.LogInfo("-----------------------------------")
	w.Logger.LogInfo("WLD extraction complete")

	w.ProcessData()

	if exportData {
		w.ExportData()
	}

	return nil
}

// parseStringHash parses the decoded string hash into a map.
func (w *BaseWldFile) parseStringHash(decodedHash string) {
	index := 0
	splitHash := strings.Split(decodedHash, "\x00")

	for _, hashString := range splitHash {
		w.StringHash[index] = hashString
		// Advance by string length + null terminator
		index += len(hashString) + 1
	}
}

// GetWldType returns the WLD file type.
func (w *BaseWldFile) GetWldType() WldType {
	return w.WldType
}

// GetZoneName returns the zone shortname.
func (w *BaseWldFile) GetZoneName() string {
	return w.ZoneName
}

// GetFragments returns all fragments.
func (w *BaseWldFile) GetFragments() []fragments.Fragment {
	return w.Fragments
}

// GetFragmentsOfType returns all fragments of a specific type.
func (w *BaseWldFile) GetFragmentsOfType(fragmentType interface{}) []fragments.Fragment {
	fragType := reflect.TypeOf(fragmentType)
	if frags, ok := w.FragmentTypeDictionary[fragType]; ok {
		return frags
	}
	return make([]fragments.Fragment, 0)
}

// GetFragmentsByType returns fragments of a specific type using generics pattern.
// Usage: frags := GetFragmentsByType[*fragments.Mesh](wldFile)
func GetFragmentsByType[T fragments.Fragment](w WldFile) []T {
	allFrags := w.GetFragments()
	result := make([]T, 0)
	for _, frag := range allFrags {
		if typed, ok := frag.(T); ok {
			result = append(result, typed)
		}
	}
	return result
}

// GetFragmentByName returns a fragment by name.
func (w *BaseWldFile) GetFragmentByName(name string) fragments.Fragment {
	if frag, ok := w.FragmentNameDictionary[name]; ok {
		return frag
	}
	return nil
}

// GetFragmentByNameTyped returns a fragment by name with type assertion.
func GetFragmentByNameTyped[T fragments.Fragment](w WldFile, name string) T {
	frag := w.GetFragmentByName(name)
	if frag != nil {
		if typed, ok := frag.(T); ok {
			return typed
		}
	}
	var zero T
	return zero
}

// ProcessData processes the parsed fragment data.
// Override in derived types for specific processing.
func (w *BaseWldFile) ProcessData() {
	// Base implementation - can be overridden
}

// ExportData exports the WLD data.
// Override in derived types for specific export logic.
func (w *BaseWldFile) ExportData() {
	// Base implementation - can be overridden
}

// GetRootExportFolder returns the root export folder.
func (w *BaseWldFile) GetRootExportFolder() string {
	return w.RootExportFolder + w.ZoneName + "/"
}

// GetExportFolderForWldType returns the export folder for this WLD type.
func (w *BaseWldFile) GetExportFolderForWldType() string {
	switch w.WldType {
	case WldTypeZone, WldTypeLights, WldTypeZoneObjects:
		return w.GetRootExportFolder() + "Zone/"
	case WldTypeEquipment:
		return w.GetRootExportFolder()
	case WldTypeObjects:
		return w.GetRootExportFolder() + "Objects/"
	case WldTypeSky:
		return w.GetRootExportFolder()
	case WldTypeCharacters:
		if w.Settings != nil && w.Settings.ExportCharactersToSingleFolder &&
			w.Settings.ModelExportFormat == ModelExportFormatIntermediate {
			return w.GetRootExportFolder()
		}
		return w.GetRootExportFolder() + "Characters/"
	default:
		return ""
	}
}

// GetFilenameChanges returns filename changes.
func (w *BaseWldFile) GetFilenameChanges() map[string]string {
	return w.FilenameChanges
}

// GetAllBitmapNames returns all bitmap names from the WLD file.
func (w *BaseWldFile) GetAllBitmapNames() []string {
	bitmaps := make([]string, 0)
	for _, frag := range w.Fragments {
		if bitmapName, ok := frag.(*fragments.BitmapName); ok {
			bitmaps = append(bitmaps, bitmapName.Filename)
		}
	}
	return bitmaps
}

// GetActors returns all Actor fragments from the WLD file.
func (w *BaseWldFile) GetActors() []*fragments.Actor {
	return GetFragmentsByType[*fragments.Actor](w)
}

// GetMeshes returns all Mesh fragments from the WLD file.
func (w *BaseWldFile) GetMeshes() []*fragments.Mesh {
	return GetFragmentsByType[*fragments.Mesh](w)
}

// GetMaterialLists returns all MaterialList fragments from the WLD file.
func (w *BaseWldFile) GetMaterialLists() []*fragments.MaterialList {
	return GetFragmentsByType[*fragments.MaterialList](w)
}
