package sound

// SoundInstance3D represents a 3D sound with distance-based falloff.
type SoundInstance3D struct {
	SoundInstance
	Multiplier int32
}

// NewSoundInstance3D creates a new SoundInstance3D with the given parameters.
func NewSoundInstance3D(
	audioType AudioType,
	posX, posY, posZ, radius float32,
	volume float32,
	sound1 string,
	cooldown1, cooldownRandom int32,
	multiplier int32,
) *SoundInstance3D {
	return &SoundInstance3D{
		SoundInstance: NewSoundInstance(audioType, posX, posY, posZ, radius, volume, sound1, cooldown1, cooldownRandom),
		Multiplier:    multiplier,
	}
}
