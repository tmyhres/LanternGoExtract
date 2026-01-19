package sound

// SoundInstance represents a sound with common properties shared by 2D and 3D sounds.
type SoundInstance struct {
	BaseAudioInstance
	Sound1         string
	Volume1        float32
	Cooldown1      int32
	CooldownRandom int32
}

// NewSoundInstance creates a new SoundInstance with the given parameters.
func NewSoundInstance(
	audioType AudioType,
	posX, posY, posZ, radius float32,
	volume1 float32,
	sound1 string,
	cooldown1, cooldownRandom int32,
) SoundInstance {
	return SoundInstance{
		BaseAudioInstance: NewBaseAudioInstance(audioType, posX, posY, posZ, radius),
		Sound1:            sound1,
		Volume1:           volume1,
		Cooldown1:         cooldown1,
		CooldownRandom:    cooldownRandom,
	}
}
