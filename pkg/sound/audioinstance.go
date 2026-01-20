package sound

// AudioInstance is the interface for all audio instance types.
// It provides common properties for position and audibility.
type AudioInstance interface {
	// Type returns the AudioType of this instance.
	Type() AudioType

	// Position returns the X, Y, Z coordinates of the audio source.
	Position() (x, y, z float32)

	// Radius returns the audible radius of the sound.
	Radius() float32
}

// BaseAudioInstance contains the common fields for all audio instances.
type BaseAudioInstance struct {
	AudioType AudioType
	PosX      float32
	PosY      float32
	PosZ      float32
	PosRadius float32
}

// Type returns the AudioType of this instance.
func (b *BaseAudioInstance) Type() AudioType {
	return b.AudioType
}

// Position returns the X, Y, Z coordinates of the audio source.
func (b *BaseAudioInstance) Position() (x, y, z float32) {
	return b.PosX, b.PosY, b.PosZ
}

// Radius returns the audible radius of the sound.
func (b *BaseAudioInstance) Radius() float32 {
	return b.PosRadius
}

// NewBaseAudioInstance creates a new BaseAudioInstance with the given parameters.
func NewBaseAudioInstance(audioType AudioType, posX, posY, posZ, radius float32) BaseAudioInstance {
	return BaseAudioInstance{
		AudioType: audioType,
		PosX:      posX,
		PosY:      posY,
		PosZ:      posZ,
		PosRadius: radius,
	}
}
