package sound

// MusicInstance represents a music track that can play different tracks
// during day and night.
type MusicInstance struct {
	BaseAudioInstance
	TrackIndexDay   int32
	TrackIndexNight int32
	LoopCountDay    int32
	LoopCountNight  int32
	FadeOutMs       int32
}

// NewMusicInstance creates a new MusicInstance with the given parameters.
func NewMusicInstance(
	audioType AudioType,
	posX, posY, posZ, radius float32,
	trackIndexDay, trackIndexNight int32,
	loopCountDay, loopCountNight int32,
	fadeOutMs int32,
) *MusicInstance {
	return &MusicInstance{
		BaseAudioInstance: NewBaseAudioInstance(audioType, posX, posY, posZ, radius),
		TrackIndexDay:     trackIndexDay,
		TrackIndexNight:   trackIndexNight,
		LoopCountDay:      loopCountDay,
		LoopCountNight:    loopCountNight,
		FadeOutMs:         fadeOutMs,
	}
}
