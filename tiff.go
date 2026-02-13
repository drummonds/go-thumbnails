package thumbnails

import (
	"fmt"
	"image"
	"io"
	"os"

	"golang.org/x/image/tiff"
)

// renderTIFFPages decodes all pages from a TIFF file.
func renderTIFFPages(path string) ([]image.Image, error) {
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

	return pages, nil
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

	// golang.org/x/image/tiff doesn't have a DecodeAll function,
	// so multi-page TIFF support is limited to the first page for now.

	return pages, nil
}
