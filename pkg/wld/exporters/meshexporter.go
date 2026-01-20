package exporters

import (
	"os"
	"path/filepath"

	"github.com/tmyhres/LanternGoExtract/pkg/infrastructure/logger"
	"github.com/tmyhres/LanternGoExtract/pkg/wld"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/helpers"
)

// MeshExporterSettings contains configuration for mesh export.
type MeshExporterSettings struct {
	// ExportZoneMeshGroups exports zone meshes as groups.
	ExportZoneMeshGroups bool

	// ExportCharactersToSingleFolder exports all characters to a single folder.
	ExportCharactersToSingleFolder bool
}

// WldFileInterface defines the minimal interface needed for mesh export.
type WldFileInterface interface {
	GetWldType() wld.WldType
	GetExportFolderForWldType() string
	GetZoneName() string
}

// ExportMeshes exports all meshes from a WLD file.
func ExportMeshes(
	meshes []*fragments.Mesh,
	legacyMeshes []*fragments.LegacyMesh,
	materialLists []*fragments.MaterialList,
	wldType wld.WldType,
	exportFolder string,
	zoneShortname string,
	settings *MeshExporterSettings,
	log logger.Logger,
) error {
	meshFolder := "Meshes/"
	legacyMeshFolder := "AlternateMeshes/"

	// Merge folders by default
	mergeMeshFolders := true
	if mergeMeshFolders {
		legacyMeshFolder = meshFolder
	}

	if len(meshes) == 0 && len(legacyMeshes) == 0 {
		return nil
	}

	meshWriter := NewMeshIntermediateAssetWriter(settings.ExportZoneMeshGroups, false)
	legacyMeshWriter := NewLegacyMeshIntermediateAssetWriter(settings.ExportZoneMeshGroups, false)
	collisionMeshWriter := NewMeshIntermediateAssetWriter(settings.ExportZoneMeshGroups, true)
	collisionLegacyMeshWriter := NewLegacyMeshIntermediateAssetWriter(settings.ExportZoneMeshGroups, true)
	materialListWriter := NewMeshIntermediateMaterialsWriter()

	exportEachPass := wldType != wld.WldTypeZone || settings.ExportZoneMeshGroups

	// Determine if we need to export collision mesh for zones
	exportCollisionMesh := false
	if !exportEachPass {
		for _, m := range meshes {
			if m.ExportSeparateCollision {
				exportCollisionMesh = true
				break
			}
		}
		if !exportCollisionMesh {
			for _, m := range legacyMeshes {
				if m.ExportSeparateCollision {
					exportCollisionMesh = true
					break
				}
			}
		}
	}

	// Export materials
	if materialLists != nil {
		for _, materialList := range materialLists {
			materialListWriter.AddFragmentData(materialList)

			newExportFolder := exportFolder + "/MaterialLists/"
			materialListName := helpers.CleanName(materialList.GetName(), "MaterialList", true)
			filePath := newExportFolder + materialListName + ".txt"

			if !exportEachPass {
				continue
			}

			if settings.ExportCharactersToSingleFolder && wldType == wld.WldTypeCharacters {
				if fileExists(filePath) {
					oldContent, err := os.ReadFile(filePath)
					if err == nil {
						oldFileSize := len(oldContent)
						newFileSize := materialListWriter.GetExportByteCount()

						if newFileSize <= oldFileSize {
							materialListWriter.ClearExportData()
							continue
						}
					}
				}
			}

			materialList.HasBeenExported = true
			if err := os.MkdirAll(newExportFolder, 0755); err != nil {
				return err
			}
			if err := materialListWriter.WriteAssetToFile(filePath); err != nil {
				return err
			}
			materialListWriter.ClearExportData()
		}
	}

	if !exportEachPass {
		filePath := exportFolder + "/MaterialLists/" + zoneShortname + ".txt"
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}
		if err := materialListWriter.WriteAssetToFile(filePath); err != nil {
			return err
		}
	}

	// Export legacy meshes
	if legacyMeshes != nil {
		for _, alternateMesh := range legacyMeshes {
			legacyMeshWriter.AddFragmentData(alternateMesh)

			// Determine if we need collision
			if exportEachPass {
				exportCollisionMesh = alternateMesh.ExportSeparateCollision
			}

			if exportCollisionMesh {
				collisionLegacyMeshWriter.AddFragmentData(alternateMesh)
			}

			if exportEachPass {
				newExportFolder := exportFolder + legacyMeshFolder
				if err := os.MkdirAll(newExportFolder, 0755); err != nil {
					return err
				}

				meshName := helpers.CleanName(alternateMesh.GetName(), "LegacyMesh", true)
				if err := legacyMeshWriter.WriteAssetToFile(newExportFolder + meshName + ".txt"); err != nil {
					return err
				}
				legacyMeshWriter.ClearExportData()

				if exportCollisionMesh {
					if err := collisionLegacyMeshWriter.WriteAssetToFile(exportFolder + legacyMeshFolder + meshName + "_collision.txt"); err != nil {
						return err
					}
					collisionLegacyMeshWriter.ClearExportData()
				}
			}
		}
	}

	// Export meshes
	for _, mesh := range meshes {
		if mesh == nil {
			continue
		}

		meshWriter.AddFragmentData(mesh)

		// Determine if we need collision
		if exportEachPass {
			exportCollisionMesh = mesh.ExportSeparateCollision
		}

		if exportCollisionMesh {
			collisionMeshWriter.AddFragmentData(mesh)
		}

		if exportEachPass {
			// Skip if material list hasn't been exported for characters
			if wldType == wld.WldTypeCharacters && settings.ExportCharactersToSingleFolder {
				if mesh.MaterialList != nil && !mesh.MaterialList.HasBeenExported {
					meshWriter.ClearExportData()
					collisionMeshWriter.ClearExportData()
					continue
				}
			}

			meshName := helpers.CleanName(mesh.GetName(), "Mesh", true)
			meshPath := exportFolder + meshFolder + meshName + ".txt"
			if err := os.MkdirAll(filepath.Dir(meshPath), 0755); err != nil {
				return err
			}
			if err := meshWriter.WriteAssetToFile(meshPath); err != nil {
				return err
			}
			meshWriter.ClearExportData()

			if exportCollisionMesh {
				collisionPath := exportFolder + meshFolder + meshName + "_collision.txt"
				if err := collisionMeshWriter.WriteAssetToFile(collisionPath); err != nil {
					return err
				}
				collisionMeshWriter.ClearExportData()
			}
		}
	}

	if !exportEachPass {
		legacyPath := exportFolder + legacyMeshFolder + zoneShortname + ".txt"
		meshPath := exportFolder + meshFolder + zoneShortname + ".txt"

		if err := os.MkdirAll(filepath.Dir(legacyPath), 0755); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(meshPath), 0755); err != nil {
			return err
		}

		if err := legacyMeshWriter.WriteAssetToFile(legacyPath); err != nil {
			return err
		}
		if err := meshWriter.WriteAssetToFile(meshPath); err != nil {
			return err
		}

		if exportCollisionMesh {
			collisionFileName := zoneShortname + "_collision.txt"
			if err := collisionLegacyMeshWriter.WriteAssetToFile(exportFolder + legacyMeshFolder + collisionFileName); err != nil {
				return err
			}
			if err := collisionMeshWriter.WriteAssetToFile(exportFolder + meshFolder + collisionFileName); err != nil {
				return err
			}
		}
	}

	return nil
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
