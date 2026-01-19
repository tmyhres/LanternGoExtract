// Package wld provides functionality for reading and extracting data from WLD files.
// WLD files are used by the EverQuest client to store zone geometry, objects,
// characters, equipment, and other world data.
package wld

// WldType represents the type of WLD file.
type WldType int

const (
	// WldTypeZone contains main zone geometry and BSP tree.
	// Example: arena.s3d, arena.wld
	WldTypeZone WldType = iota

	// WldTypeZoneObjects contains the zone object instance data.
	// Example: arena.s3d, objects.wld
	WldTypeZoneObjects

	// WldTypeLights contains light instances.
	// Example: arena.s3d, lights.wld
	WldTypeLights

	// WldTypeObjects contains zone object model geometry.
	// Example: arena_obj.s3d, arena_obj.wld
	WldTypeObjects

	// WldTypeSky contains sky data, models and animations.
	// Example: sky.s3d, sky.wld
	WldTypeSky

	// WldTypeCharacters contains zone character models and animations.
	// Example: arena_chr.s3d, arena_chr.wld
	WldTypeCharacters

	// WldTypeEquipment contains general models - only a few of these exist.
	// Example: gequip.s3d
	WldTypeEquipment
)

// String returns the string representation of the WldType.
func (t WldType) String() string {
	switch t {
	case WldTypeZone:
		return "Zone"
	case WldTypeZoneObjects:
		return "ZoneObjects"
	case WldTypeLights:
		return "Lights"
	case WldTypeObjects:
		return "Objects"
	case WldTypeSky:
		return "Sky"
	case WldTypeCharacters:
		return "Characters"
	case WldTypeEquipment:
		return "Equipment"
	default:
		return "Unknown"
	}
}
