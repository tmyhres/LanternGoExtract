package fragments

import (
	"fmt"
	"strings"
)

// TrackFragment (0x13)
// Internal name: _TRACK
// Reference to a TrackDefFragment, a bone in a skeleton.
type TrackFragment struct {
	BaseFragment

	// TrackDefFragment is the reference to a skeleton piece.
	TrackDefFragment *TrackDefFragment

	// IsPoseAnimation indicates if this is a pose animation track.
	IsPoseAnimation bool

	// IsProcessed indicates if this track has been processed.
	IsProcessed bool

	// FrameMs is the frame time in milliseconds.
	FrameMs int32

	// ModelName is the model name parsed from the track name.
	ModelName string

	// AnimationName is the animation name parsed from the track name.
	AnimationName string

	// PieceName is the skeleton piece name parsed from the track name.
	PieceName string

	// IsNameParsed indicates if the track name has been parsed.
	IsNameParsed bool
}

// FragmentType returns the fragment type ID.
func (f *TrackFragment) FragmentType() uint32 {
	return 0x13
}

// GetModelName returns the model name.
func (f *TrackFragment) GetModelName() string {
	return f.ModelName
}

// GetFrameMs returns the frame time in milliseconds.
func (f *TrackFragment) GetFrameMs() int {
	return int(f.FrameMs)
}

// GetTrackDefFragment returns the track definition fragment interface.
func (f *TrackFragment) GetTrackDefFragment() interface{} {
	return f.TrackDefFragment
}

// Initialize parses the fragment data.
func (f *TrackFragment) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	// Read TrackDefFragment reference
	reference, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read reference: %w", err)
	}

	fragIndex := int(reference) - 1
	if fragIndex >= 0 && fragIndex < len(fragments) {
		if trackDef, ok := fragments[fragIndex].(*TrackDefFragment); ok {
			f.TrackDefFragment = trackDef
		}
	}

	if f.TrackDefFragment == nil {
		// Log error: Bad track def reference
	}

	// Either 4 or 5 - maybe something to look into
	// Bits are set 0, or 2. 0 has the extra field for delay.
	// 2 doesn't have any additional fields.
	flags, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	if IsBitSet(flags, 0) {
		f.FrameMs, err = r.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read frameMs: %w", err)
		}
	} else {
		f.FrameMs = 0
	}

	return nil
}

// SetTrackData sets the track data (model name, animation name, piece name).
func (f *TrackFragment) SetTrackData(modelName, animationName, pieceName string) {
	f.ModelName = modelName
	f.AnimationName = animationName
	f.PieceName = pieceName
}

// ParseTrackData parses the track name to extract model, animation, and piece names.
// This is only ever called when we are finding additional animations.
// All animations that are not the default skeleton animations:
// 1. Start with a 3 letter animation abbreviation (e.g. C05)
// 2. Continue with a 3 letter model name
// 3. Continue with the skeleton piece name
// 4. End with _TRACK
func (f *TrackFragment) ParseTrackData() {
	cleanedName := CleanTrackName(f.Name)

	if len(cleanedName) < 6 {
		if len(cleanedName) == 3 {
			f.ModelName = cleanedName
			f.IsNameParsed = true
			return
		}

		f.ModelName = cleanedName
		return
	}

	// Equipment edge case
	if len(cleanedName) >= 6 && cleanedName[:3] == cleanedName[3:6] {
		f.AnimationName = cleanedName[:3]
		if len(cleanedName) > 7 {
			f.ModelName = cleanedName[7:]
		}
		f.PieceName = "root"
		f.IsNameParsed = true
		return
	}

	f.AnimationName = cleanedName[:3]
	cleanedName = cleanedName[3:]
	f.ModelName = cleanedName[:3]
	cleanedName = cleanedName[3:]
	f.PieceName = cleanedName

	f.IsNameParsed = true
}

// ParseTrackDataEquipment parses the track name for equipment.
func (f *TrackFragment) ParseTrackDataEquipment(modelBase string) {
	cleanedName := CleanTrackName(f.Name)

	// Equipment edge case
	if (cleanedName == modelBase && len(cleanedName) > 6) || (len(cleanedName) >= 6 && cleanedName[:3] == cleanedName[3:6]) {
		f.AnimationName = cleanedName[:3]
		if len(cleanedName) > 7 {
			f.ModelName = cleanedName[7:]
		}
		f.PieceName = "root"
		f.IsNameParsed = true
		return
	}

	f.AnimationName = cleanedName[:3]
	cleanedName = cleanedName[3:]
	f.ModelName = modelBase
	cleanedName = strings.ReplaceAll(cleanedName, modelBase, "")
	f.PieceName = cleanedName
	f.IsNameParsed = true
}

// CleanTrackName cleans a track name by removing the _TRACK suffix.
func CleanTrackName(name string) string {
	cleanedName := strings.ToLower(name)
	cleanedName = strings.TrimSuffix(cleanedName, "_track")
	return strings.TrimSpace(cleanedName)
}
