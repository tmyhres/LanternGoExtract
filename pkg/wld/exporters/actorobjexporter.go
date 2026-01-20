package exporters

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tmyhres/LanternGoExtract/pkg/wld"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/datatypes"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/helpers"
)

// ActorObjExporterSettings contains configuration for actor OBJ export.
type ActorObjExporterSettings struct {
	// ExportHiddenGeometry exports invisible and boundary surfaces.
	ExportHiddenGeometry bool

	// ExportZoneMeshGroups exports zone meshes as groups.
	ExportZoneMeshGroups bool

	// ExportAllAnimationFrames exports all animation frames.
	ExportAllAnimationFrames bool

	// ExportZoneWithObjects includes object instances in zone export.
	ExportZoneWithObjects bool
}

// ObjBackupVertices stores original vertices for restoration after skeleton transformations.
var ObjBackupVertices = make(map[*fragments.Mesh][]fragments.Vec3)

// ExportActorsToObj exports all actors from a WLD file to OBJ format.
func ExportActorsToObj(
	actors []*fragments.Actor,
	meshes []*fragments.Mesh,
	materialLists []*fragments.MaterialList,
	objectInstances []*fragments.ObjectInstance,
	wldType wld.WldType,
	exportFolder string,
	zoneShortname string,
	settings *ActorObjExporterSettings,
) error {
	// For zone WLD, ignore actors and export all meshes
	if wldType == wld.WldTypeZone {
		return exportZoneToObj(meshes, materialLists, objectInstances, exportFolder, zoneShortname, settings)
	}

	for _, actor := range actors {
		switch actor.ActorType {
		case datatypes.ActorTypeStatic:
			if err := exportStaticActorToObj(actor, exportFolder, settings); err != nil {
				return err
			}
		case datatypes.ActorTypeSkeletal:
			if err := exportSkeletalActorToObj(actor, wldType, exportFolder, settings); err != nil {
				return err
			}
		}
	}

	return nil
}

// exportZoneToObj exports all meshes of a zone in a single OBJ file.
func exportZoneToObj(
	meshes []*fragments.Mesh,
	materialLists []*fragments.MaterialList,
	objectInstances []*fragments.ObjectInstance,
	exportFolder string,
	zoneShortname string,
	settings *ActorObjExporterSettings,
) error {
	if len(meshes) == 0 {
		return nil
	}

	meshWriter := NewMeshObjWriter(datatypes.ObjExportTypeTextured, settings.ExportHiddenGeometry,
		settings.ExportZoneMeshGroups, zoneShortname, "")
	collisionMeshWriter := NewMeshObjWriter(datatypes.ObjExportTypeCollision, settings.ExportHiddenGeometry,
		settings.ExportZoneMeshGroups, zoneShortname, "")
	materialListWriter := NewMeshObjMtlWriter(settings.ExportHiddenGeometry, zoneShortname)

	for _, mesh := range meshes {
		// Find associated objects
		var associatedObjects []*fragments.ObjectInstance
		if settings.ExportZoneWithObjects && objectInstances != nil {
			for _, obj := range objectInstances {
				if obj != nil && !strings.Contains(obj.ObjectName, "door") &&
					strings.HasPrefix(strings.ToLower(mesh.GetName()), strings.ToLower(obj.ObjectName)) {
					associatedObjects = append(associatedObjects, obj)
				}
			}
		}

		// Add mesh for each associated object
		for _, assocObj := range associatedObjects {
			meshWriter.AddFragmentDataWithObject(mesh, assocObj)
		}

		if len(associatedObjects) == 0 {
			meshWriter.AddFragmentData(mesh)
			collisionMeshWriter.AddFragmentData(mesh)
		}
	}

	// Write mesh files
	meshPath := getMeshPath(exportFolder, zoneShortname)
	if err := os.MkdirAll(filepath.Dir(meshPath), 0755); err != nil {
		return err
	}
	if err := meshWriter.WriteAssetToFile(meshPath); err != nil {
		return err
	}

	collisionPath := getCollisionMeshPath(exportFolder, zoneShortname)
	if err := collisionMeshWriter.WriteAssetToFile(collisionPath); err != nil {
		return err
	}

	// Write material list
	for _, materialList := range materialLists {
		materialListWriter.AddFragmentData(materialList)
	}

	if len(materialLists) > 0 {
		materialListName := helpers.CleanName(materialLists[0].GetName(), "MaterialList", true)
		materialPath := getMaterialListPath(exportFolder, materialListName)
		if err := os.MkdirAll(filepath.Dir(materialPath), 0755); err != nil {
			return err
		}
		if err := materialListWriter.WriteAssetToFile(materialPath); err != nil {
			return err
		}
	}

	return nil
}

// exportStaticActorToObj exports a static actor to OBJ format.
func exportStaticActorToObj(actor *fragments.Actor, exportFolder string, settings *ActorObjExporterSettings) error {
	if actor.MeshReference == nil {
		return nil
	}

	var mesh *fragments.Mesh

	// Try direct Mesh type
	if m, ok := actor.MeshReference.(*fragments.Mesh); ok {
		mesh = m
	}

	// Try MeshReference type
	if mesh == nil {
		if meshRef, ok := actor.MeshReference.(*fragments.MeshReference); ok {
			if m, ok := meshRef.Mesh.(*fragments.Mesh); ok {
				mesh = m
			}
		}
	}

	if mesh == nil {
		return nil
	}

	meshWriter := NewMeshObjWriter(datatypes.ObjExportTypeTextured, settings.ExportHiddenGeometry,
		settings.ExportZoneMeshGroups, "", "")
	collisionMeshWriter := NewMeshObjWriter(datatypes.ObjExportTypeCollision, settings.ExportHiddenGeometry,
		settings.ExportZoneMeshGroups, "", "")
	materialListWriter := NewMeshObjMtlWriter(settings.ExportHiddenGeometry, "")

	meshWriter.AddFragmentData(mesh)
	meshName := helpers.CleanName(mesh.GetName(), "Mesh", true)
	meshPath := getMeshPath(exportFolder, meshName)

	if err := os.MkdirAll(filepath.Dir(meshPath), 0755); err != nil {
		return err
	}
	if err := meshWriter.WriteAssetToFile(meshPath); err != nil {
		return err
	}

	if mesh.ExportSeparateCollision {
		collisionMeshWriter.AddFragmentData(mesh)
		collisionPath := getCollisionMeshPath(exportFolder, meshName)
		if err := collisionMeshWriter.WriteAssetToFile(collisionPath); err != nil {
			return err
		}
	}

	if mesh.MaterialList != nil {
		materialListWriter.AddFragmentData(mesh.MaterialList)
		materialListName := helpers.CleanName(mesh.MaterialList.GetName(), "MaterialList", true)
		materialPath := getMaterialListPath(exportFolder, materialListName)
		if err := os.MkdirAll(filepath.Dir(materialPath), 0755); err != nil {
			return err
		}
		if err := materialListWriter.WriteAssetToFile(materialPath); err != nil {
			return err
		}
	}

	return nil
}

// exportSkeletalActorToObj exports a skeletal actor to OBJ format.
func exportSkeletalActorToObj(actor *fragments.Actor, wldType wld.WldType, exportFolder string, settings *ActorObjExporterSettings) error {
	if actor.SkeletonReference == nil {
		return nil
	}

	skeleton := actor.SkeletonReference.SkeletonHierarchy
	if skeleton == nil {
		return nil
	}

	// Export animation frames
	if settings.ExportAllAnimationFrames {
		for animName, animation := range skeleton.Animations {
			for frameIdx := 0; frameIdx < animation.FrameCount; frameIdx++ {
				if err := writeObjAnimationFrame(skeleton, animName, frameIdx, wldType, exportFolder, settings); err != nil {
					return err
				}
			}
		}
	} else {
		if err := writeObjAnimationFrame(skeleton, "pos", 0, wldType, exportFolder, settings); err != nil {
			return err
		}
	}

	// Export material lists
	var materialLists []*fragments.MaterialList

	// Get material list from meshes
	if len(skeleton.Meshes) > 0 {
		if mesh, ok := skeleton.Meshes[0].(*fragments.Mesh); ok && mesh.MaterialList != nil {
			materialLists = append(materialLists, mesh.MaterialList)
		}
	}

	// Get material lists from skeleton bones
	for _, bone := range skeleton.Skeleton {
		if bone.MeshReference != nil {
			if meshRef, ok := bone.MeshReference.(*fragments.MeshReference); ok {
				if meshRef.Mesh != nil {
					if mesh, ok := meshRef.Mesh.(*fragments.Mesh); ok && mesh.MaterialList != nil {
						found := false
						for _, ml := range materialLists {
							if ml == mesh.MaterialList {
								found = true
								break
							}
						}
						if !found {
							materialLists = append(materialLists, mesh.MaterialList)
						}
					}
				}
			}
		}
	}

	materialListWriter := NewMeshObjMtlWriter(settings.ExportHiddenGeometry, "")

	for _, ml := range materialLists {
		materialListWriter.AddFragmentData(ml)

		var savePath string
		materialListName := helpers.CleanName(ml.GetName(), "MaterialList", true)

		if wldType == wld.WldTypeCharacters {
			savePath = getMaterialListPathWithSkin(exportFolder, materialListName, 0)
		} else {
			savePath = getMaterialListPath(exportFolder, materialListName)
		}

		if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
			return err
		}
		if err := materialListWriter.WriteAssetToFile(savePath); err != nil {
			return err
		}

		// Export skin variants
		for i := 0; i < ml.VariantCount; i++ {
			materialListWriter.ClearExportData()
			materialListWriter.SetSkinID(i + 1)
			materialListWriter.AddFragmentData(ml)

			variantPath := getMaterialListPathWithSkin(exportFolder, materialListName, i+1)
			if err := materialListWriter.WriteAssetToFile(variantPath); err != nil {
				return err
			}
		}

		materialListWriter.ClearExportData()
	}

	return nil
}

// writeObjAnimationFrame writes a single animation frame to OBJ.
func writeObjAnimationFrame(
	skeleton *fragments.SkeletonHierarchy,
	animation string,
	frameIndex int,
	wldType wld.WldType,
	exportFolder string,
	settings *ActorObjExporterSettings,
) error {
	meshWriter := NewMeshObjWriter(datatypes.ObjExportTypeTextured, settings.ExportHiddenGeometry,
		settings.ExportZoneMeshGroups, "", "")
	meshWriter.SetIsCharacterModel(wldType == wld.WldTypeCharacters)

	// Add bone meshes
	for _, bone := range skeleton.Skeleton {
		if bone.MeshReference != nil {
			if meshRef, ok := bone.MeshReference.(*fragments.MeshReference); ok {
				if meshRef.Mesh != nil {
					meshWriter.AddFragmentData(meshRef.Mesh)
				}
			}
		}
	}

	// Add skeleton meshes
	if skeleton.Meshes != nil {
		for _, meshFrag := range skeleton.Meshes {
			if mesh, ok := meshFrag.(*fragments.Mesh); ok {
				// In full implementation, would call MeshExportHelper.ShiftMeshVertices here
				// ObjBackupVertices[mesh] = mesh.Vertices // backup
				meshWriter.AddFragmentData(mesh)
			}
		}

		// Export secondary meshes
		for i, secMesh := range skeleton.SecondaryMeshes {
			if mesh, ok := secMesh.(*fragments.Mesh); ok {
				meshWriter2 := NewMeshObjWriter(datatypes.ObjExportTypeTextured, settings.ExportHiddenGeometry,
					settings.ExportZoneMeshGroups, "", "")
				meshWriter2.SetIsCharacterModel(true)

				if len(skeleton.Meshes) > 0 {
					meshWriter2.AddFragmentData(skeleton.Meshes[0])
				}
				// ObjBackupVertices[mesh] = mesh.Vertices // backup
				meshWriter2.AddFragmentData(mesh)

				skeletonName := helpers.CleanName(skeleton.GetName(), "SkeletonHierarchy", true)
				meshPath := getMeshPathWithVariant(exportFolder, skeletonName, i+1)
				if err := os.MkdirAll(filepath.Dir(meshPath), 0755); err != nil {
					return err
				}
				if err := meshWriter2.WriteAssetToFile(meshPath); err != nil {
					return err
				}
			}
		}
	}

	// Determine filename
	var fileName string
	skeletonName := helpers.CleanName(skeleton.GetName(), "SkeletonHierarchy", true)
	if settings.ExportAllAnimationFrames {
		fileName = fmt.Sprintf("%s_%s_%d", skeletonName, animation, frameIndex)
	} else {
		fileName = skeletonName
	}

	meshPath := getMeshPath(exportFolder, fileName)
	if err := os.MkdirAll(filepath.Dir(meshPath), 0755); err != nil {
		return err
	}
	if err := meshWriter.WriteAssetToFile(meshPath); err != nil {
		return err
	}

	// Restore vertices
	restoreVertices()

	return nil
}

// restoreVertices restores backed up vertices.
func restoreVertices() {
	for mesh, vertices := range ObjBackupVertices {
		mesh.Vertices = vertices
	}
	ObjBackupVertices = make(map[*fragments.Mesh][]fragments.Vec3)
}

// getMeshPath returns the OBJ mesh file path.
func getMeshPath(exportFolder, meshName string) string {
	return exportFolder + meshName + ".obj"
}

// getMeshPathWithVariant returns the OBJ mesh file path with a variant index.
func getMeshPathWithVariant(exportFolder, meshName string, variant int) string {
	return fmt.Sprintf("%s%s_%d.obj", exportFolder, meshName, variant)
}

// getCollisionMeshPath returns the collision mesh file path.
func getCollisionMeshPath(exportFolder, meshName string) string {
	return exportFolder + meshName + "_collision.obj"
}

// getMaterialListPath returns the material list file path.
func getMaterialListPath(exportFolder, materialListName string) string {
	return exportFolder + "/" + materialListName + ".mtl"
}

// getMaterialListPathWithSkin returns the material list file path with a skin index.
func getMaterialListPathWithSkin(exportFolder, materialListName string, skinIndex int) string {
	return fmt.Sprintf("%s/%s_%d.mtl", exportFolder, materialListName, skinIndex)
}
