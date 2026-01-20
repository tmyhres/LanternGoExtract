package wld

// hashKey is the XOR key used to decode WLD strings.
// The key is 8 bytes and is applied cyclically to the encoded string.
var hashKey = []byte{0x95, 0x3A, 0xC5, 0x2A, 0x95, 0x7A, 0x95, 0x6A}

// DecodeString decodes an XOR-encoded WLD string using the hash key.
// The encoding is a simple XOR cipher where each byte of the input
// is XORed with the corresponding byte in the 8-byte hash key (cycled).
func DecodeString(encodedString []byte) string {
	if encodedString == nil {
		return ""
	}

	// Create a copy to avoid modifying the original slice
	decoded := make([]byte, len(encodedString))
	for i := range encodedString {
		decoded[i] = encodedString[i] ^ hashKey[i%8]
	}

	return string(decoded)
}

// DecodeStringInPlace decodes an XOR-encoded WLD string in place.
// This modifies the original byte slice.
func DecodeStringInPlace(encodedString []byte) string {
	if encodedString == nil {
		return ""
	}

	for i := range encodedString {
		encodedString[i] ^= hashKey[i%8]
	}

	return string(encodedString)
}

// EncodeString encodes a string using the same XOR cipher.
// Since XOR is its own inverse, encoding and decoding use the same operation.
func EncodeString(plainString string) []byte {
	if plainString == "" {
		return nil
	}

	plainBytes := []byte(plainString)
	encoded := make([]byte, len(plainBytes))
	for i := range plainBytes {
		encoded[i] = plainBytes[i] ^ hashKey[i%8]
	}

	return encoded
}
