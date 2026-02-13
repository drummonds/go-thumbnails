package thumbnails

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// Style controls the thumbnail rendering mode.
type Style int

const (
	// StyleComposite renders multi-page documents as side-by-side page tiles
	// with a "+" indicator for documents with more than 4 pages.
	StyleComposite Style = iota
	// StyleUniform renders all documents as a fixed width × 1.42×width thumbnail
	// with a page-count watermark for multi-page documents.
	StyleUniform
)

// pageHeight returns the height for a composite-style page thumbnail,
// using the A4 / ISO 216 aspect ratio (1 : √2).
func pageHeight(width uint) uint {
	return uint(math.Round(float64(width) * math.Sqrt2))
}

// renderPages extracts page images from a document file.
func renderPages(filePath string) ([]image.Image, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".pdf":
		return renderPDFPages(filePath)
	case ".tif", ".tiff":
		return renderTIFFPages(filePath)
	case ".jpg", ".jpeg", ".png":
		img, err := renderImagePage(filePath)
		if err != nil {
			return nil, err
		}
		return []image.Image{img}, nil
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

// Generate reads a file from disk and returns a composite-style thumbnail.
// Width is the desired thumbnail width in pixels; height is width × √2 (A4 ratio).
// Supported formats: PDF, TIFF (multi-page composite), JPG, PNG (simple resize).
func Generate(filePath string, width uint) (image.Image, error) {
	return GenerateStyled(filePath, width, StyleComposite)
}

// GenerateStyled reads a file and returns a thumbnail in the given style.
func GenerateStyled(filePath string, width uint, style Style) (image.Image, error) {
	pages, err := renderPages(filePath)
	if err != nil {
		return nil, err
	}

	switch style {
	case StyleUniform:
		return uniformPage(pages[0], len(pages), width), nil
	default:
		return compositePages(pages, width), nil
	}
}

// GenerateAndSave generates a composite-style thumbnail and saves it as PNG to outputPath.
func GenerateAndSave(filePath, outputPath string, width uint) error {
	return GenerateStyledAndSave(filePath, outputPath, width, StyleComposite)
}

// GenerateStyledAndSave generates a styled thumbnail and saves it as PNG to outputPath.
func GenerateStyledAndSave(filePath, outputPath string, width uint, style Style) error {
	img, err := GenerateStyled(filePath, width, style)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return nil
}

// DefaultThumbnailPath returns the conventional thumbnail path for a document.
// e.g. "doc.pdf" -> "doc.tn_64.png"
func DefaultThumbnailPath(docPath string, width uint) string {
	ext := filepath.Ext(docPath)
	base := docPath[:len(docPath)-len(ext)]
	return fmt.Sprintf("%s.tn_%d.png", base, width)
}
