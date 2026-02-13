package thumbnails

import (
	"fmt"
	"image"

	"github.com/drummonds/go-thumbnails/pdfrenderer"
)

// renderPDFPages renders all pages of a PDF file as images.
func renderPDFPages(path string) ([]image.Image, error) {
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

	return pages, nil
}
