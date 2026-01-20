package helpers

import (
	"strings"

	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
)

// CharacterFixer fixes numerous issues in EverQuest's character model files.
// A number of character models have incorrect shader assignments.
// As character models are specific to each zone, there are also conflicts with:
// 1. Texture names being used for different characters
// 2. Zones using different models/skeleton for the same race id
// This code is only run when exporting all characters to a single folder for batch importing.

// WldFileCharacters is an interface that a character WLD file must implement
// to work with CharacterFixer.
type WldFileCharacters interface {
	// GetFragmentsOfType returns all fragments of a specific type.
	GetFragmentsOfType(fragmentType uint32) []fragments.Fragment

	// GetFragmentByName returns a fragment by its name.
	GetFragmentByName(name string) fragments.Fragment

	// GetFilenameChanges returns the map of filename changes (new name -> original name).
	GetFilenameChanges() map[string]string

	// SetFilenameChange records a filename change.
	SetFilenameChange(newName, originalName string)
}

// CharacterFixer provides methods to fix character model issues.
type CharacterFixer struct {
	wld WldFileCharacters
}

// NewCharacterFixer creates a new CharacterFixer instance.
func NewCharacterFixer() *CharacterFixer {
	return &CharacterFixer{}
}

// Fix applies all character fixes to the given WLD file.
func (f *CharacterFixer) Fix(wld WldFileCharacters) {
	f.wld = wld
	f.fixShipNames()
	f.fixGolemElemental()
	f.fixSnowDervish()
	f.fixAkanonKing()
	f.fixKaladimKing()
	f.fixFayDrake()
	f.fixTurtleTextures()
	f.fixBlackAndWhiteDragon()
	f.fixGhoulTextures()
	f.fixHalasFemale()
	f.fixBetaBeetle()
	f.fixHighpassMale()
}

// FixCharacters is a convenience function that creates a CharacterFixer and applies fixes.
func FixCharacters(wld WldFileCharacters) {
	fixer := NewCharacterFixer()
	fixer.Fix(wld)
}

// fixGhoulTextures fixes Ghoul face being applied to the back of the leg.
// Thanks to modestlaw for requesting this fix.
func (f *CharacterFixer) fixGhoulTextures() {
	meshFrags := f.wld.GetFragmentsOfType(0x36) // Mesh fragment type

	if len(meshFrags) == 0 {
		return
	}

	for _, frag := range meshFrags {
		mesh, ok := frag.(*fragments.Mesh)
		if !ok || mesh == nil {
			continue
		}

		if strings.HasPrefix(mesh.Name, "GHUHE00") {
			if len(mesh.MaterialGroups) > 1 {
				mesh.MaterialGroups[1].MaterialIndex = 7
			}
		} else if strings.HasPrefix(mesh.Name, "GHU_") {
			if len(mesh.MaterialGroups) > 0 {
				mesh.MaterialGroups[0].MaterialIndex = 0
			}
		}
	}
}

// fixTurtleTextures fixes the turtle textures being named incorrectly.
// They use the seahorse prefix.
func (f *CharacterFixer) fixTurtleTextures() {
	actorFrags := f.wld.GetFragmentsOfType(0x14) // Actor fragment type

	for _, frag := range actorFrags {
		actor, ok := frag.(*fragments.Actor)
		if !ok || actor == nil {
			continue
		}

		if !strings.HasPrefix(actor.Name, "STU") {
			continue
		}

		if actor.SkeletonReference == nil ||
			actor.SkeletonReference.SkeletonHierarchy == nil ||
			len(actor.SkeletonReference.SkeletonHierarchy.Meshes) == 0 {
			continue
		}

		firstMesh := actor.SkeletonReference.SkeletonHierarchy.Meshes[0]
		mesh, ok := firstMesh.(*fragments.Mesh)
		if !ok || mesh == nil || mesh.MaterialList == nil {
			continue
		}

		materialList := mesh.MaterialList
		materialList.Name = strings.Replace(materialList.Name, "SEA", "STU", -1)

		for _, material := range materialList.Materials {
			material.Name = strings.Replace(material.Name, "SEA", "STU", -1)

			bitmapNames := getMaterialBitmapNames(material)
			for i, originalName := range bitmapNames {
				newName := strings.Replace(originalName, "sea", "stu", -1)
				setMaterialBitmapName(material, i, newName)
				f.wld.SetFilenameChange(newName, originalName)
			}
		}
	}
}

// fixFayDrake fixes the Fay Drake model naming conflict.
// There are two different versions of the Fay Drake (colorful and brown).
// The easiest way to differentiate is that the colorful version has two meshes and the brown has one.
// This changes the colorful variant to have a unique model name.
func (f *CharacterFixer) fixFayDrake() {
	actorFrags := f.wld.GetFragmentsOfType(0x14)

	for _, frag := range actorFrags {
		actor, ok := frag.(*fragments.Actor)
		if !ok || actor == nil {
			continue
		}

		if !strings.HasPrefix(actor.Name, "FDR") {
			continue
		}

		if actor.SkeletonReference == nil ||
			actor.SkeletonReference.SkeletonHierarchy == nil {
			continue
		}

		skeleton := actor.SkeletonReference.SkeletonHierarchy
		if len(skeleton.Meshes) != 2 {
			continue
		}

		// Rename actor
		actor.Name = strings.Replace(actor.Name, "FDR", "FDF", -1)

		// Rename skeleton reference
		skeletonRef := actor.SkeletonReference
		skeletonRef.Name = strings.Replace(skeletonRef.Name, "FDR", "FDF", -1)

		// Rename skeleton
		skeleton.Name = strings.Replace(skeleton.Name, "FDR", "FDF", -1)

		// Rename skeleton bones
		skeleton.RenameNodeBase("fdf")

		// Rename all main meshes
		for _, meshFrag := range skeleton.Meshes {
			meshFrag.SetName(strings.Replace(meshFrag.GetName(), "FDR", "FDF", -1))
		}

		// Rename all secondary meshes
		for _, meshFrag := range skeleton.SecondaryMeshes {
			meshFrag.SetName(strings.Replace(meshFrag.GetName(), "FDR", "FDF", -1))
		}

		// Rename all materials
		if len(skeleton.Meshes) == 0 {
			continue
		}

		firstMesh, ok := skeleton.Meshes[0].(*fragments.Mesh)
		if !ok || firstMesh == nil || firstMesh.MaterialList == nil {
			continue
		}

		materialList := firstMesh.MaterialList
		materialList.Name = strings.Replace(materialList.Name, "FDR", "FDF", -1)

		for _, material := range materialList.Materials {
			material.Name = strings.Replace(material.Name, "FDR", "FDF", -1)

			bitmapNames := getMaterialBitmapNames(material)
			for i, originalName := range bitmapNames {
				newName := strings.Replace(originalName, "fdr", "fdf", -1)
				setMaterialBitmapName(material, i, newName)
				f.wld.SetFilenameChange(newName, originalName)
			}
		}
	}
}

// fixKaladimKing fixes the unused Kaladim king model crown shader assignment.
func (f *CharacterFixer) fixKaladimKing() {
	frag := f.wld.GetFragmentByName("KAHE0001_MDF")
	if frag == nil {
		return
	}

	material, ok := frag.(*fragments.Material)
	if !ok || material == nil {
		return
	}

	material.ShaderType = fragments.ShaderTypeTransparentMasked
}

// fixAkanonKing fixes the unused Ak'Anon king model crown shader assignment.
func (f *CharacterFixer) fixAkanonKing() {
	frag := f.wld.GetFragmentByName("CLHE0004_MDF")
	if frag == nil {
		return
	}

	material, ok := frag.(*fragments.Material)
	if !ok || material == nil {
		return
	}

	material.ShaderType = fragments.ShaderTypeTransparentMasked
}

// fixSnowDervish fixes incorrect shader assignment and texture naming conflicts with normal Dervish.
func (f *CharacterFixer) fixSnowDervish() {
	actorFrags := f.wld.GetFragmentsOfType(0x14)

	for _, frag := range actorFrags {
		actor, ok := frag.(*fragments.Actor)
		if !ok || actor == nil {
			continue
		}

		if actor.SkeletonReference == nil ||
			actor.SkeletonReference.SkeletonHierarchy == nil {
			continue
		}

		if !strings.HasPrefix(actor.Name, "SDE") {
			continue
		}

		skeleton := actor.SkeletonReference.SkeletonHierarchy
		for _, meshFrag := range skeleton.Meshes {
			mesh, ok := meshFrag.(*fragments.Mesh)
			if !ok || mesh == nil || mesh.MaterialList == nil {
				continue
			}

			for _, material := range mesh.MaterialList.Materials {
				// This texture needs to be masked
				if material.Name == "SDEUA0006_MDF" {
					material.ShaderType = fragments.ShaderTypeTransparentMasked
				}

				bitmapNames := getMaterialBitmapNames(material)
				for i, name := range bitmapNames {
					if !strings.HasPrefix(name, "dml") {
						continue
					}

					originalName := name
					newName := strings.Replace(originalName, "dml", "sde", -1)
					setMaterialBitmapName(material, i, newName)
					f.wld.SetFilenameChange(newName, originalName)
				}
			}
		}
	}
}

// fixGolemElemental fixes Golem Elemental incorrectly using normal Golem material names.
func (f *CharacterFixer) fixGolemElemental() {
	frag := f.wld.GetFragmentByName("GOM_ACTORDEF")
	if frag == nil {
		return
	}

	actor, ok := frag.(*fragments.Actor)
	if !ok || actor == nil {
		return
	}

	if actor.SkeletonReference == nil ||
		actor.SkeletonReference.SkeletonHierarchy == nil {
		return
	}

	skeleton := actor.SkeletonReference.SkeletonHierarchy
	for _, meshFrag := range skeleton.Meshes {
		mesh, ok := meshFrag.(*fragments.Mesh)
		if !ok || mesh == nil || mesh.MaterialList == nil {
			continue
		}

		for _, material := range mesh.MaterialList.Materials {
			if !strings.HasPrefix(material.Name, "GOL") {
				continue
			}

			material.Name = strings.Replace(material.Name, "GOL", "GOM", -1)

			bitmapNames := getMaterialBitmapNames(material)
			for i, originalName := range bitmapNames {
				newName := strings.Replace(originalName, "gol", "gom", -1)
				setMaterialBitmapName(material, i, newName)
				f.wld.SetFilenameChange(newName, originalName)
			}
		}
	}
}

// fixShipNames fixes various ship model naming issues.
func (f *CharacterFixer) fixShipNames() {
	actorFrags := f.wld.GetFragmentsOfType(0x14)

	for _, frag := range actorFrags {
		actor, ok := frag.(*fragments.Actor)
		if !ok || actor == nil {
			continue
		}

		// Fix ghost ship
		if strings.HasPrefix(actor.Name, "GSP") {
			mesh := getActorMesh(actor)
			if mesh != nil {
				mesh.Name = strings.Replace(mesh.Name, "GHOSTSHIP", "GSP", -1)
				if mesh.MaterialList != nil {
					mesh.MaterialList.Name = strings.Replace(mesh.MaterialList.Name, "GHOSTSHIP", "GSP", -1)
				}
			}
		}

		// Fix launch
		if strings.HasPrefix(actor.Name, "LAUNCH") {
			mesh := getActorMesh(actor)
			if mesh != nil {
				actor.Name = strings.Replace(mesh.Name, "DMSPRITEDEF", "ACTORDEF", -1)
			}
		}

		// Fix PRE ships
		if strings.HasPrefix(actor.Name, "PRE") {
			if actor.SkeletonReference == nil ||
				actor.SkeletonReference.SkeletonHierarchy == nil {
				continue
			}

			skeleton := actor.SkeletonReference.SkeletonHierarchy
			switch skeleton.Name {
			case "OGS_HS_DEF":
				// Bloated Belly in Iceclad
				actor.Name = strings.Replace(actor.Name, "PRE", "OGS", -1)
			case "PRE_HS_DEF":
				// Sea King, Golden Maiden, StormBreaker, SirensBane
				// No change needed
			}
		}

		// Fix SHIP
		if strings.HasPrefix(actor.Name, "SHIP") {
			if actor.SkeletonReference == nil ||
				actor.SkeletonReference.SkeletonHierarchy == nil {
				continue
			}

			skeleton := actor.SkeletonReference.SkeletonHierarchy
			switch skeleton.Name {
			case "GNS_HS_DEF":
				// Icebreaker in Iceclad
				actor.Name = strings.Replace(actor.Name, "SHIP", "GNS", -1)
			case "ELS_HS_DEF":
				// Maidens Voyage in Firiona Vie
				actor.Name = strings.Replace(actor.Name, "SHIP", "ELS", -1)
			}
		}
	}
}

// fixBlackAndWhiteDragon fixes the Black and White Dragon material issues.
func (f *CharacterFixer) fixBlackAndWhiteDragon() {
	actorFrags := f.wld.GetFragmentsOfType(0x14)

	for _, frag := range actorFrags {
		actor, ok := frag.(*fragments.Actor)
		if !ok || actor == nil {
			continue
		}

		if !strings.HasPrefix(actor.Name, "BWD") {
			continue
		}

		materialFrag := f.wld.GetFragmentByName("BWDCH0101_MDF")
		if materialFrag != nil {
			if material, ok := materialFrag.(*fragments.Material); ok {
				material.ShaderType = fragments.ShaderTypeDiffuse
			}
		}

		// Note: The C# code has a TODO for fixing the material list slots
		// This would require additional MaterialList functionality
	}
}

// fixHalasFemale fixes the Halas Female skeleton bone naming issues.
func (f *CharacterFixer) fixHalasFemale() {
	frag := f.wld.GetFragmentByName("HLF_HS_DEF")
	if frag == nil {
		return
	}

	skeleton, ok := frag.(*fragments.SkeletonHierarchy)
	if !ok || skeleton == nil {
		return
	}

	renameBone(skeleton, 7, "bi_l")
	renameBone(skeleton, 10, "l_point")
	renameBone(skeleton, 15, "head_point")
}

// fixBetaBeetle fixes early beta beetles that have SPI named bones.
func (f *CharacterFixer) fixBetaBeetle() {
	frag := f.wld.GetFragmentByName("BET_HS_DEF")
	if frag == nil {
		return
	}

	skeleton, ok := frag.(*fragments.SkeletonHierarchy)
	if !ok || skeleton == nil {
		return
	}

	for i, bone := range skeleton.Skeleton {
		boneName := bone.CleanedName
		if strings.HasPrefix(boneName, "spi") {
			renameBone(skeleton, i, boneName[3:])
		}
	}
}

// fixHighpassMale fixes Highpass Citizen (Guard) having two head attach points.
// This can cause issues with particle emitting from incorrect skeleton point.
func (f *CharacterFixer) fixHighpassMale() {
	frag := f.wld.GetFragmentByName("HHM_HS_DEF")
	if frag == nil {
		return
	}

	skeleton, ok := frag.(*fragments.SkeletonHierarchy)
	if !ok || skeleton == nil {
		return
	}

	if len(skeleton.Skeleton) < 25 {
		return
	}

	skeleton.Skeleton[0].Children = []int{1}

	// Remove the extra bone at index 24
	if len(skeleton.Skeleton) > 24 {
		skeleton.Skeleton = append(skeleton.Skeleton[:24], skeleton.Skeleton[25:]...)
	}
}

// renameBone renames a bone in the skeleton hierarchy.
func renameBone(skeleton *fragments.SkeletonHierarchy, index int, newBoneName string) {
	if index >= len(skeleton.Skeleton) || index < 0 {
		return
	}

	oldBoneName := skeleton.BoneMappingClean[index]

	skeleton.BoneMappingClean[index] = newBoneName
	skeleton.Skeleton[index].CleanedName = newBoneName

	// Update full paths that contain the old bone name
	for i := range skeleton.Skeleton {
		fullPath := skeleton.Skeleton[i].CleanedFullPath
		if strings.Contains(fullPath, oldBoneName) {
			skeleton.Skeleton[i].CleanedFullPath = strings.Replace(fullPath, oldBoneName, newBoneName, -1)
		}
	}
}

// getActorMesh returns the mesh for an actor, handling both skeleton and direct mesh references.
func getActorMesh(actor *fragments.Actor) *fragments.Mesh {
	if actor == nil {
		return nil
	}

	// Try skeleton reference first
	if actor.SkeletonReference != nil &&
		actor.SkeletonReference.SkeletonHierarchy != nil &&
		len(actor.SkeletonReference.SkeletonHierarchy.Meshes) > 0 {
		if mesh, ok := actor.SkeletonReference.SkeletonHierarchy.Meshes[0].(*fragments.Mesh); ok {
			return mesh
		}
	}

	// Try direct mesh reference
	if actor.MeshReference != nil {
		// MeshReference could be either a MeshReference wrapper or a Mesh directly
		if meshRef, ok := actor.MeshReference.(*fragments.MeshReference); ok {
			if mesh, ok := meshRef.Mesh.(*fragments.Mesh); ok {
				return mesh
			}
		}
		if mesh, ok := actor.MeshReference.(*fragments.Mesh); ok {
			return mesh
		}
	}

	return nil
}

// getMaterialBitmapNames returns all bitmap names for a material.
// This is a helper function that handles the material's bitmap reference chain.
func getMaterialBitmapNames(material *fragments.Material) []string {
	// Note: This requires access to the BitmapInfoReference -> BitmapInfo -> BitmapName chain
	// The actual implementation depends on how these types are connected in the Go codebase.
	// For now, return empty slice - this should be implemented based on the actual fragment structure.
	return material.GetAllBitmapNames(true)
}

// setMaterialBitmapName sets a bitmap name at the given index.
// This is a helper function that handles the material's bitmap reference chain.
func setMaterialBitmapName(material *fragments.Material, index int, newName string) {
	// Note: This requires modification of the BitmapInfoReference -> BitmapInfo -> BitmapName chain
	// The actual implementation depends on how these types are connected in the Go codebase.
	// This is a placeholder - the actual implementation should modify the bitmap name at the given index.
}
