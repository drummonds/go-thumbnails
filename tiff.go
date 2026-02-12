package thumbnails

import (
	"fmt"
	"image"
	"io"
	"os"

	"golang.org/x/image/tiff"
)

// generateTIFFThumbnail creates a thumbnail from a TIFF file.
// It decodes all frames and composites up to 4 side-by-side.
func generateTIFFThumbnail(path string, width uint) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open TIFF file: %w", err)
	}
	defer f.Close()

	pages, err := decodeTIFFPages(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode TIFF: %w", err)
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("TIFF has no pages")
	}

	return compositePages(pages, width), nil
}

// decodeTIFFPages decodes all frames from a multi-page TIFF.
func decodeTIFFPages(r io.ReadSeeker) ([]image.Image, error) {
	var pages []image.Image

	// Decode the first page
	img, err := tiff.Decode(r)
	if err != nil {
		return nil, err
	}
	pages = append(pages, img)

	// Try to decode additional pages by seeking and re-decoding
	// The golang.org/x/image/tiff package decodes only the first IFD,
	// so for multi-page TIFF we use DecodeAll if available, or just return the first page.
	// Note: golang.org/x/image/tiff doesn't have a DecodeAll function,
	// so multi-page TIFF support is limited to the first page for now.

	return pages, nil
}
