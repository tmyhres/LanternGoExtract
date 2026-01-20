// Package infrastructure provides common utilities and helpers for the Lantern extractor.
package infrastructure

// BitAnalyzer analyzes the individual bits of an integer value.
type BitAnalyzer struct {
	value int32
}

// NewBitAnalyzer creates a new BitAnalyzer for the given integer.
func NewBitAnalyzer(value int32) *BitAnalyzer {
	return &BitAnalyzer{
		value: value,
	}
}

// IsBitSet returns whether the bit at the specified position (0-31) is set.
// Position 0 is the least significant bit.
func (b *BitAnalyzer) IsBitSet(position int) bool {
	if position < 0 || position > 31 {
		return false
	}
	return (b.value & (1 << position)) != 0
}

// GetValue returns the underlying integer value.
func (b *BitAnalyzer) GetValue() int32 {
	return b.value
}

// CountSetBits returns the number of bits that are set (population count).
func (b *BitAnalyzer) CountSetBits() int {
	count := 0
	v := b.value
	for v != 0 {
		count += int(v & 1)
		v >>= 1
	}
	return count
}
