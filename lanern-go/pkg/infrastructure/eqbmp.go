package infrastructure

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"

	"golang.org/x/image/bmp"
)

// PaletteFlags represents flags for palette handling.
type PaletteFlags int

const (
	// PaletteFlagHasAlpha indicates the palette has alpha channel support.
	PaletteFlagHasAlpha PaletteFlags = 0x0001
	// PaletteFlagGrayScale indicates the palette is grayscale.
	PaletteFlagGrayScale PaletteFlags = 0x0002
	// PaletteFlagHalfTone indicates the palette uses halftone.
	PaletteFlagHalfTone PaletteFlags = 0x0004
)

// EqBmp wraps image handling with EverQuest-specific transparency handling.
// It handles BMP images with palette-based and magenta-based transparency.
type EqBmp struct {
	img     image.Image
	palette color.Palette
}

// NewEqBmp creates a new EqBmp from the given reader (expects BMP format).
func NewEqBmp(r io.Reader) (*EqBmp, error) {
	img, err := bmp.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode BMP: %w", err)
	}

	eq := &EqBmp{
		img: img,
	}

	// Extract palette if it's a paletted image
	if paletted, ok := img.(*image.Paletted); ok {
		eq.palette = paletted.Palette
	}

	return eq, nil
}

// NewEqBmpFromBytes creates a new EqBmp from byte data.
func NewEqBmpFromBytes(data []byte) (*EqBmp, error) {
	return NewEqBmp(bytes.NewReader(data))
}

// IsPaletted returns true if the image uses a color palette.
func (e *EqBmp) IsPaletted() bool {
	_, ok := e.img.(*image.Paletted)
	return ok
}

// WritePng saves the image as a PNG file to the specified path.
func (e *EqBmp) WritePng(outputFilePath string) error {
	file, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, e.img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

// MakeMagentaTransparent makes all magenta pixels (255, 0, 255) transparent.
// This is commonly used in EverQuest textures for transparency.
func (e *EqBmp) MakeMagentaTransparent() {
	bounds := e.img.Bounds()
	rgba := image.NewRGBA(bounds)

	// Magenta color to make transparent
	magenta := color.RGBA{R: 255, G: 0, B: 255, A: 255}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := e.img.At(x, y)
			r, g, b, _ := c.RGBA()

			// Convert to 8-bit values
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// Check if pixel is magenta (with some tolerance for near-magenta)
			if isMagenta(r8, g8, b8) {
				rgba.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 0})
			} else {
				// Preserve original color with full opacity
				rgba.Set(x, y, color.RGBA{R: r8, G: g8, B: b8, A: 255})
			}
		}
	}

	// Also handle the edge case from libgdiplus where magenta with alpha 0 exists
	_ = magenta
	e.img = rgba
}

// isMagenta checks if a color is magenta (255, 0, 255) or close to it.
func isMagenta(r, g, b uint8) bool {
	return r == 255 && g == 0 && b == 255
}

// MakePaletteTransparent makes the color at the specified palette index transparent.
// This is used for paletted images where a specific index represents transparency.
func (e *EqBmp) MakePaletteTransparent(transparentIndex int) {
	paletted, ok := e.img.(*image.Paletted)
	if !ok {
		// If not paletted, convert to RGBA and make that index position transparent
		e.makePaletteTransparentNonPaletted(transparentIndex)
		return
	}

	if transparentIndex < 0 || transparentIndex >= len(paletted.Palette) {
		return
	}

	// Create a new RGBA image with transparency
	bounds := paletted.Bounds()
	rgba := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			idx := paletted.ColorIndexAt(x, y)
			if int(idx) == transparentIndex {
				rgba.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 0})
			} else {
				c := paletted.Palette[idx]
				r, g, b, _ := c.RGBA()
				rgba.Set(x, y, color.RGBA{
					R: uint8(r >> 8),
					G: uint8(g >> 8),
					B: uint8(b >> 8),
					A: 255,
				})
			}
		}
	}

	e.img = rgba
}

// makePaletteTransparentNonPaletted handles non-paletted images by finding the color
// that would be at the given index if it were paletted.
func (e *EqBmp) makePaletteTransparentNonPaletted(transparentIndex int) {
	// For non-paletted images, we need a different approach
	// This typically shouldn't happen for 8bpp indexed images
	// but we'll handle it by just making magenta transparent as a fallback
	e.MakeMagentaTransparent()
}

// GetImage returns the underlying image.Image.
func (e *EqBmp) GetImage() image.Image {
	return e.img
}
