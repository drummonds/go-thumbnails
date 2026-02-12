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

// pageHeight returns the height for a page thumbnail of the given width,
// using the A4 / ISO 216 aspect ratio (1 : √2).
func pageHeight(width uint) uint {
	return uint(math.Round(float64(width) * math.Sqrt2))
}

// Generate reads a file from disk and returns a thumbnail image.
// Width is the desired thumbnail width in pixels; height is width × √2 (A4 ratio).
// Supported formats: PDF, TIFF (multi-page composite), JPG, PNG (simple resize).
func Generate(filePath string, width uint) (image.Image, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".pdf":
		return generatePDFThumbnail(filePath, width)
	case ".tif", ".tiff":
		return generateTIFFThumbnail(filePath, width)
	case ".jpg", ".jpeg", ".png":
		return generateImageThumbnail(filePath, width)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

// GenerateAndSave generates a thumbnail and saves it as PNG to outputPath.
func GenerateAndSave(filePath, outputPath string, width uint) error {
	img, err := Generate(filePath, width)
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
