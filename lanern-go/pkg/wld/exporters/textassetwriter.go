// Package exporters provides mesh and material export functionality for WLD files.
package exporters

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/fragments"
)

// ExportHeaderTitle is the header title for exported files.
const ExportHeaderTitle = "# Lantern Extractor 0.1.7 - "

// ExportHeaderFormat is the header format descriptor.
const ExportHeaderFormat = "# Format: "

// ObjMaterialHeader is the mtllib directive prefix for OBJ files.
const ObjMaterialHeader = "mtllib "

// ObjNewMaterialPrefix is the newmtl directive prefix for MTL files.
const ObjNewMaterialPrefix = "newmtl"

// ObjUseMtlPrefix is the usemtl directive prefix for OBJ files.
const ObjUseMtlPrefix = "usemtl "

// TextAssetWriter provides a base for text-based asset export.
type TextAssetWriter struct {
	export strings.Builder
}

// AddFragmentData adds fragment data to the export.
// This is a no-op in the base implementation; subclasses override it.
func (w *TextAssetWriter) AddFragmentData(data fragments.Fragment) {
	// Base implementation - override in subclasses
}

// WriteAssetToFile writes the export content to a file.
func (w *TextAssetWriter) WriteAssetToFile(fileName string) error {
	dir := filepath.Dir(fileName)
	if dir == "" {
		return nil
	}

	if w.export.Len() == 0 {
		return nil
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(fileName, []byte(w.export.String()), 0644)
}

// ClearExportData clears the export buffer.
func (w *TextAssetWriter) ClearExportData() {
	w.export.Reset()
}

// GetExportByteCount returns the length of the export content.
func (w *TextAssetWriter) GetExportByteCount() int {
	return w.export.Len()
}

// GetExport returns the underlying string builder.
func (w *TextAssetWriter) GetExport() *strings.Builder {
	return &w.export
}

// SetExport sets the export content from a string builder.
func (w *TextAssetWriter) SetExport(sb *strings.Builder) {
	w.export = *sb
}

// AppendString appends a string to the export.
func (w *TextAssetWriter) AppendString(s string) {
	w.export.WriteString(s)
}

// AppendLine appends a string followed by a newline to the export.
func (w *TextAssetWriter) AppendLine(s string) {
	w.export.WriteString(s)
	w.export.WriteString("\n")
}

// TextAssetWriterInterface defines the interface for text-based asset writers.
type TextAssetWriterInterface interface {
	// AddFragmentData adds fragment data to the export buffer.
	AddFragmentData(data fragments.Fragment)
	// WriteAssetToFile writes the accumulated data to a file.
	WriteAssetToFile(fileName string) error
	// ClearExportData clears the export buffer.
	ClearExportData()
	// GetExportByteCount returns the current byte count of the export buffer.
	GetExportByteCount() int
}

// BaseTextAssetWriter is an alias for TextAssetWriter for compatibility with new exporter implementations.
type BaseTextAssetWriter = TextAssetWriter
