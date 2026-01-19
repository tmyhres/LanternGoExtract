// Package exporters provides export functionality for WLD data to various formats.
package exporters

import (
	"fmt"
	"strings"

	"github.com/lanterneq/lanern-go/pkg/wld"
	"github.com/lanterneq/lanern-go/pkg/wld/datatypes"
	"github.com/lanterneq/lanern-go/pkg/wld/fragments"
	"github.com/lanterneq/lanern-go/pkg/wld/helpers"
)

// ActorExportSettings contains settings for actor export.
type ActorExportSettings struct {
	ExportGltfInGlbFormat          bool
	ExportGltfVertexColors         bool
	ExportZoneWithObjects          bool
	ExportAllAnimationFrames       bool
	ExportCharactersToSingleFolder bool
}

// ExportActorsToGltf exports all actors from a WLD file to glTF format.
func ExportActorsToGltf(actors []*fragments.Actor, meshes []*fragments.Mesh, materialLists []*fragments.MaterialList,
	wldType wld.WldType, shortName string, exportFolder string, settings *ActorExportSettings) error {

	// For a zone wld, we ignore actors and just export all meshes
	if wldType == wld.WldTypeZone {
		return ExportZoneMeshes(meshes, materialLists, shortName, exportFolder, settings)
	}

	for _, actor := range actors {
		switch actor.ActorType {
		case datatypes.ActorTypeStatic:
			if err := ExportStaticActor(actor, settings, exportFolder); err != nil {
				return fmt.Errorf("failed to export static actor: %w", err)
			}
		case datatypes.ActorTypeSkeletal:
			if err := ExportSkeletalActor(actor, settings, exportFolder, wldType); err != nil {
				return fmt.Errorf("failed to export skeletal actor: %w", err)
			}
		default:
			continue
		}
	}

	return nil
}

// ExportZoneMeshes exports zone meshes to a combined glTF file.
func ExportZoneMeshes(zoneMeshes []*fragments.Mesh, materialLists []*fragments.MaterialList,
	shortName string, exportFolder string, settings *ActorExportSettings) error {

	if len(zoneMeshes) == 0 {
		return nil
	}

	exportFormat := GltfExportFormatGlTF
	if settings.ExportGltfInGlbFormat {
		exportFormat = GltfExportFormatGlb
	}

	gltfWriter := NewGltfWriter(settings.ExportGltfVertexColors, exportFormat)
	textureImageFolder := exportFolder + "Textures/"
	gltfWriter.GenerateGltfMaterials(materialLists, textureImageFolder)

	for _, mesh := range zoneMeshes {
		gltfWriter.AddFragmentDataWithOptions(
			mesh,
			ModelGenerationModeCombine,
			false,
			shortName,
			-1,
			nil,
			0,
			true,
		)
	}

	gltfWriter.AddCombinedMeshToScene(true, shortName, "", nil)

	exportFilePath := fmt.Sprintf("%s%s.gltf", exportFolder, shortName)
	return gltfWriter.WriteAssetToFile(exportFilePath, true, "")
}

// ExportStaticActor exports a static actor (mesh-only) to glTF.
func ExportStaticActor(actor *fragments.Actor, settings *ActorExportSettings, exportFolder string) error {
	if actor == nil || actor.MeshReference == nil {
		return nil
	}

	mesh, ok := getMeshFromReference(actor.MeshReference)
	if !ok || mesh == nil {
		return nil
	}

	exportFormat := GltfExportFormatGlTF
	if settings.ExportGltfInGlbFormat {
		exportFormat = GltfExportFormatGlb
	}

	gltfWriter := NewGltfWriter(settings.ExportGltfVertexColors, exportFormat)

	textureImageFolder := exportFolder + "Textures/"
	materialLists := []*fragments.MaterialList{mesh.MaterialList}
	gltfWriter.GenerateGltfMaterials(materialLists, textureImageFolder)
	gltfWriter.AddFragmentData(mesh)

	meshName := helpers.CleanMeshName(mesh.GetName())
	exportFilePath := fmt.Sprintf("%s%s.gltf", exportFolder, meshName)
	return gltfWriter.WriteAssetToFile(exportFilePath, true, "")
}

// ExportSkeletalActor exports a skeletal actor (with animations) to glTF.
func ExportSkeletalActor(actor *fragments.Actor, settings *ActorExportSettings, exportFolder string, wldType wld.WldType) error {
	if actor == nil || actor.SkeletonReference == nil {
		return nil
	}

	skeleton := actor.SkeletonReference.SkeletonHierarchy
	if skeleton == nil {
		return nil
	}

	exportFormat := GltfExportFormatGlTF
	if settings.ExportGltfInGlbFormat {
		exportFormat = GltfExportFormatGlb
	}

	gltfWriter := NewGltfWriter(settings.ExportGltfVertexColors, exportFormat)

	// Collect all material lists
	materialLists := collectSkeletonMaterialLists(skeleton)

	textureImageFolder := exportFolder + "Textures/"
	gltfWriter.GenerateGltfMaterials(materialLists, textureImageFolder)

	isCharacterAnimation := wldType == wld.WldTypeCharacters

	// Add bone meshes
	for i, bone := range skeleton.Skeleton {
		mesh := getMeshFromBone(bone)
		if mesh == nil {
			continue
		}

		// Shift mesh vertices (this would modify the mesh in place in the original)
		// For now, we add the mesh data directly
		gltfWriter.AddFragmentDataWithSkeleton(mesh, skeleton, "", i)
	}

	// Add skeleton meshes
	if skeleton.Meshes != nil {
		for _, meshFrag := range skeleton.Meshes {
			mesh, ok := meshFrag.(*fragments.Mesh)
			if !ok {
				continue
			}
			gltfWriter.AddFragmentDataWithSkeleton(mesh, skeleton, "", -1)
		}

		// Handle secondary meshes
		for i, secondaryMeshFrag := range skeleton.SecondaryMeshes {
			secondaryMesh, ok := secondaryMeshFrag.(*fragments.Mesh)
			if !ok {
				continue
			}

			secondaryGltfWriter := NewGltfWriter(settings.ExportGltfVertexColors, exportFormat)
			secondaryGltfWriter.CopyMaterialList(gltfWriter)

			// Add primary mesh
			if len(skeleton.Meshes) > 0 {
				if primaryMesh, ok := skeleton.Meshes[0].(*fragments.Mesh); ok {
					secondaryGltfWriter.AddFragmentDataWithSkeleton(primaryMesh, skeleton, "", -1)
				}
			}

			// Add secondary mesh
			secondaryGltfWriter.AddFragmentDataWithSkeleton(secondaryMesh, skeleton, "", -1)
			secondaryGltfWriter.ApplyAnimationToSkeleton(skeleton, "pos", isCharacterAnimation, true)

			if settings.ExportAllAnimationFrames {
				for animationKey := range skeleton.Animations {
					secondaryGltfWriter.ApplyAnimationToSkeleton(skeleton, animationKey, isCharacterAnimation, false)
				}
			}

			skeletonName := helpers.CleanSkeletonName(skeleton.GetName())
			secondaryExportPath := fmt.Sprintf("%s%s_%02d.gltf", exportFolder, skeletonName, i)
			if err := secondaryGltfWriter.WriteAssetToFile(secondaryExportPath, true, skeleton.ModelBase); err != nil {
				return fmt.Errorf("failed to write secondary mesh: %w", err)
			}
		}
	}

	// Apply pose animation
	gltfWriter.ApplyAnimationToSkeleton(skeleton, "pos", isCharacterAnimation, true)

	// Export all animations if requested
	if settings.ExportAllAnimationFrames {
		for animationKey := range skeleton.Animations {
			gltfWriter.ApplyAnimationToSkeleton(skeleton, animationKey, isCharacterAnimation, false)
		}
	}

	skeletonName := helpers.CleanSkeletonName(skeleton.GetName())
	exportFilePath := fmt.Sprintf("%s%s.gltf", exportFolder, skeletonName)
	return gltfWriter.WriteAssetToFile(exportFilePath, true, skeleton.ModelBase)
}

// ExportZoneWithObjects exports zone meshes along with object instances.
func ExportZoneWithObjects(
	zoneMeshes []*fragments.Mesh,
	zoneMaterialLists []*fragments.MaterialList,
	actors []*fragments.Actor,
	objectInstances []*fragments.ObjectInstance,
	shortName string,
	exportFolder string,
	settings *ActorExportSettings,
) error {
	if len(zoneMeshes) == 0 {
		return nil
	}

	exportFormat := GltfExportFormatGlTF
	if settings.ExportGltfInGlbFormat {
		exportFormat = GltfExportFormatGlb
	}

	gltfWriter := NewGltfWriter(settings.ExportGltfVertexColors, exportFormat)
	textureImageFolder := exportFolder + "Textures/"

	// Combine all material lists
	allMaterialLists := make([]*fragments.MaterialList, 0)
	allMaterialLists = append(allMaterialLists, zoneMaterialLists...)

	// Add actor material lists
	for _, actor := range actors {
		if actor.ActorType == datatypes.ActorTypeStatic {
			if mesh, ok := getMeshFromReference(actor.MeshReference); ok && mesh != nil && mesh.MaterialList != nil {
				allMaterialLists = append(allMaterialLists, mesh.MaterialList)
			}
		} else if actor.ActorType == datatypes.ActorTypeSkeletal {
			if actor.SkeletonReference != nil && actor.SkeletonReference.SkeletonHierarchy != nil {
				skeletonMats := collectSkeletonMaterialLists(actor.SkeletonReference.SkeletonHierarchy)
				allMaterialLists = append(allMaterialLists, skeletonMats...)
			}
		}
	}

	gltfWriter.GenerateGltfMaterials(allMaterialLists, textureImageFolder)

	// Add zone meshes
	for _, mesh := range zoneMeshes {
		gltfWriter.AddFragmentDataWithOptions(
			mesh,
			ModelGenerationModeCombine,
			false,
			shortName,
			-1,
			nil,
			0,
			true,
		)
	}

	gltfWriter.AddCombinedMeshToScene(true, shortName, "", nil)

	// Add object instances
	for _, actor := range actors {
		if actor.ActorType == datatypes.ActorTypeStatic {
			mesh, ok := getMeshFromReference(actor.MeshReference)
			if !ok || mesh == nil {
				continue
			}

			instances := findObjectInstances(mesh.GetName(), objectInstances)
			instanceIndex := 0
			for _, instance := range instances {
				if instance.Position.Z < -32768 {
					continue
				}

				gltfWriter.AddFragmentDataWithOptions(
					mesh,
					ModelGenerationModeSeparate,
					false,
					"",
					-1,
					instance,
					instanceIndex,
					true,
				)
				instanceIndex++
			}
		} else if actor.ActorType == datatypes.ActorTypeSkeletal {
			skeleton := actor.SkeletonReference.SkeletonHierarchy
			if skeleton == nil {
				continue
			}

			instances := findObjectInstances(skeleton.GetName(), objectInstances)
			combinedMeshName := helpers.CleanSkeletonName(skeleton.GetName())
			instanceIndex := 0
			addedMeshOnce := false

			for _, instance := range instances {
				if instance.Position.Z < -32768 {
					continue
				}

				if !addedMeshOnce ||
					(settings.ExportGltfVertexColors && hasObjectColors(instance)) {

					for i, bone := range skeleton.Skeleton {
						mesh := getMeshFromBone(bone)
						if mesh == nil {
							continue
						}

						gltfWriter.AddFragmentDataWithOptions(
							mesh,
							ModelGenerationModeCombine,
							false,
							combinedMeshName,
							i,
							instance,
							instanceIndex,
							true,
						)
					}
				}

				gltfWriter.AddCombinedMeshToScene(true, combinedMeshName, "", instance)
				addedMeshOnce = true
				instanceIndex++
			}
		}
	}

	exportFilePath := fmt.Sprintf("%s%s.gltf", exportFolder, shortName)
	return gltfWriter.WriteAssetToFile(exportFilePath, true, "")
}

// Helper functions

// collectSkeletonMaterialLists collects all material lists from a skeleton.
func collectSkeletonMaterialLists(skeleton *fragments.SkeletonHierarchy) []*fragments.MaterialList {
	materialListSet := make(map[*fragments.MaterialList]bool)
	var materialLists []*fragments.MaterialList

	// Add material lists from skeleton meshes
	if skeleton.Meshes != nil {
		for _, meshFrag := range skeleton.Meshes {
			if mesh, ok := meshFrag.(*fragments.Mesh); ok && mesh.MaterialList != nil {
				if !materialListSet[mesh.MaterialList] {
					materialListSet[mesh.MaterialList] = true
					materialLists = append(materialLists, mesh.MaterialList)
				}
			}
		}
	}

	// Add material lists from bone meshes
	for _, bone := range skeleton.Skeleton {
		mesh := getMeshFromBone(bone)
		if mesh != nil && mesh.MaterialList != nil {
			if !materialListSet[mesh.MaterialList] {
				materialListSet[mesh.MaterialList] = true
				materialLists = append(materialLists, mesh.MaterialList)
			}
		}
	}

	return materialLists
}

// getMeshFromReference extracts a Mesh from a fragment reference.
func getMeshFromReference(ref fragments.Fragment) (*fragments.Mesh, bool) {
	if ref == nil {
		return nil, false
	}

	// Direct mesh
	if mesh, ok := ref.(*fragments.Mesh); ok {
		return mesh, true
	}

	// Mesh reference
	if meshRef, ok := ref.(*fragments.MeshReference); ok {
		if meshRef.Mesh != nil {
			if mesh, ok := meshRef.Mesh.(*fragments.Mesh); ok {
				return mesh, true
			}
		}
	}

	return nil, false
}

// getMeshFromBone extracts a Mesh from a skeleton bone.
func getMeshFromBone(bone *fragments.SkeletonBone) *fragments.Mesh {
	if bone == nil || bone.MeshReference == nil {
		return nil
	}

	mesh, _ := getMeshFromReference(bone.MeshReference)
	return mesh
}

// findObjectInstances finds object instances matching a name prefix.
func findObjectInstances(objectName string, instances []*fragments.ObjectInstance) []*fragments.ObjectInstance {
	var result []*fragments.ObjectInstance
	cleanName := strings.ToLower(objectName)

	for _, instance := range instances {
		instanceName := strings.ToLower(instance.ObjectName)
		if strings.HasPrefix(cleanName, instanceName) || strings.HasPrefix(instanceName, cleanName) {
			result = append(result, instance)
		}
	}

	return result
}
