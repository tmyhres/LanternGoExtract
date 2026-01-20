package archive

// File represents a single file within an archive.
type File interface {
	// GetName returns the name of the file.
	GetName() string
	// SetName sets the name of the file.
	SetName(name string)
	// GetSize returns the inflated size of the file in bytes.
	GetSize() uint32
	// GetOffset returns the positional offset of the file in the archive.
	GetOffset() uint32
	// GetBytes returns the inflated bytes of the file.
	GetBytes() []byte
}

// BaseFile provides a base implementation of the File interface.
type BaseFile struct {
	Name   string
	Size   uint32
	Offset uint32
	Bytes  []byte
}

// GetName returns the name of the file.
func (f *BaseFile) GetName() string {
	return f.Name
}

// SetName sets the name of the file.
func (f *BaseFile) SetName(name string) {
	f.Name = name
}

// GetSize returns the inflated size of the file in bytes.
func (f *BaseFile) GetSize() uint32 {
	return f.Size
}

// GetOffset returns the positional offset of the file in the archive.
func (f *BaseFile) GetOffset() uint32 {
	return f.Offset
}

// GetBytes returns the inflated bytes of the file.
func (f *BaseFile) GetBytes() []byte {
	return f.Bytes
}

// NewBaseFile creates a new BaseFile with the given parameters.
func NewBaseFile(size, offset uint32, data []byte) *BaseFile {
	return &BaseFile{
		Size:   size,
		Offset: offset,
		Bytes:  data,
	}
}
