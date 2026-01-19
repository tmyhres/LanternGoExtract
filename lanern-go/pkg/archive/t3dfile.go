package archive

// T3dFile represents a single file within a T3D archive.
type T3dFile struct {
	*BaseFile
}

// NewT3dFile creates a new T3dFile with the given parameters.
func NewT3dFile(size, offset uint32, data []byte) *T3dFile {
	return &T3dFile{
		BaseFile: NewBaseFile(size, offset, data),
	}
}
