// Package eq provides EverQuest file handling utilities for the Lantern extractor.
package eq

// LanternStrings contains a collection of Lantern-related string constants.
const (
	// ExportHeaderTitle is the title prefix for exported files.
	ExportHeaderTitle = "# Lantern Extractor 0.1.7 - "
	// ExportHeaderFormat is the format prefix for exported files.
	ExportHeaderFormat = "# Format: "

	// ObjMaterialHeader is the material library header for OBJ files.
	ObjMaterialHeader = "mtllib "
	// ObjUseMtlPrefix is the use material prefix for OBJ files.
	ObjUseMtlPrefix = "usemtl "
	// ObjNewMaterialPrefix is the new material prefix for MTL files.
	ObjNewMaterialPrefix = "newmtl"
	// ObjFormatExtension is the file extension for OBJ files.
	ObjFormatExtension = ".obj"
	// FormatMtlExtension is the file extension for MTL files.
	FormatMtlExtension = ".mtl"

	// WldFormatExtension is the file extension for WLD files.
	WldFormatExtension = ".wld"
	// S3dFormatExtension is the file extension for S3D archive files.
	S3dFormatExtension = ".s3d"
	// PfsFormatExtension is the file extension for PFS archive files.
	PfsFormatExtension = ".pfs"
	// PakFormatExtension is the file extension for PAK archive files.
	PakFormatExtension = ".pak"
	// T3dFormatExtension is the file extension for T3D archive files.
	T3dFormatExtension = ".t3d"
	// SoundFormatExtension is the file extension for sound effect files.
	SoundFormatExtension = ".eff"
)
