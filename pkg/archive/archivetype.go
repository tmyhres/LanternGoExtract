// Package archive provides functionality for reading EverQuest archive files.
package archive

// Type represents the type of archive format.
type Type int

const (
	// TypeUnknown indicates an unrecognized archive format.
	TypeUnknown Type = iota
	// TypePfs indicates a PFS/S3D/PAK archive format.
	TypePfs
	// TypeT3d indicates a T3D archive format.
	TypeT3d
)

// String returns the string representation of the archive type.
func (t Type) String() string {
	switch t {
	case TypePfs:
		return "PFS"
	case TypeT3d:
		return "T3D"
	default:
		return "Unknown"
	}
}
