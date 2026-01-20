package fragments

import (
	"fmt"

	"github.com/tmyhres/LanternGoExtract/pkg/wld/datatypes"
)

// TrackDefFragment (0x12)
// Internal name: _TRACKDEF
// Describes how a bone of a skeleton is rotated and shifted in relation to the parent.
type TrackDefFragment struct {
	BaseFragment

	// Frames is a list of bone positions for each frame.
	Frames []datatypes.BoneTransform

	// IsAssigned indicates whether this track has been assigned to a skeleton.
	IsAssigned bool
}

// FragmentType returns the fragment type ID.
func (f *TrackDefFragment) FragmentType() uint32 {
	return 0x12
}

// GetFrameCount returns the number of frames.
func (f *TrackDefFragment) GetFrameCount() int {
	return len(f.Frames)
}

// Initialize parses the fragment data.
func (f *TrackDefFragment) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	// Read flags
	flags, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	// Flags are always 8 when dealing with object animations
	isS3dTrack2 := IsBitSet(flags, 3)

	// Read frame count
	frameCount, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read frame count: %w", err)
	}

	f.Frames = make([]datatypes.BoneTransform, 0, frameCount)

	if isS3dTrack2 {
		for i := int32(0); i < frameCount; i++ {
			rotDenominator, err := r.ReadInt16()
			if err != nil {
				return fmt.Errorf("failed to read rotDenominator: %w", err)
			}
			rotX, err := r.ReadInt16()
			if err != nil {
				return fmt.Errorf("failed to read rotX: %w", err)
			}
			rotY, err := r.ReadInt16()
			if err != nil {
				return fmt.Errorf("failed to read rotY: %w", err)
			}
			rotZ, err := r.ReadInt16()
			if err != nil {
				return fmt.Errorf("failed to read rotZ: %w", err)
			}
			shiftX, err := r.ReadInt16()
			if err != nil {
				return fmt.Errorf("failed to read shiftX: %w", err)
			}
			shiftY, err := r.ReadInt16()
			if err != nil {
				return fmt.Errorf("failed to read shiftY: %w", err)
			}
			shiftZ, err := r.ReadInt16()
			if err != nil {
				return fmt.Errorf("failed to read shiftZ: %w", err)
			}
			shiftDenominator, err := r.ReadInt16()
			if err != nil {
				return fmt.Errorf("failed to read shiftDenominator: %w", err)
			}

			frameTransform := datatypes.BoneTransform{}

			if shiftDenominator != 0 {
				x := float32(shiftX) / 256.0
				y := float32(shiftY) / 256.0
				z := float32(shiftZ) / 256.0

				frameTransform.Scale = float32(shiftDenominator) / 256.0
				frameTransform.Translation = datatypes.Vec3{X: x, Y: y, Z: z}
			} else {
				frameTransform.Translation = datatypes.Vec3{X: 0, Y: 0, Z: 0}
			}

			// Normalize the quaternion
			frameTransform.Rotation = normalizeQuat(datatypes.Quat{
				X: float32(rotX),
				Y: float32(rotY),
				Z: float32(rotZ),
				W: float32(rotDenominator),
			})
			f.Frames = append(f.Frames, frameTransform)
		}
	} else {
		for i := int32(0); i < frameCount; i++ {
			shiftDenominator, err := r.ReadFloat32()
			if err != nil {
				return fmt.Errorf("failed to read shiftDenominator: %w", err)
			}
			shiftX, err := r.ReadFloat32()
			if err != nil {
				return fmt.Errorf("failed to read shiftX: %w", err)
			}
			shiftY, err := r.ReadFloat32()
			if err != nil {
				return fmt.Errorf("failed to read shiftY: %w", err)
			}
			shiftZ, err := r.ReadFloat32()
			if err != nil {
				return fmt.Errorf("failed to read shiftZ: %w", err)
			}
			rotW, err := r.ReadFloat32()
			if err != nil {
				return fmt.Errorf("failed to read rotW: %w", err)
			}
			rotX, err := r.ReadFloat32()
			if err != nil {
				return fmt.Errorf("failed to read rotX: %w", err)
			}
			rotY, err := r.ReadFloat32()
			if err != nil {
				return fmt.Errorf("failed to read rotY: %w", err)
			}
			rotZ, err := r.ReadFloat32()
			if err != nil {
				return fmt.Errorf("failed to read rotZ: %w", err)
			}

			frameTransform := datatypes.BoneTransform{
				Scale:       shiftDenominator,
				Translation: datatypes.Vec3{X: shiftX, Y: shiftY, Z: shiftZ},
				Rotation:    normalizeQuat(datatypes.Quat{X: rotX, Y: rotY, Z: rotZ, W: rotW}),
			}

			f.Frames = append(f.Frames, frameTransform)
		}
	}

	return nil
}

// normalizeQuat normalizes a quaternion.
func normalizeQuat(q datatypes.Quat) datatypes.Quat {
	length := q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W
	if length > 0 {
		invLen := 1.0 / sqrt32(length)
		return datatypes.Quat{
			X: q.X * invLen,
			Y: q.Y * invLen,
			Z: q.Z * invLen,
			W: q.W * invLen,
		}
	}
	return q
}

// sqrt32 computes the square root of a float32.
func sqrt32(x float32) float32 {
	// Using Newton-Raphson method for float32
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}
