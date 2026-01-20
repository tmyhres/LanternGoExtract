package fragments

import (
	"fmt"
	"strings"

	"github.com/tmyhres/LanternGoExtract/pkg/wld/datatypes"
)

// ObjectInstance (0x15)
// Internal name: None
// Information about a single instance of an object.
type ObjectInstance struct {
	BaseFragment

	// ObjectName is the name of the object model.
	ObjectName string

	// Position is the instance position in the world.
	Position datatypes.Vec3

	// Rotation is the instance rotation in the world.
	Rotation datatypes.Vec3

	// Scale is the instance scale in the world.
	Scale datatypes.Vec3

	// Colors is the vertex colors lighting data for this instance.
	Colors Fragment
}

// FragmentType returns the fragment type ID.
func (f *ObjectInstance) FragmentType() uint32 {
	return 0x15
}

// Initialize parses the fragment data.
func (f *ObjectInstance) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	// In main zone, points to 0x16, in object wld, it contains the object name
	reference, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read reference: %w", err)
	}

	if reference < 0 {
		objectName := GetStringFromHash(stringHash, reference)
		objectName = strings.ReplaceAll(objectName, "_ACTORDEF", "")
		f.ObjectName = strings.ToLower(objectName)
	} else {
		f.ObjectName = ""
	}

	// Main zone: 0x2E, Objects: 0x32E
	_, err = r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	// Fragment reference
	// In main zone, it points to a 0x16 fragment
	// In objects.wld, it is 0
	_, err = r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read unknown2: %w", err)
	}

	// Read position
	posX, err := r.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read position X: %w", err)
	}
	posY, err := r.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read position Y: %w", err)
	}
	posZ, err := r.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read position Z: %w", err)
	}
	f.Position = datatypes.Vec3{X: posX, Y: posY, Z: posZ}

	// Rotation is strange. There is never any x rotation (roll)
	// The z rotation is negated
	value0, err := r.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read rotation value0: %w", err)
	}
	value1, err := r.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read rotation value1: %w", err)
	}
	_, err = r.ReadFloat32() // value2 unused
	if err != nil {
		return fmt.Errorf("failed to read rotation value2: %w", err)
	}

	modifier := float32(1.0 / 512.0 * 360.0)
	f.Rotation = datatypes.Vec3{X: 0, Y: value1 * modifier, Z: -(value0 * modifier)}

	// Only scale y is used
	_, err = r.ReadFloat32() // scaleX unused
	if err != nil {
		return fmt.Errorf("failed to read scaleX: %w", err)
	}
	scaleY, err := r.ReadFloat32()
	if err != nil {
		return fmt.Errorf("failed to read scaleY: %w", err)
	}
	_, err = r.ReadFloat32() // scaleZ unused
	if err != nil {
		return fmt.Errorf("failed to read scaleZ: %w", err)
	}

	f.Scale = datatypes.Vec3{X: scaleY, Y: scaleY, Z: scaleY}

	colorFragment, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read color fragment: %w", err)
	}

	if colorFragment != 0 {
		fragIdx := int(colorFragment) - 1
		if fragIdx >= 0 && fragIdx < len(fragments) {
			f.Colors = fragments[fragIdx]
		}
	}

	return nil
}
