package datatypes

import "strings"

// TrackFragment is a forward declaration interface for track fragments.
// The actual implementation should be in the fragments package.
type TrackFragment interface {
	GetModelName() string
	GetFrameMs() int
	GetTrackDefFragment() TrackDefFragment
}

// TrackDefFragment is a forward declaration interface for track definition fragments.
type TrackDefFragment interface {
	GetFrameCount() int
}

// Animation holds animation data including tracks for skeletal animation.
type Animation struct {
	AnimModelBase        string
	Tracks               map[string]TrackFragment
	TracksCleaned        map[string]TrackFragment
	TracksCleanedStripped map[string]TrackFragment
	FrameCount           int
	AnimationTimeMs      int
}

// NewAnimation creates a new Animation with initialized maps.
func NewAnimation() *Animation {
	return &Animation{
		Tracks:               make(map[string]TrackFragment),
		TracksCleaned:        make(map[string]TrackFragment),
		TracksCleanedStripped: make(map[string]TrackFragment),
	}
}

// CleanBoneName cleans a bone name by removing the _dag suffix.
func CleanBoneName(boneName string) string {
	if boneName == "" {
		return boneName
	}

	cleanedName := CleanBoneNameDag(boneName)
	if len(cleanedName) == 0 {
		return "root"
	}
	return cleanedName
}

// CleanBoneAndStripBase cleans a bone name and strips the model base prefix.
func CleanBoneAndStripBase(boneName, modelBase string) string {
	cleanedName := CleanBoneNameDag(boneName)

	if strings.HasPrefix(cleanedName, modelBase) {
		cleanedName = cleanedName[len(modelBase):]
	}

	if len(cleanedName) == 0 {
		return "root"
	}
	return cleanedName
}

// CleanBoneNameDag removes the _dag suffix from a bone name.
func CleanBoneNameDag(boneName string) string {
	if boneName == "" {
		return boneName
	}

	return strings.ReplaceAll(strings.ToLower(boneName), "_dag", "")
}

// AddTrack adds a track to the animation.
func (a *Animation) AddTrack(track TrackFragment, pieceName, cleanedName, cleanStrippedName string) {
	// Prevent overwriting tracks
	// Drachnid edge case
	if _, exists := a.Tracks[pieceName]; exists {
		return
	}

	a.Tracks[pieceName] = track
	a.TracksCleaned[cleanedName] = track
	a.TracksCleanedStripped[cleanStrippedName] = track

	if a.AnimModelBase == "" && track.GetModelName() != "" {
		a.AnimModelBase = track.GetModelName()
	}

	trackDef := track.GetTrackDefFragment()
	if trackDef != nil {
		frameCount := trackDef.GetFrameCount()
		if frameCount > a.FrameCount {
			a.FrameCount = frameCount
		}

		totalTime := frameCount * track.GetFrameMs()
		if totalTime > a.AnimationTimeMs {
			a.AnimationTimeMs = totalTime
		}
	}
}
