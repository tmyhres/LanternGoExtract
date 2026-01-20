package helpers

import (
	"os"
	"path/filepath"
)

// MissingTextureFixer handles copying textures that are missing from zone S3D files.
//
// In some cases, textures used in zones are not included in the zone S3D
// because they are also used in the object S3D. The EQ client reads all
// textures into a pool so this is not a problem. With Lantern, textures
// are separated by S3D type. Therefore, some post extraction copying
// must be done to fix the missing textures.

// TextureCopyInfo holds information about a texture that needs to be copied.
type TextureCopyInfo struct {
	Source      string
	Destination string
}

// missingTextureMap maps zone shortnames to their required texture copies.
var missingTextureMap = map[string]TextureCopyInfo{
	"oasis_obj": {
		Source:      "Exports/oasis/Objects/Textures/canwall1.png",
		Destination: "Exports/oasis/Zone/Textures/canwall1.png",
	},
	"fearplane_obj": {
		Source:      "Exports/fearplane/Objects/Textures/maywall.png",
		Destination: "Exports/fearplane/Zone/Textures/maywall.png",
	},
	"swampofnohope_obj": {
		Source:      "Exports/swampofnohope/Objects/Textures/kruphse3.png",
		Destination: "Exports/swampofnohope/Zone/Textures/kruphse3.png",
	},
}

// FixMissingTextures copies any missing textures for the given zone shortname.
// It checks if the zone requires texture fixes and copies them if the source
// exists and the destination does not already exist.
func FixMissingTextures(shortname string) error {
	copyInfo, needsFix := missingTextureMap[shortname]
	if !needsFix {
		return nil
	}

	return copyTexture(copyInfo.Source, copyInfo.Destination)
}

// FixMissingTexturesWithBase copies missing textures using a base directory path.
// This is useful when the export directory is not in the current working directory.
func FixMissingTexturesWithBase(shortname, basePath string) error {
	copyInfo, needsFix := missingTextureMap[shortname]
	if !needsFix {
		return nil
	}

	source := filepath.Join(basePath, copyInfo.Source)
	destination := filepath.Join(basePath, copyInfo.Destination)

	return copyTexture(source, destination)
}

// copyTexture copies a texture file from source to destination.
// It only copies if the source exists, the destination does not exist,
// and the destination directory exists.
func copyTexture(source, destination string) error {
	// Check if source exists
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return nil // Source doesn't exist, nothing to copy
	}

	// Check if destination already exists
	if _, err := os.Stat(destination); err == nil {
		return nil // Destination already exists, no need to copy
	}

	// Check if destination directory exists
	destDir := filepath.Dir(destination)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		return nil // Destination directory doesn't exist, skip
	}

	// Read source file
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	// Write to destination
	return os.WriteFile(destination, data, 0644)
}

// GetMissingTextureList returns a copy of the missing texture map.
// This can be used for debugging or logging purposes.
func GetMissingTextureList() map[string]TextureCopyInfo {
	result := make(map[string]TextureCopyInfo, len(missingTextureMap))
	for k, v := range missingTextureMap {
		result[k] = v
	}
	return result
}

// AddMissingTextureEntry adds a custom missing texture entry to the map.
// This allows for extending the missing texture fixes without modifying the package.
func AddMissingTextureEntry(shortname, source, destination string) {
	missingTextureMap[shortname] = TextureCopyInfo{
		Source:      source,
		Destination: destination,
	}
}

// NeedsTextureFix returns true if the given shortname requires texture fixes.
func NeedsTextureFix(shortname string) bool {
	_, needsFix := missingTextureMap[shortname]
	return needsFix
}
