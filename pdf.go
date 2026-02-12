package thumbnails

import (
	"fmt"
	"image"

	"github.com/drummonds/go-thumbnails/pdfrenderer"
)

// generatePDFThumbnail creates a thumbnail from a PDF file.
// It renders all pages and composites up to 4 side-by-side.
func generatePDFThumbnail(path string, width uint) (image.Image, error) {
	renderer, err := pdfrenderer.NewPDFiumRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF renderer: %w", err)
	}
	defer renderer.Close()

	pages, err := renderer.RenderPDF(path)
	if err != nil {
		return nil, fmt.Errorf("failed to render PDF pages: %w", err)
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("PDF has no pages")
	}

	return compositePages(pages, width), nil
}
