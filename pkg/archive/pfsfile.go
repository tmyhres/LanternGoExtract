package archive

// PfsFile represents a single file within a PFS archive.
type PfsFile struct {
	*BaseFile
	Crc uint32
}

// NewPfsFile creates a new PfsFile with the given parameters.
func NewPfsFile(crc, size, offset uint32, data []byte) *PfsFile {
	return &PfsFile{
		BaseFile: NewBaseFile(size, offset, data),
		Crc:      crc,
	}
}

// GetCrc returns the CRC of this file.
func (f *PfsFile) GetCrc() uint32 {
	return f.Crc
}
