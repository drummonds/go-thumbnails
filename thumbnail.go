package thumbnails

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// Generate reads a file from disk and returns a thumbnail image.
// Height is the desired thumbnail height in pixels; width scales proportionally.
// Supported formats: PDF, TIFF (multi-page composite), JPG, PNG (simple resize).
func Generate(filePath string, height uint) (image.Image, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".pdf":
		return generatePDFThumbnail(filePath, height)
	case ".tif", ".tiff":
		return generateTIFFThumbnail(filePath, height)
	case ".jpg", ".jpeg", ".png":
		return generateImageThumbnail(filePath, height)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

// GenerateAndSave generates a thumbnail and saves it as PNG to outputPath.
func GenerateAndSave(filePath, outputPath string, height uint) error {
	img, err := Generate(filePath, height)
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
func DefaultThumbnailPath(docPath string, height uint) string {
	ext := filepath.Ext(docPath)
	base := docPath[:len(docPath)-len(ext)]
	return fmt.Sprintf("%s.tn_%d.png", base, height)
}
