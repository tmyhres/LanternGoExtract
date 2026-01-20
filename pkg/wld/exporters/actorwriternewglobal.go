package exporters

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tmyhres/LanternGoExtract/pkg/wld/datatypes"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/helpers"
)

// ActorWriterNewGlobal exports actor data globally, merging with existing files.
type ActorWriterNewGlobal struct {
	TextAssetWriter
	actorType  datatypes.ActorType
	actorCount int
	objects    []string
	objectSet  map[string]bool
}

// NewActorWriterNewGlobal creates a new ActorWriterNewGlobal for the specified actor type.
func NewActorWriterNewGlobal(actorType datatypes.ActorType, rootExportFolder string) *ActorWriterNewGlobal {
	w := &ActorWriterNewGlobal{
		actorType: actorType,
		objects:   make([]string, 0),
		objectSet: make(map[string]bool),
	}

	// Try to read existing file
	filePath := filepath.Join(rootExportFolder, "actors_"+strings.ToLower(actorTypeToString(actorType))+".txt")
	file, err := os.Open(filePath)
	if err != nil {
		return w
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		w.objects = append(w.objects, line)
		w.objectSet[line] = true
	}

	return w
}

// AddFragmentData adds actor fragment data to the export buffer.
func (w *ActorWriterNewGlobal) AddFragmentData(data fragments.Fragment) {
	actor, ok := data.(*fragments.Actor)
	if !ok || actor == nil {
		return
	}

	if actor.ActorType != w.actorType {
		return
	}

	var newActor strings.Builder

	switch w.actorType {
	case datatypes.ActorTypeSkeletal:
		newActor.WriteString(helpers.CleanName(actor.GetName(), "Actor", true))
		newActor.WriteString(",")
		if actor.SkeletonReference != nil && actor.SkeletonReference.SkeletonHierarchy != nil {
			newActor.WriteString(helpers.CleanName(actor.SkeletonReference.SkeletonHierarchy.GetName(), "SkeletonHierarchy", true))
		}

	case datatypes.ActorTypeStatic:
		newActor.WriteString(helpers.CleanName(actor.GetName(), "Actor", true))
		newActor.WriteString(",")

		if meshRef, ok := actor.MeshReference.(*fragments.Mesh); ok && meshRef != nil {
			newActor.WriteString(helpers.CleanName(meshRef.GetName(), "Mesh", true))
		} else if legacyMeshRef, ok := actor.MeshReference.(*fragments.LegacyMesh); ok && legacyMeshRef != nil {
			newActor.WriteString(helpers.CleanName(legacyMeshRef.GetName(), "LegacyMesh", true))
		}

	default:
		newActor.WriteString(helpers.CleanName(actor.GetName(), "Actor", true))
	}

	actorStr := newActor.String()
	if w.objectSet[actorStr] {
		return
	}

	w.objects = append(w.objects, actorStr)
	w.objectSet[actorStr] = true
	w.actorCount++
}

// WriteAssetToFile writes the sorted actor data to a file.
func (w *ActorWriterNewGlobal) WriteAssetToFile(fileName string) error {
	sort.Strings(w.objects)

	w.ClearExportData()
	for _, obj := range w.objects {
		w.AppendLine(obj)
	}

	return w.TextAssetWriter.WriteAssetToFile(fileName)
}

// actorTypeToString converts an ActorType to its string representation.
func actorTypeToString(actorType datatypes.ActorType) string {
	switch actorType {
	case datatypes.ActorTypeCamera:
		return "Camera"
	case datatypes.ActorTypeStatic:
		return "Static"
	case datatypes.ActorTypeSkeletal:
		return "Skeletal"
	case datatypes.ActorTypeParticle:
		return "Particle"
	case datatypes.ActorTypeSprite:
		return "Sprite"
	default:
		return "Unknown"
	}
}
