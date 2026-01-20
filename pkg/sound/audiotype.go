// Package sound provides types and utilities for parsing EverQuest sound data.
package sound

// AudioType describes how a sound is heard by the player.
type AudioType byte

const (
	// AudioTypeSound2D represents sounds that play at a constant volume.
	AudioTypeSound2D AudioType = 0

	// AudioTypeMusic represents music instances that can specify both day and night track IDs.
	AudioTypeMusic AudioType = 1

	// AudioTypeSound3D represents sounds with falloff - the farther the player is from
	// the center, the quieter it becomes.
	AudioTypeSound3D AudioType = 2
)

// String returns the string representation of the AudioType.
func (t AudioType) String() string {
	switch t {
	case AudioTypeSound2D:
		return "Sound2D"
	case AudioTypeMusic:
		return "Music"
	case AudioTypeSound3D:
		return "Sound3D"
	default:
		return "Unknown"
	}
}

// IsValid returns true if the AudioType is a known valid value.
func (t AudioType) IsValid() bool {
	return t <= AudioTypeSound3D
}
