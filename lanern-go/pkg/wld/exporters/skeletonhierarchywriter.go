package exporters

import (
	"strconv"
	"strings"

	"github.com/lanterneq/lanern-go/pkg/wld/fragments"
	"github.com/lanterneq/lanern-go/pkg/wld/helpers"
)

// SkeletonHierarchyWriter exports skeleton hierarchy data to a text format.
type SkeletonHierarchyWriter struct {
	TextAssetWriter
	stripModelBase bool
}

// NewSkeletonHierarchyWriter creates a new SkeletonHierarchyWriter.
func NewSkeletonHierarchyWriter(stripModelBase bool) *SkeletonHierarchyWriter {
	return &SkeletonHierarchyWriter{
		stripModelBase: stripModelBase,
	}
}

// AddFragmentData adds skeleton hierarchy fragment data to the export buffer.
func (w *SkeletonHierarchyWriter) AddFragmentData(data fragments.Fragment) {
	w.AppendLine(ExportHeaderTitle + "Skeleton Hierarchy")
	w.AppendLine(ExportHeaderFormat + "BoneName, Children, Mesh, AlternateMesh, ParticleCloud")

	skeleton, ok := data.(*fragments.SkeletonHierarchy)
	if !ok || skeleton == nil {
		return
	}

	// Write primary meshes
	if skeleton.Meshes != nil && len(skeleton.Meshes) != 0 {
		w.AppendString("meshes")
		for _, mesh := range skeleton.Meshes {
			w.AppendString(",")
			w.AppendString(cleanMeshName(mesh))
		}
		w.AppendString("\n")

		w.AppendString("secondary_meshes")
		for _, mesh := range skeleton.SecondaryMeshes {
			w.AppendString(",")
			w.AppendString(cleanMeshName(mesh))
		}
		w.AppendString("\n")
	}

	// Write alternate meshes
	if skeleton.AlternateMeshes != nil && len(skeleton.AlternateMeshes) != 0 {
		w.AppendString("meshes")
		for _, mesh := range skeleton.AlternateMeshes {
			w.AppendString(",")
			w.AppendString(cleanMeshName(mesh))
		}
		w.AppendString("\n")

		w.AppendString("secondary_meshes")
		for _, mesh := range skeleton.SecondaryAlternateMeshes {
			w.AppendString(",")
			w.AppendString(cleanMeshName(mesh))
		}
		w.AppendString("\n")
	}

	// Write skeleton nodes
	for _, node := range skeleton.Skeleton {
		// Build children list
		childrenParts := make([]string, len(node.Children))
		for i, child := range node.Children {
			childrenParts[i] = strconv.Itoa(child)
		}
		childrenList := strings.Join(childrenParts, ";")

		boneName := node.CleanedName
		if w.stripModelBase {
			boneName = w.stripModelBaseFromName(boneName, skeleton.ModelBase)
		}

		w.AppendString(cleanSkeletonNodeName(boneName))
		w.AppendString(",")
		w.AppendString(childrenList)

		w.AppendString(",")

		// Mesh reference
		if node.MeshReference != nil {
			if mesh, ok := node.MeshReference.(*fragments.Mesh); ok && mesh != nil {
				w.AppendString(helpers.CleanName(mesh.GetName(), "Mesh", true))
			}
		}

		w.AppendString(",")

		// Legacy mesh reference - stored in MeshReference for bones
		if node.MeshReference != nil {
			if legacyMesh, ok := node.MeshReference.(*fragments.LegacyMesh); ok && legacyMesh != nil {
				w.AppendString(helpers.CleanName(legacyMesh.GetName(), "LegacyMesh", true))
			}
		}

		w.AppendString(",")

		// Particle cloud
		if node.ParticleCloud != nil {
			if particleCloud, ok := node.ParticleCloud.(*fragments.ParticleCloud); ok && particleCloud != nil {
				w.AppendString(helpers.CleanName(particleCloud.GetName(), "ParticleCloud", true))
			}
		}

		w.AppendString("\n")
	}
}

// cleanSkeletonNodeName cleans a skeleton node name.
func cleanSkeletonNodeName(name string) string {
	name = strings.ReplaceAll(name, "_DAG", "")
	return strings.ToLower(name)
}

// stripModelBaseFromName strips the model base from a bone name.
func (w *SkeletonHierarchyWriter) stripModelBaseFromName(boneName, modelBase string) string {
	if strings.HasPrefix(boneName, modelBase) {
		boneName = boneName[len(modelBase):]
	}

	if boneName == "" {
		boneName = "root"
	}

	return boneName
}

// cleanMeshName returns the cleaned name for a mesh fragment.
func cleanMeshName(mesh fragments.Fragment) string {
	if mesh == nil {
		return ""
	}

	if m, ok := mesh.(*fragments.Mesh); ok {
		return helpers.CleanName(m.GetName(), "Mesh", true)
	}
	if lm, ok := mesh.(*fragments.LegacyMesh); ok {
		return helpers.CleanName(lm.GetName(), "LegacyMesh", true)
	}

	return helpers.CleanName(mesh.GetName(), "", true)
}
