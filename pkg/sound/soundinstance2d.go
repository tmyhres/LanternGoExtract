package sound

// SoundInstance2D represents a 2D sound that plays at constant volume.
// It can have two sounds for day and night variants.
type SoundInstance2D struct {
	SoundInstance
	Sound2    string
	Volume2   float32
	Cooldown2 int32
}

// NewSoundInstance2D creates a new SoundInstance2D with the given parameters.
func NewSoundInstance2D(
	audioType AudioType,
	posX, posY, posZ, radius float32,
	volume1 float32,
	sound1 string,
	cooldown1 int32,
	sound2 string,
	cooldown2, cooldownRandom int32,
	volume2 float32,
) *SoundInstance2D {
	return &SoundInstance2D{
		SoundInstance: NewSoundInstance(audioType, posX, posY, posZ, radius, volume1, sound1, cooldown1, cooldownRandom),
		Sound2:        sound2,
		Cooldown2:     cooldown2,
		Volume2:       volume2,
	}
}
