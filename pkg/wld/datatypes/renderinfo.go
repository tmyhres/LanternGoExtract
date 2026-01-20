package datatypes

import (
	"encoding/binary"
	"io"
)

// BitmapInfoReference is a forward declaration interface for bitmap info reference fragments.
// The actual implementation should be in the fragments package.
type BitmapInfoReference interface{}

// WldFragment is a forward declaration interface for WLD fragments.
// The actual implementation should be in the fragments package.
type WldFragment interface{}

// RenderInfo contains rendering information for a surface.
type RenderInfo struct {
	Flags                 int
	Pen                   int
	Brightness            float32
	ScaledAmbient         float32
	SimpleSpriteReference BitmapInfoReference
	UvInfo                *UvInfo
	UvMap                 []Vec2
}

// BitAnalyzer provides bit manipulation utilities.
type BitAnalyzer struct {
	value int
}

// NewBitAnalyzer creates a new BitAnalyzer with the given value.
func NewBitAnalyzer(value int) *BitAnalyzer {
	return &BitAnalyzer{value: value}
}

// IsBitSet checks if the specified bit is set.
func (b *BitAnalyzer) IsBitSet(bit int) bool {
	return (b.value & (1 << bit)) != 0
}

// ParseRenderInfo parses RenderInfo from a binary reader.
func ParseRenderInfo(r io.Reader, fragments []WldFragment) (*RenderInfo, error) {
	renderInfo := &RenderInfo{}

	var flags int32
	if err := binary.Read(r, binary.LittleEndian, &flags); err != nil {
		return nil, err
	}
	renderInfo.Flags = int(flags)

	ba := NewBitAnalyzer(renderInfo.Flags)

	hasPen := ba.IsBitSet(0)
	hasBrightness := ba.IsBitSet(1)
	hasScaledAmbient := ba.IsBitSet(2)
	hasSimpleSprite := ba.IsBitSet(3)
	hasUvInfo := ba.IsBitSet(4)
	hasUvMap := ba.IsBitSet(5)
	// isTwoSided := ba.IsBitSet(6) // Currently unused

	if hasPen {
		var pen int32
		if err := binary.Read(r, binary.LittleEndian, &pen); err != nil {
			return nil, err
		}
		renderInfo.Pen = int(pen)
	}

	if hasBrightness {
		if err := binary.Read(r, binary.LittleEndian, &renderInfo.Brightness); err != nil {
			return nil, err
		}
	}

	if hasScaledAmbient {
		if err := binary.Read(r, binary.LittleEndian, &renderInfo.ScaledAmbient); err != nil {
			return nil, err
		}
	}

	if hasSimpleSprite {
		var fragmentRef int32
		if err := binary.Read(r, binary.LittleEndian, &fragmentRef); err != nil {
			return nil, err
		}
		if fragmentRef > 0 && int(fragmentRef-1) < len(fragments) {
			renderInfo.SimpleSpriteReference = fragments[fragmentRef-1]
		}
	}

	if hasUvInfo {
		uvInfo := &UvInfo{}
		var x, y, z float32

		// UvOrigin
		if err := binary.Read(r, binary.LittleEndian, &x); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &y); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &z); err != nil {
			return nil, err
		}
		uvInfo.UvOrigin = Vec3{X: x, Y: y, Z: z}

		// UAxis
		if err := binary.Read(r, binary.LittleEndian, &x); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &y); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &z); err != nil {
			return nil, err
		}
		uvInfo.UAxis = Vec3{X: x, Y: y, Z: z}

		// VAxis
		if err := binary.Read(r, binary.LittleEndian, &x); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &y); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &z); err != nil {
			return nil, err
		}
		uvInfo.VAxis = Vec3{X: x, Y: y, Z: z}

		renderInfo.UvInfo = uvInfo
	}

	if hasUvMap {
		var uvMapCount int32
		if err := binary.Read(r, binary.LittleEndian, &uvMapCount); err != nil {
			return nil, err
		}

		renderInfo.UvMap = make([]Vec2, uvMapCount)
		for i := int32(0); i < uvMapCount; i++ {
			var u, v float32
			if err := binary.Read(r, binary.LittleEndian, &u); err != nil {
				return nil, err
			}
			if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
				return nil, err
			}
			renderInfo.UvMap[i] = Vec2{X: u, Y: v}
		}
	}

	return renderInfo, nil
}
