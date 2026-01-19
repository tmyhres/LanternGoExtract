package infrastructure

import (
	"encoding/binary"
	"io"
)

// ReadNullTerminatedString reads bytes from the reader until a null byte (0x00) is encountered.
// The null byte is consumed but not included in the returned string.
func ReadNullTerminatedString(r io.Reader) (string, error) {
	var result []byte
	buf := make([]byte, 1)

	for {
		_, err := r.Read(buf)
		if err != nil {
			if err == io.EOF {
				// Return what we have if we hit EOF
				return string(result), nil
			}
			return "", err
		}

		if buf[0] == 0 {
			break
		}
		result = append(result, buf[0])
	}

	return string(result), nil
}

// ReadNullTerminatedStringFromBytes reads a null-terminated string from a byte slice starting at the given offset.
// Returns the string and the number of bytes consumed (including the null terminator).
func ReadNullTerminatedStringFromBytes(data []byte, offset int) (string, int) {
	if offset >= len(data) {
		return "", 0
	}

	end := offset
	for end < len(data) && data[end] != 0 {
		end++
	}

	consumed := end - offset
	if end < len(data) {
		consumed++ // Include the null terminator in consumed count
	}

	return string(data[offset:end]), consumed
}

// ReadInt32 reads a little-endian int32 from the reader.
func ReadInt32(r io.Reader) (int32, error) {
	var value int32
	err := binary.Read(r, binary.LittleEndian, &value)
	return value, err
}

// ReadUint32 reads a little-endian uint32 from the reader.
func ReadUint32(r io.Reader) (uint32, error) {
	var value uint32
	err := binary.Read(r, binary.LittleEndian, &value)
	return value, err
}

// ReadInt16 reads a little-endian int16 from the reader.
func ReadInt16(r io.Reader) (int16, error) {
	var value int16
	err := binary.Read(r, binary.LittleEndian, &value)
	return value, err
}

// ReadUint16 reads a little-endian uint16 from the reader.
func ReadUint16(r io.Reader) (uint16, error) {
	var value uint16
	err := binary.Read(r, binary.LittleEndian, &value)
	return value, err
}

// ReadFloat32 reads a little-endian float32 from the reader.
func ReadFloat32(r io.Reader) (float32, error) {
	var value float32
	err := binary.Read(r, binary.LittleEndian, &value)
	return value, err
}

// ReadBytes reads exactly n bytes from the reader.
func ReadBytes(r io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
