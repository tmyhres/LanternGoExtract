// Package fragments contains WLD fragment definitions and parsers.
package fragments

import (
	"bytes"
	"encoding/binary"
	"io"
)

// Fragment is the interface that all WLD fragments must implement.
type Fragment interface {
	// Initialize parses the fragment data and initializes the fragment.
	// Parameters:
	//   - index: The fragment's index in the WLD file
	//   - id: The fragment type ID
	//   - size: The size of the fragment data in bytes
	//   - data: The raw fragment data bytes
	//   - fragments: List of all previously parsed fragments for reference resolution
	//   - stringHash: Map of string hash indices to decoded strings
	//   - isNewFormat: True if this is the new WLD format
	Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error

	// FragmentType returns the fragment type ID.
	FragmentType() uint32

	// SetIndex sets the fragment index.
	SetIndex(index int)

	// GetIndex returns the fragment index.
	GetIndex() int

	// SetName sets the fragment name.
	SetName(name string)

	// GetName returns the fragment name.
	GetName() string

	// GetSize returns the fragment size in bytes.
	GetSize() int
}

// BaseFragment provides common functionality for all fragments.
type BaseFragment struct {
	Index int
	Size  int
	Name  string
}

// SetIndex sets the fragment index.
func (f *BaseFragment) SetIndex(index int) {
	f.Index = index
}

// GetIndex returns the fragment index.
func (f *BaseFragment) GetIndex() int {
	return f.Index
}

// SetName sets the fragment name.
func (f *BaseFragment) SetName(name string) {
	f.Name = name
}

// GetName returns the fragment name.
func (f *BaseFragment) GetName() string {
	return f.Name
}

// GetSize returns the fragment size in bytes.
func (f *BaseFragment) GetSize() int {
	return f.Size
}

// FragmentType returns 0 for the base fragment. Override in derived types.
func (f *BaseFragment) FragmentType() uint32 {
	return 0
}

// initBase initializes the base fragment fields.
func (f *BaseFragment) initBase(index, size int) {
	f.Index = index
	f.Size = size
}

// FragmentReader wraps a bytes.Reader with helper methods for reading WLD data.
type FragmentReader struct {
	*bytes.Reader
}

// NewFragmentReader creates a new FragmentReader from the given data.
func NewFragmentReader(data []byte) *FragmentReader {
	return &FragmentReader{bytes.NewReader(data)}
}

// ReadInt32 reads a little-endian int32.
func (r *FragmentReader) ReadInt32() (int32, error) {
	var v int32
	err := binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

// ReadUint32 reads a little-endian uint32.
func (r *FragmentReader) ReadUint32() (uint32, error) {
	var v uint32
	err := binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

// ReadInt16 reads a little-endian int16.
func (r *FragmentReader) ReadInt16() (int16, error) {
	var v int16
	err := binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

// ReadUint16 reads a little-endian uint16.
func (r *FragmentReader) ReadUint16() (uint16, error) {
	var v uint16
	err := binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

// ReadFloat32 reads a little-endian float32.
func (r *FragmentReader) ReadFloat32() (float32, error) {
	var v float32
	err := binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

// ReadByte reads a single byte.
func (r *FragmentReader) ReadByte() (byte, error) {
	return r.Reader.ReadByte()
}

// ReadBytes reads n bytes.
func (r *FragmentReader) ReadBytes(n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := io.ReadFull(r, buf)
	return buf, err
}

// Skip skips n bytes.
func (r *FragmentReader) Skip(n int64) error {
	_, err := r.Seek(n, io.SeekCurrent)
	return err
}

// IsBitSet checks if a specific bit is set in a flags integer.
func IsBitSet(flags int32, position int) bool {
	return (flags & (1 << position)) != 0
}

// GetStringFromHash retrieves a string from the string hash by negated index.
// Returns empty string if the key is not found.
func GetStringFromHash(stringHash map[int]string, index int32) string {
	if s, ok := stringHash[-int(index)]; ok {
		return s
	}
	return ""
}

// IAnimatedVertices is an interface for fragments that contain animated vertex data.
type IAnimatedVertices interface {
	GetDelay() int
	GetFrames() [][]Vec3
}

// Vec3 represents a 3D vector (local definition for use in interface).
type Vec3 struct {
	X float32
	Y float32
	Z float32
}

// HashKey is the key used to decode WLD strings.
var HashKey = []byte{0x95, 0x3A, 0xC5, 0x2A, 0x95, 0x7A, 0x95, 0x6A}

// DecodeString decodes an encoded WLD string.
func DecodeString(encodedString []byte) string {
	decoded := make([]byte, len(encodedString))
	for i := 0; i < len(encodedString); i++ {
		decoded[i] = encodedString[i] ^ HashKey[i%8]
	}
	// Trim null terminator if present
	for len(decoded) > 0 && decoded[len(decoded)-1] == 0 {
		decoded = decoded[:len(decoded)-1]
	}
	return string(decoded)
}
