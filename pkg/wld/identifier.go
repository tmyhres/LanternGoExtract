package wld

// WLD file format magic numbers and version identifiers.
const (
	// WldFileIdentifier is the magic number at the start of all WLD files.
	// This is the hex value 0x54503D02.
	WldFileIdentifier int32 = 0x54503D02

	// WldFormatOldIdentifier indicates the old WLD format version.
	// This is the hex value 0x00015500.
	WldFormatOldIdentifier int32 = 0x00015500

	// WldFormatNewIdentifier indicates the new WLD format version.
	// This is the hex value 0x1000C800.
	WldFormatNewIdentifier int32 = 0x1000C800
)

// IsValidWldIdentifier checks if the given identifier is a valid WLD file identifier.
func IsValidWldIdentifier(identifier int32) bool {
	return identifier == WldFileIdentifier
}

// IsOldFormat checks if the given version indicates the old WLD format.
func IsOldFormat(version int32) bool {
	return version == WldFormatOldIdentifier
}

// IsNewFormat checks if the given version indicates the new WLD format.
func IsNewFormat(version int32) bool {
	return version == WldFormatNewIdentifier
}
