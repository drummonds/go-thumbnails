package thumbnails

import (
	"errors"
	"fmt"
	"image"
	"path/filepath"
)

// ErrPageOutOfRange is returned when a requested page number exceeds the document's page count.
var ErrPageOutOfRange = errors.New("page number out of range")

// PageResult holds a rendered page image along with its metadata.
type PageResult struct {
	Image     image.Image
	PageNum   int // 1-based page number
	PageCount int // total pages in the document
}

// RenderPages renders all pages of a document at full resolution.
func RenderPages(filePath string) ([]PageResult, error) {
	pages, err := renderPages(filePath)
	if err != nil {
		return nil, err
	}

	results := make([]PageResult, len(pages))
	for i, img := range pages {
		results[i] = PageResult{
			Image:     img,
			PageNum:   i + 1,
			PageCount: len(pages),
		}
	}
	return results, nil
}

// RenderPage renders a single page (1-based) of a document at full resolution.
func RenderPage(filePath string, pageNum int) (PageResult, error) {
	pages, err := RenderPages(filePath)
	if err != nil {
		return PageResult{}, err
	}

	if pageNum < 1 || pageNum > len(pages) {
		return PageResult{}, fmt.Errorf("%w: requested page %d, document has %d pages", ErrPageOutOfRange, pageNum, len(pages))
	}

	return pages[pageNum-1], nil
}

// ResizePage scales a page image to the given width, preserving A4 aspect ratio.
func ResizePage(img image.Image, width uint) image.Image {
	return resizeToPage(img, width)
}

// DefaultPageThumbnailPath returns the conventional per-page thumbnail path.
// e.g. "doc.pdf", 3, 128 -> "doc.p3.tn_128.png"
func DefaultPageThumbnailPath(docPath string, pageNum int, width uint) string {
	ext := filepath.Ext(docPath)
	base := docPath[:len(docPath)-len(ext)]
	return fmt.Sprintf("%s.p%d.tn_%d.png", base, pageNum, width)
}
