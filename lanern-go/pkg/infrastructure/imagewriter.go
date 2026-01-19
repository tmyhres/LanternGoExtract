package infrastructure

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/infrastructure/logger"
)

// WriteImageAsPng writes image bytes (BMP or DDS format) to a PNG file.
// It automatically detects the format based on the file magic bytes.
// If isMasked is true, transparency handling will be applied.
func WriteImageAsPng(data []byte, filePath, fileName string, isMasked bool, log logger.Logger) error {
	// Check for DDS magic number: "DDS "
	// https://docs.microsoft.com/en-us/windows/win32/direct3ddds/dx-graphics-dds-pguide#dds-file-layout
	isDDS := len(data) >= 4 && string(data[0:4]) == "DDS "

	if strings.HasSuffix(strings.ToLower(fileName), ".bmp") && !isDDS {
		outputName := strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ".png"
		return writeBmpAsPng(data, filePath, outputName, isMasked, log)
	}

	outputName := strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ".png"
	return writeDdsAsPng(data, filePath, outputName)
}

// writeBmpAsPng converts a BMP image to PNG with optional transparency handling.
func writeBmpAsPng(data []byte, filePath, fileName string, isMasked bool, log logger.Logger) error {
	if filePath == "" {
		return nil
	}

	if err := os.MkdirAll(filePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	img, err := NewEqBmpFromBytes(data)
	if err != nil {
		log.LogError(fmt.Sprintf("Caught exception while creating bitmap: %v", err))
		return err
	}

	// Handle misspelled filename in the archive
	if fileName == "canwall1a.png" {
		fileName = "canwall1.png"
	}

	if img.IsPaletted() && isMasked {
		paletteIndex := getPaletteIndex(fileName)
		img.MakePaletteTransparent(paletteIndex)
	} else {
		img.MakeMagentaTransparent()
	}

	outputPath := filepath.Join(filePath, fileName)
	return img.WritePng(outputPath)
}

// writeDdsAsPng converts a DDS texture to PNG.
// Supports DXT1, DXT3, DXT5, and uncompressed RGBA32 formats.
func writeDdsAsPng(data []byte, filePath, fileName string) error {
	if len(data) < 128 {
		return fmt.Errorf("DDS file too small")
	}

	// Parse DDS header
	dds, err := parseDDSHeader(data)
	if err != nil {
		return err
	}

	// Decode the pixel data based on format
	img, err := decodeDDSPixels(dds, data[128:])
	if err != nil {
		return err
	}

	// Flip vertically (DDS textures are stored bottom-up)
	img = flipVertical(img)

	// Ensure directory exists
	if err := os.MkdirAll(filePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write PNG
	outputPath := filepath.Join(filePath, fileName)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	return png.Encode(file, img)
}

// DDSHeader represents a DDS file header.
type DDSHeader struct {
	Width       int
	Height      int
	MipMapCount int
	PixelFormat DDSPixelFormat
}

// DDSPixelFormat represents the pixel format in a DDS file.
type DDSPixelFormat struct {
	Flags       uint32
	FourCC      string
	RGBBitCount uint32
	RBitMask    uint32
	GBitMask    uint32
	BBitMask    uint32
	ABitMask    uint32
}

// DDS pixel format flags
const (
	ddpfAlphaPixels = 0x1
	ddpfFourCC      = 0x4
	ddpfRGB         = 0x40
)

// parseDDSHeader parses a DDS file header.
func parseDDSHeader(data []byte) (*DDSHeader, error) {
	if len(data) < 128 {
		return nil, fmt.Errorf("DDS data too short for header")
	}

	reader := bytes.NewReader(data)

	// Skip magic number (4 bytes "DDS ")
	reader.Seek(4, 0)

	// Read header size (should be 124)
	var headerSize uint32
	binary.Read(reader, binary.LittleEndian, &headerSize)

	// Read flags
	var flags uint32
	binary.Read(reader, binary.LittleEndian, &flags)

	// Read height and width
	var height, width uint32
	binary.Read(reader, binary.LittleEndian, &height)
	binary.Read(reader, binary.LittleEndian, &width)

	// Skip pitch/linear size and depth
	reader.Seek(4+4, 1)

	// Read mipmap count
	var mipMapCount uint32
	binary.Read(reader, binary.LittleEndian, &mipMapCount)

	// Skip reserved (11 * 4 bytes)
	reader.Seek(11*4, 1)

	// Read pixel format
	var pfSize, pfFlags uint32
	binary.Read(reader, binary.LittleEndian, &pfSize)
	binary.Read(reader, binary.LittleEndian, &pfFlags)

	fourCC := make([]byte, 4)
	reader.Read(fourCC)

	var rgbBitCount, rMask, gMask, bMask, aMask uint32
	binary.Read(reader, binary.LittleEndian, &rgbBitCount)
	binary.Read(reader, binary.LittleEndian, &rMask)
	binary.Read(reader, binary.LittleEndian, &gMask)
	binary.Read(reader, binary.LittleEndian, &bMask)
	binary.Read(reader, binary.LittleEndian, &aMask)

	return &DDSHeader{
		Width:       int(width),
		Height:      int(height),
		MipMapCount: int(mipMapCount),
		PixelFormat: DDSPixelFormat{
			Flags:       pfFlags,
			FourCC:      string(fourCC),
			RGBBitCount: rgbBitCount,
			RBitMask:    rMask,
			GBitMask:    gMask,
			BBitMask:    bMask,
			ABitMask:    aMask,
		},
	}, nil
}

// decodeDDSPixels decodes DDS pixel data based on the format.
func decodeDDSPixels(header *DDSHeader, pixelData []byte) (*image.RGBA, error) {
	pf := header.PixelFormat

	// Check for FourCC compressed formats
	if pf.Flags&ddpfFourCC != 0 {
		switch pf.FourCC {
		case "DXT1":
			return decodeDXT1(header.Width, header.Height, pixelData)
		case "DXT3":
			return decodeDXT3(header.Width, header.Height, pixelData)
		case "DXT5":
			return decodeDXT5(header.Width, header.Height, pixelData)
		default:
			return nil, fmt.Errorf("unsupported DDS FourCC format: %s", pf.FourCC)
		}
	}

	// Uncompressed RGB/RGBA
	if pf.Flags&ddpfRGB != 0 {
		return decodeUncompressedRGBA(header, pixelData)
	}

	return nil, fmt.Errorf("unsupported DDS pixel format")
}

// decodeUncompressedRGBA decodes uncompressed RGBA pixel data.
func decodeUncompressedRGBA(header *DDSHeader, pixelData []byte) (*image.RGBA, error) {
	width := header.Width
	height := header.Height
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	bytesPerPixel := int(header.PixelFormat.RGBBitCount / 8)
	if bytesPerPixel < 3 || bytesPerPixel > 4 {
		return nil, fmt.Errorf("unsupported bits per pixel: %d", header.PixelFormat.RGBBitCount)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := (y*width + x) * bytesPerPixel
			if offset+bytesPerPixel > len(pixelData) {
				break
			}

			// DDS typically stores BGRA
			b := pixelData[offset]
			g := pixelData[offset+1]
			r := pixelData[offset+2]
			a := uint8(255)
			if bytesPerPixel == 4 {
				a = pixelData[offset+3]
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}

	return img, nil
}

// decodeDXT1 decodes DXT1 (BC1) compressed texture data.
func decodeDXT1(width, height int, data []byte) (*image.RGBA, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	blockWidth := (width + 3) / 4
	blockHeight := (height + 3) / 4

	for by := 0; by < blockHeight; by++ {
		for bx := 0; bx < blockWidth; bx++ {
			blockIndex := (by*blockWidth + bx) * 8
			if blockIndex+8 > len(data) {
				continue
			}

			block := data[blockIndex : blockIndex+8]
			decodeDXT1Block(img, bx*4, by*4, width, height, block)
		}
	}

	return img, nil
}

// decodeDXT1Block decodes a single 4x4 DXT1 block.
func decodeDXT1Block(img *image.RGBA, startX, startY, width, height int, block []byte) {
	// Read two 16-bit colors
	c0 := uint16(block[0]) | uint16(block[1])<<8
	c1 := uint16(block[2]) | uint16(block[3])<<8

	// Expand to RGB
	colors := [4]color.RGBA{}
	colors[0] = rgb565ToRGBA(c0)
	colors[1] = rgb565ToRGBA(c1)

	if c0 > c1 {
		// 4-color block
		colors[2] = interpolateColors(colors[0], colors[1], 2, 1)
		colors[3] = interpolateColors(colors[0], colors[1], 1, 2)
	} else {
		// 3-color block with transparency
		colors[2] = interpolateColors(colors[0], colors[1], 1, 1)
		colors[3] = color.RGBA{R: 0, G: 0, B: 0, A: 0}
	}

	// Read 4 bytes of indices (2 bits per pixel)
	indices := binary.LittleEndian.Uint32(block[4:8])

	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			px := startX + x
			py := startY + y
			if px >= width || py >= height {
				continue
			}

			idx := (indices >> (uint(y*4+x) * 2)) & 0x3
			img.Set(px, py, colors[idx])
		}
	}
}

// decodeDXT3 decodes DXT3 (BC2) compressed texture data.
func decodeDXT3(width, height int, data []byte) (*image.RGBA, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	blockWidth := (width + 3) / 4
	blockHeight := (height + 3) / 4

	for by := 0; by < blockHeight; by++ {
		for bx := 0; bx < blockWidth; bx++ {
			blockIndex := (by*blockWidth + bx) * 16
			if blockIndex+16 > len(data) {
				continue
			}

			block := data[blockIndex : blockIndex+16]
			decodeDXT3Block(img, bx*4, by*4, width, height, block)
		}
	}

	return img, nil
}

// decodeDXT3Block decodes a single 4x4 DXT3 block.
func decodeDXT3Block(img *image.RGBA, startX, startY, width, height int, block []byte) {
	// First 8 bytes are explicit alpha (4 bits per pixel)
	alphaData := block[0:8]

	// Next 8 bytes are DXT1 color data
	colorData := block[8:16]

	c0 := uint16(colorData[0]) | uint16(colorData[1])<<8
	c1 := uint16(colorData[2]) | uint16(colorData[3])<<8

	colors := [4]color.RGBA{}
	colors[0] = rgb565ToRGBA(c0)
	colors[1] = rgb565ToRGBA(c1)
	colors[2] = interpolateColors(colors[0], colors[1], 2, 1)
	colors[3] = interpolateColors(colors[0], colors[1], 1, 2)

	indices := binary.LittleEndian.Uint32(colorData[4:8])

	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			px := startX + x
			py := startY + y
			if px >= width || py >= height {
				continue
			}

			idx := (indices >> (uint(y*4+x) * 2)) & 0x3
			c := colors[idx]

			// Get alpha from alpha data (4 bits per pixel)
			alphaIndex := y*4 + x
			alphaByte := alphaData[alphaIndex/2]
			var alpha uint8
			if alphaIndex%2 == 0 {
				alpha = (alphaByte & 0x0F) * 17 // Scale 0-15 to 0-255
			} else {
				alpha = (alphaByte >> 4) * 17
			}

			img.Set(px, py, color.RGBA{R: c.R, G: c.G, B: c.B, A: alpha})
		}
	}
}

// decodeDXT5 decodes DXT5 (BC3) compressed texture data.
func decodeDXT5(width, height int, data []byte) (*image.RGBA, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	blockWidth := (width + 3) / 4
	blockHeight := (height + 3) / 4

	for by := 0; by < blockHeight; by++ {
		for bx := 0; bx < blockWidth; bx++ {
			blockIndex := (by*blockWidth + bx) * 16
			if blockIndex+16 > len(data) {
				continue
			}

			block := data[blockIndex : blockIndex+16]
			decodeDXT5Block(img, bx*4, by*4, width, height, block)
		}
	}

	return img, nil
}

// decodeDXT5Block decodes a single 4x4 DXT5 block.
func decodeDXT5Block(img *image.RGBA, startX, startY, width, height int, block []byte) {
	// First 8 bytes are interpolated alpha
	alpha0 := block[0]
	alpha1 := block[1]
	alphaBits := uint64(block[2]) | uint64(block[3])<<8 | uint64(block[4])<<16 |
		uint64(block[5])<<24 | uint64(block[6])<<32 | uint64(block[7])<<40

	alphas := [8]uint8{}
	alphas[0] = alpha0
	alphas[1] = alpha1
	if alpha0 > alpha1 {
		alphas[2] = uint8((6*int(alpha0) + 1*int(alpha1)) / 7)
		alphas[3] = uint8((5*int(alpha0) + 2*int(alpha1)) / 7)
		alphas[4] = uint8((4*int(alpha0) + 3*int(alpha1)) / 7)
		alphas[5] = uint8((3*int(alpha0) + 4*int(alpha1)) / 7)
		alphas[6] = uint8((2*int(alpha0) + 5*int(alpha1)) / 7)
		alphas[7] = uint8((1*int(alpha0) + 6*int(alpha1)) / 7)
	} else {
		alphas[2] = uint8((4*int(alpha0) + 1*int(alpha1)) / 5)
		alphas[3] = uint8((3*int(alpha0) + 2*int(alpha1)) / 5)
		alphas[4] = uint8((2*int(alpha0) + 3*int(alpha1)) / 5)
		alphas[5] = uint8((1*int(alpha0) + 4*int(alpha1)) / 5)
		alphas[6] = 0
		alphas[7] = 255
	}

	// Next 8 bytes are DXT1 color data
	colorData := block[8:16]

	c0 := uint16(colorData[0]) | uint16(colorData[1])<<8
	c1 := uint16(colorData[2]) | uint16(colorData[3])<<8

	colors := [4]color.RGBA{}
	colors[0] = rgb565ToRGBA(c0)
	colors[1] = rgb565ToRGBA(c1)
	colors[2] = interpolateColors(colors[0], colors[1], 2, 1)
	colors[3] = interpolateColors(colors[0], colors[1], 1, 2)

	indices := binary.LittleEndian.Uint32(colorData[4:8])

	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			px := startX + x
			py := startY + y
			if px >= width || py >= height {
				continue
			}

			colorIdx := (indices >> (uint(y*4+x) * 2)) & 0x3
			c := colors[colorIdx]

			// Get alpha index (3 bits per pixel)
			alphaIdx := (alphaBits >> (uint(y*4+x) * 3)) & 0x7
			alpha := alphas[alphaIdx]

			img.Set(px, py, color.RGBA{R: c.R, G: c.G, B: c.B, A: alpha})
		}
	}
}

// rgb565ToRGBA converts a 16-bit RGB565 color to RGBA.
func rgb565ToRGBA(c uint16) color.RGBA {
	r := uint8((c >> 11) & 0x1F)
	g := uint8((c >> 5) & 0x3F)
	b := uint8(c & 0x1F)

	// Expand to 8 bits
	r = (r << 3) | (r >> 2)
	g = (g << 2) | (g >> 4)
	b = (b << 3) | (b >> 2)

	return color.RGBA{R: r, G: g, B: b, A: 255}
}

// interpolateColors interpolates between two colors.
func interpolateColors(c0, c1 color.RGBA, w0, w1 int) color.RGBA {
	total := w0 + w1
	return color.RGBA{
		R: uint8((int(c0.R)*w0 + int(c1.R)*w1) / total),
		G: uint8((int(c0.G)*w0 + int(c1.G)*w1) / total),
		B: uint8((int(c0.B)*w0 + int(c1.B)*w1) / total),
		A: 255,
	}
}

// flipVertical flips an image vertically.
func flipVertical(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	flipped := image.NewRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			flipped.Set(x, height-1-y, img.At(x, y))
		}
	}

	return flipped
}

// getPaletteIndex returns the palette index to use for transparency for specific files.
// Some EverQuest textures use non-standard palette indices for transparency.
func getPaletteIndex(fileName string) int {
	switch fileName {
	case "clhe0004.png", "kahe0001.png":
		return 255
	case "furpile1.png":
		return 250
	case "bearrug.png":
		return 47
	default:
		return 0
	}
}
