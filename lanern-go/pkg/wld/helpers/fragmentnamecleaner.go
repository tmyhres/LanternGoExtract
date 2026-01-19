// Package helpers provides utility functions for WLD file processing.
package helpers

import (
	"strings"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/fragments"
)

// FragmentPrefixes maps fragment type names to their internal name prefixes.
// These prefixes are used in EverQuest WLD files to identify fragment types.
var FragmentPrefixes = map[string]string{
	"MaterialList":      "_MP",
	"Material":          "_MDF",
	"Mesh":              "_DMSPRITEDEF",
	"LegacyMesh":        "_DMSPRITEDEF",
	"Actor":             "_ACTORDEF",
	"SkeletonHierarchy": "_HS_DEF",
	"TrackDefFragment":  "_TRACKDEF",
	"TrackFragment":     "_TRACK",
	"ParticleCloud":     "_PCD",
}

// fragmentTypePrefixes maps fragment type IDs to their internal name prefixes.
var fragmentTypePrefixes = map[uint32]string{
	0x31: "_MP",          // MaterialList
	0x30: "_MDF",         // Material
	0x36: "_DMSPRITEDEF", // Mesh
	0x2C: "_DMSPRITEDEF", // LegacyMesh
	0x14: "_ACTORDEF",    // Actor
	0x10: "_HS_DEF",      // SkeletonHierarchy
	0x12: "_TRACKDEF",    // TrackDefFragment
	0x13: "_TRACK",       // TrackFragment
	0x34: "_PCD",         // ParticleCloud
}

// CleanName cleans a fragment name by removing the type-specific prefix.
// If toLower is true, the result is converted to lowercase.
func CleanName(name, fragmentType string, toLower bool) string {
	cleanedName := name

	if prefix, exists := FragmentPrefixes[fragmentType]; exists {
		cleanedName = strings.ReplaceAll(cleanedName, prefix, "")
	}

	if toLower {
		cleanedName = strings.ToLower(cleanedName)
	}

	return strings.TrimSpace(cleanedName)
}

// CleanFragmentName cleans a fragment's name using its type ID.
// This is the Go equivalent of the C# FragmentNameCleaner.CleanName method.
func CleanFragmentName(fragment fragments.Fragment, toLower bool) string {
	if fragment == nil {
		return ""
	}

	cleanedName := fragment.GetName()

	if prefix, ok := fragmentTypePrefixes[fragment.FragmentType()]; ok {
		cleanedName = strings.Replace(cleanedName, prefix, "", 1)
	}

	if toLower {
		cleanedName = strings.ToLower(cleanedName)
	}

	return strings.TrimSpace(cleanedName)
}

// CleanFragmentNameDefault cleans a fragment name with lowercase conversion enabled.
func CleanFragmentNameDefault(fragment fragments.Fragment) string {
	return CleanFragmentName(fragment, true)
}

// CleanActorName cleans an actor name by removing the _ACTORDEF suffix.
func CleanActorName(name string) string {
	cleanedName := strings.ToLower(name)
	cleanedName = strings.TrimSuffix(cleanedName, "_actordef")
	return strings.TrimSpace(cleanedName)
}

// CleanSkeletonName cleans a skeleton name by removing the _HS_DEF suffix.
func CleanSkeletonName(name string) string {
	cleanedName := strings.ToLower(name)
	cleanedName = strings.TrimSuffix(cleanedName, "_hs_def")
	return strings.TrimSpace(cleanedName)
}

// CleanTrackName cleans a track name by removing the _TRACK suffix.
func CleanTrackName(name string) string {
	cleanedName := strings.ToLower(name)
	cleanedName = strings.TrimSuffix(cleanedName, "_track")
	return strings.TrimSpace(cleanedName)
}

// CleanTrackDefName cleans a track definition name by removing the _TRACKDEF suffix.
func CleanTrackDefName(name string) string {
	cleanedName := strings.ToLower(name)
	cleanedName = strings.TrimSuffix(cleanedName, "_trackdef")
	return strings.TrimSpace(cleanedName)
}

// CleanMeshName cleans a mesh name by removing the _DMSPRITEDEF suffix.
func CleanMeshName(name string) string {
	cleanedName := strings.ToLower(name)
	cleanedName = strings.TrimSuffix(cleanedName, "_dmspritedef")
	return strings.TrimSpace(cleanedName)
}

// CleanMaterialName cleans a material name by removing the _MDF suffix.
func CleanMaterialName(name string) string {
	cleanedName := strings.ToLower(name)
	cleanedName = strings.TrimSuffix(cleanedName, "_mdf")
	return strings.TrimSpace(cleanedName)
}

// CleanMaterialListName cleans a material list name by removing the _MP suffix.
func CleanMaterialListName(name string) string {
	cleanedName := strings.ToLower(name)
	cleanedName = strings.TrimSuffix(cleanedName, "_mp")
	return strings.TrimSpace(cleanedName)
}
