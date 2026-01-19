package fragments

import (
	"fmt"
	"strings"

	"github.com/lanterneq/lanern-go/pkg/wld/datatypes"
)

// SkeletonHierarchy (0x10)
// Internal name: _HS_DEF
// Describes the layout of a complete skeleton and which pieces connect to each other.
type SkeletonHierarchy struct {
	BaseFragment

	// Meshes contains mesh references for the skeleton.
	Meshes []Fragment

	// AlternateMeshes contains legacy mesh references.
	AlternateMeshes []Fragment

	// Skeleton contains the skeleton bones.
	Skeleton []*SkeletonBone

	// ModelBase is the base model name.
	ModelBase string

	// IsAssigned indicates if this skeleton is assigned to an actor.
	IsAssigned bool

	// Animations contains animation data keyed by animation name.
	Animations map[string]*datatypes.Animation

	// BoneMappingClean maps bone index to cleaned bone name.
	BoneMappingClean map[int]string

	// BoneMapping maps bone index to bone name.
	BoneMapping map[int]string

	// BoundingRadius is the sphere radius for frustum culling.
	BoundingRadius float32

	// SecondaryMeshes contains additional mesh references.
	SecondaryMeshes []Fragment

	// SecondaryAlternateMeshes contains additional legacy mesh references.
	SecondaryAlternateMeshes []Fragment

	// Fragment18Reference is a reference to a polyhedron fragment.
	fragment18Reference Fragment

	// skeletonPieceDictionary maps piece names to bones.
	skeletonPieceDictionary map[string]*SkeletonBone

	// hasBuiltData indicates if skeleton data has been built.
	hasBuiltData bool
}

// SkeletonBone represents a bone in the skeleton hierarchy.
type SkeletonBone struct {
	Index           int
	Name            string
	FullPath        string
	CleanedName     string
	CleanedFullPath string
	Children        []int
	Track           *TrackFragment
	MeshReference   Fragment
	ParticleCloud   Fragment
	AnimationTracks map[string]*TrackFragment
	Parent          *SkeletonBone
}

// FragmentType returns the fragment type ID.
func (f *SkeletonHierarchy) FragmentType() uint32 {
	return 0x10
}

// Initialize parses the fragment data.
func (f *SkeletonHierarchy) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)

	f.Skeleton = make([]*SkeletonBone, 0)
	f.Meshes = make([]Fragment, 0)
	f.AlternateMeshes = make([]Fragment, 0)
	f.skeletonPieceDictionary = make(map[string]*SkeletonBone)
	f.Animations = make(map[string]*datatypes.Animation)
	f.BoneMappingClean = make(map[int]string)
	f.BoneMapping = make(map[int]string)
	f.SecondaryMeshes = make([]Fragment, 0)
	f.SecondaryAlternateMeshes = make([]Fragment, 0)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)
	f.ModelBase = CleanSkeletonName(f.Name)

	// Always 2 when used in main zone, and object files.
	// This means, it has a bounding radius
	// Some differences in character + model archives
	flags, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	hasUnknownParams := IsBitSet(flags, 0)
	hasBoundingRadius := IsBitSet(flags, 1)
	hasMeshReferences := IsBitSet(flags, 9)

	// Number of bones in the skeleton
	boneCount, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read bone count: %w", err)
	}

	// Fragment 18 reference
	fragment18Reference, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read fragment18 reference: %w", err)
	}

	fragIndex := int(fragment18Reference) - 1
	if fragIndex > 0 && fragIndex < len(fragments) {
		f.fragment18Reference = fragments[fragIndex]
	}

	// Three sequential DWORDs
	// This will never be hit for object animations.
	if hasUnknownParams {
		err = r.Skip(3 * 4)
		if err != nil {
			return fmt.Errorf("failed to skip unknown params: %w", err)
		}
	}

	// This is the sphere radius checked against the frustum to cull this object
	if hasBoundingRadius {
		f.BoundingRadius, err = r.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read bounding radius: %w", err)
		}
	}

	for i := int32(0); i < boneCount; i++ {
		// An index into the string hash to get this bone's name
		boneNameIndex, err := r.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read bone name index: %w", err)
		}

		boneName := ""
		if s, ok := stringHash[-int(boneNameIndex)]; ok {
			boneName = s
		}

		// Always 0 for object bones - confirmed
		_, err = r.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read bone flags: %w", err)
		}

		// Reference to a bone track - confirmed, is never a bad reference
		trackReferenceIndex, err := r.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read track reference index: %w", err)
		}

		var track *TrackFragment
		trackIdx := int(trackReferenceIndex) - 1
		if trackIdx >= 0 && trackIdx < len(fragments) {
			if t, ok := fragments[trackIdx].(*TrackFragment); ok {
				track = t
			}
		}

		if track != nil {
			f.addPoseTrack(track, boneName)
		}

		pieceNew := &SkeletonBone{
			Index:           int(i),
			Track:           track,
			Name:            boneName,
			AnimationTracks: make(map[string]*TrackFragment),
		}

		if track != nil {
			track.IsPoseAnimation = true
		}

		f.BoneMappingClean[int(i)] = datatypes.CleanBoneAndStripBase(boneName, f.ModelBase)
		f.BoneMapping[int(i)] = boneName

		if track == nil {
			// Log error: Unable to link track reference!
		}

		meshReferenceIndex, err := r.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read mesh reference index: %w", err)
		}

		meshRefIdx := int(meshReferenceIndex) - 1
		if meshRefIdx < 0 {
			// Name lookup (unused in this implementation)
			// _ = stringHash[-meshRefIdx-1]
		} else if meshRefIdx != 0 && meshRefIdx < len(fragments) {
			pieceNew.MeshReference = fragments[meshRefIdx]
			// Note: In the original C#, this checks for MeshReference and ParticleCloud types
			// and handles root bone renaming. That would require additional type definitions.
		}

		childCount, err := r.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read child count: %w", err)
		}

		pieceNew.Children = make([]int, 0, childCount)

		for j := int32(0); j < childCount; j++ {
			childIndex, err := r.ReadInt32()
			if err != nil {
				return fmt.Errorf("failed to read child index: %w", err)
			}
			pieceNew.Children = append(pieceNew.Children, int(childIndex))
		}

		f.Skeleton = append(f.Skeleton, pieceNew)

		if pieceNew.Name != "" {
			if _, exists := f.skeletonPieceDictionary[pieceNew.Name]; !exists {
				f.skeletonPieceDictionary[pieceNew.Name] = pieceNew
			}
		}
	}

	// Read in mesh references
	// All meshes will have vertex bone assignments
	if hasMeshReferences {
		size2, err := r.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read size2: %w", err)
		}

		for i := int32(0); i < size2; i++ {
			meshRefIndex, err := r.ReadInt32()
			if err != nil {
				return fmt.Errorf("failed to read mesh ref index: %w", err)
			}

			meshIdx := int(meshRefIndex) - 1
			if meshIdx >= 0 && meshIdx < len(fragments) {
				// Store mesh reference - type checking would be done at a higher level
				f.Meshes = append(f.Meshes, fragments[meshIdx])
			}
		}

		// Read unknown data
		for i := int32(0); i < size2; i++ {
			_, err = r.ReadInt32()
			if err != nil {
				return fmt.Errorf("failed to read unknown data: %w", err)
			}
		}
	}

	return nil
}

// addPoseTrack adds a pose track to the animations.
func (f *SkeletonHierarchy) addPoseTrack(track *TrackFragment, pieceName string) {
	if _, exists := f.Animations["pos"]; !exists {
		f.Animations["pos"] = datatypes.NewAnimation()
	}

	// Note: The original C# uses Animation.AddTrack which requires TrackFragment interface
	// This is a simplified version
	if track.TrackDefFragment != nil {
		track.TrackDefFragment.IsAssigned = true
	}
	track.IsProcessed = true
	track.IsPoseAnimation = true
}

// BuildSkeletonData builds the skeleton tree data.
func (f *SkeletonHierarchy) BuildSkeletonData(stripModelBase bool) {
	if f.hasBuiltData {
		return
	}

	f.buildSkeletonTreeData(0, nil, "", "", "", stripModelBase)
	f.hasBuiltData = true
}

// buildSkeletonTreeData recursively builds the skeleton tree.
func (f *SkeletonHierarchy) buildSkeletonTreeData(index int, parent *SkeletonBone, runningName, runningNameCleaned, runningIndex string, stripModelBase bool) {
	if index >= len(f.Skeleton) {
		return
	}

	bone := f.Skeleton[index]
	bone.Parent = parent
	bone.CleanedName = f.cleanBoneName(bone.Name, stripModelBase)
	f.BoneMappingClean[index] = bone.CleanedName

	if bone.Name != "" {
		runningIndex += fmt.Sprintf("%d/", bone.Index)
	}

	runningName += bone.Name
	runningNameCleaned += bone.CleanedName

	bone.FullPath = runningName
	bone.CleanedFullPath = runningNameCleaned

	if len(bone.Children) == 0 {
		return
	}

	runningName += "/"
	runningNameCleaned += "/"

	for _, childIndex := range bone.Children {
		f.buildSkeletonTreeData(childIndex, bone, runningName, runningNameCleaned, runningIndex, stripModelBase)
	}
}

// cleanBoneName cleans a bone name.
func (f *SkeletonHierarchy) cleanBoneName(nodeName string, stripModelBase bool) string {
	nodeName = strings.ReplaceAll(nodeName, "_DAG", "")
	nodeName = strings.ToLower(nodeName)
	if stripModelBase {
		nodeName = strings.ReplaceAll(nodeName, f.ModelBase, "")
	}

	if len(nodeName) == 0 {
		return "root"
	}
	return nodeName
}

// CleanSkeletonName cleans a skeleton name by removing the _HS_DEF suffix.
func CleanSkeletonName(name string) string {
	cleanedName := strings.ToLower(name)
	cleanedName = strings.TrimSuffix(cleanedName, "_hs_def")
	return strings.TrimSpace(cleanedName)
}

// IsValidSkeleton checks if a track name belongs to this skeleton.
func (f *SkeletonHierarchy) IsValidSkeleton(trackName string) (bool, string) {
	track := ""
	if len(trackName) > 3 {
		track = trackName[3:]
	}

	if trackName == f.ModelBase {
		return true, f.ModelBase
	}

	for _, bone := range f.Skeleton {
		cleanBoneName := strings.ToLower(strings.ReplaceAll(bone.Name, "_DAG", ""))
		if cleanBoneName == track {
			return true, strings.ToLower(bone.Name)
		}
	}

	return false, ""
}

// RenameNodeBase renames the model base in all nodes.
func (f *SkeletonHierarchy) RenameNodeBase(newBase string) {
	for _, node := range f.Skeleton {
		node.Name = strings.ReplaceAll(node.Name, strings.ToUpper(f.ModelBase), strings.ToUpper(newBase))
	}

	newNameMapping := make(map[int]string)
	for k, v := range f.BoneMapping {
		newNameMapping[k] = strings.ReplaceAll(v, strings.ToUpper(f.ModelBase), strings.ToUpper(newBase))
	}

	f.BoneMapping = newNameMapping
	f.ModelBase = newBase
}

// AddTrack adds a track to the skeleton's animations.
func (f *SkeletonHierarchy) AddTrack(track *TrackFragment) {
	if track == nil {
		return
	}

	animName := track.AnimationName
	if animName == "" {
		animName = "default"
	}

	if _, exists := f.Animations[animName]; !exists {
		f.Animations[animName] = datatypes.NewAnimation()
	}

	// Mark track as processed
	track.IsProcessed = true
	if track.TrackDefFragment != nil {
		track.TrackDefFragment.IsAssigned = true
	}
}

// AddSecondaryMesh adds a mesh to the secondary meshes list.
func (f *SkeletonHierarchy) AddSecondaryMesh(mesh Fragment) {
	if mesh == nil {
		return
	}
	f.SecondaryMeshes = append(f.SecondaryMeshes, mesh)

	// Mark mesh as handled if it supports that
	if m, ok := mesh.(*Mesh); ok {
		m.IsHandled = true
	}
}

// AddTrackEquipment adds a track for equipment with a specific bone name.
func (f *SkeletonHierarchy) AddTrackEquipment(track *TrackFragment, boneName string) {
	if track == nil {
		return
	}

	animName := track.AnimationName
	if animName == "" {
		animName = "default"
	}

	if _, exists := f.Animations[animName]; !exists {
		f.Animations[animName] = datatypes.NewAnimation()
	}

	// Mark track as processed
	track.IsProcessed = true
	if track.TrackDefFragment != nil {
		track.TrackDefFragment.IsAssigned = true
	}

	// Find and update the bone's animation tracks
	for _, bone := range f.Skeleton {
		if strings.ToLower(bone.Name) == boneName || strings.ToLower(bone.CleanedName) == boneName {
			if bone.AnimationTracks == nil {
				bone.AnimationTracks = make(map[string]*TrackFragment)
			}
			bone.AnimationTracks[animName] = track
			break
		}
	}
}
