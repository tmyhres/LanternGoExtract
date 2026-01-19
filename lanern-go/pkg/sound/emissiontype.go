package sound

// EmissionType describes the source category of a sound.
type EmissionType int

const (
	// EmissionTypeNone indicates no sound emission.
	EmissionTypeNone EmissionType = 0

	// EmissionTypeEmit represents emitted sounds like bird noises.
	EmissionTypeEmit EmissionType = 1

	// EmissionTypeLoop represents looped sounds like oceans or lakes.
	EmissionTypeLoop EmissionType = 2

	// EmissionTypeInternal represents sounds internal to the client.
	EmissionTypeInternal EmissionType = 3
)

// String returns the string representation of the EmissionType.
func (e EmissionType) String() string {
	switch e {
	case EmissionTypeNone:
		return "None"
	case EmissionTypeEmit:
		return "Emit"
	case EmissionTypeLoop:
		return "Loop"
	case EmissionTypeInternal:
		return "Internal"
	default:
		return "Unknown"
	}
}
