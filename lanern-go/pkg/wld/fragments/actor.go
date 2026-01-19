package fragments

import (
	"fmt"
	"strings"

	"github.com/lanterneq/lanern-go/pkg/wld/datatypes"
)

// Actor (0x14)
// Internal name: _ACTORDEF
// Information about an actor that can be spawned into the world.
// An actor will have one of five types and will reference another fragment.
type Actor struct {
	BaseFragment

	// MeshReference is the mesh reference (optional).
	MeshReference Fragment

	// SkeletonReference is the skeleton track reference (optional).
	SkeletonReference *SkeletonHierarchyReference

	// CameraReference is the camera reference (optional).
	CameraReference *CameraReference

	// ParticleSpriteReference is the particle sprite reference (optional).
	ParticleSpriteReference Fragment

	// Fragment07Ref is the Fragment07 reference (optional).
	Fragment07Ref *Fragment07

	// ActorType is the type of actor.
	ActorType datatypes.ActorType

	// ReferenceName is the name of the referenced fragment.
	ReferenceName string
}

// FragmentType returns the fragment type ID.
func (f *Actor) FragmentType() uint32 {
	return 0x14
}

// Initialize parses the fragment data.
func (f *Actor) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	flags, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	params1Exist := IsBitSet(flags, 0)
	params2Exist := IsBitSet(flags, 1)
	// fragment2MustContainZero := IsBitSet(flags, 7)

	// Is an index in the string hash
	fragment1, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read fragment1: %w", err)
	}

	// For objects, SPRITECALLBACK - and it's the same reference value
	_ = GetStringFromHash(stringHash, fragment1)

	// 1 for both static and animated objects
	size1, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read size1: %w", err)
	}

	// The number of components (meshes, skeletons, camera references) the actor has
	// In all Trilogy files, there is only ever 1
	componentCount, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read component count: %w", err)
	}

	// 0 for both static and animated objects
	_, err = r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read fragment2: %w", err)
	}

	if params1Exist {
		_, err = r.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read params1: %w", err)
		}
	}

	if params2Exist {
		err = r.Skip(7 * 4)
		if err != nil {
			return fmt.Errorf("failed to skip params2: %w", err)
		}
	}

	// Size 1 entries
	for i := int32(0); i < size1; i++ {
		// Always 1
		dataPairCount, err := r.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read data pair count: %w", err)
		}

		// Unknown purpose
		// Always 0 and 1.00000002E+30
		for j := int32(0); j < dataPairCount; j++ {
			_, err = r.ReadInt32()
			if err != nil {
				return fmt.Errorf("failed to read value: %w", err)
			}
			_, err = r.ReadInt16()
			if err != nil {
				return fmt.Errorf("failed to read value2: %w", err)
			}
			_, err = r.ReadInt16()
			if err != nil {
				return fmt.Errorf("failed to read value3: %w", err)
			}
		}
	}

	if componentCount > 1 {
		// Log warning: Actor: More than one component references
	}

	// Can contain either a skeleton reference (animated), mesh reference (static) or a camera reference
	for i := int32(0); i < componentCount; i++ {
		fragmentIndex, err := r.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read fragment index: %w", err)
		}

		fragIdx := int(fragmentIndex) - 1
		if fragIdx < 0 || fragIdx >= len(fragments) {
			continue
		}

		fragment := fragments[fragIdx]

		// Try SkeletonHierarchyReference
		if skelRef, ok := fragment.(*SkeletonHierarchyReference); ok {
			f.SkeletonReference = skelRef
			if skelRef.SkeletonHierarchy != nil {
				skelRef.SkeletonHierarchy.IsAssigned = true
			}
			break
		}

		// Try CameraReference
		if camRef, ok := fragment.(*CameraReference); ok {
			f.CameraReference = camRef
			break
		}

		// Try Fragment07
		if frag07, ok := fragment.(*Fragment07); ok {
			f.Fragment07Ref = frag07
			break
		}

		// Store as generic mesh/particle reference
		f.MeshReference = fragment
	}

	// Always 0 in qeynos2 objects
	_, err = r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name3Bytes: %w", err)
	}

	f.calculateActorType()

	return nil
}

// calculateActorType determines the actor type based on references.
func (f *Actor) calculateActorType() {
	if f.CameraReference != nil {
		f.ActorType = datatypes.ActorTypeCamera
		f.ReferenceName = f.CameraReference.Name
	} else if f.SkeletonReference != nil {
		f.ActorType = datatypes.ActorTypeSkeletal
	} else if f.MeshReference != nil {
		f.ActorType = datatypes.ActorTypeStatic
		f.ReferenceName = f.MeshReference.GetName()
	} else if f.ParticleSpriteReference != nil {
		f.ActorType = datatypes.ActorTypeParticle
		f.ReferenceName = f.ParticleSpriteReference.GetName()
	} else if f.Fragment07Ref != nil {
		f.ActorType = datatypes.ActorTypeSprite
		f.ReferenceName = f.Fragment07Ref.Name
	}
}

// AssignSkeletonReference assigns a skeleton reference to this actor.
func (f *Actor) AssignSkeletonReference(skeleton *SkeletonHierarchy) {
	f.SkeletonReference = &SkeletonHierarchyReference{
		SkeletonHierarchy: skeleton,
	}

	f.calculateActorType()
	skeleton.IsAssigned = true
}

// CleanActorName cleans an actor name by removing the _ACTORDEF suffix.
func CleanActorName(name string) string {
	cleanedName := strings.ToLower(name)
	cleanedName = strings.TrimSuffix(cleanedName, "_actordef")
	return strings.TrimSpace(cleanedName)
}
