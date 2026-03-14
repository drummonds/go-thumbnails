package thumbnails

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPageThumbnailPath(t *testing.T) {
	tests := []struct {
		docPath  string
		pageNum  int
		width    uint
		expected string
	}{
		{"doc.pdf", 3, 128, "doc.p3.tn_128.png"},
		{"/path/to/file.tiff", 1, 64, "/path/to/file.p1.tn_64.png"},
		{"image.jpg", 2, 32, "image.p2.tn_32.png"},
	}

	for _, tt := range tests {
		result := DefaultPageThumbnailPath(tt.docPath, tt.pageNum, tt.width)
		if result != tt.expected {
			t.Errorf("DefaultPageThumbnailPath(%q, %d, %d) = %q, want %q",
				tt.docPath, tt.pageNum, tt.width, result, tt.expected)
		}
	}
}

func TestRenderPagesPNG(t *testing.T) {
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "test.png")

	img := image.NewRGBA(image.Rect(0, 0, 100, 80))
	for y := range 80 {
		for x := range 100 {
			img.Set(x, y, color.RGBA{0, 128, 0, 255})
		}
	}
	f, err := os.Create(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()

	results, err := RenderPages(pngPath)
	if err != nil {
		t.Fatalf("RenderPages failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 page, got %d", len(results))
	}
	if results[0].PageNum != 1 {
		t.Errorf("expected PageNum 1, got %d", results[0].PageNum)
	}
	if results[0].PageCount != 1 {
		t.Errorf("expected PageCount 1, got %d", results[0].PageCount)
	}
	if results[0].Image.Bounds().Dx() != 100 || results[0].Image.Bounds().Dy() != 80 {
		t.Errorf("unexpected dimensions: %v", results[0].Image.Bounds())
	}
}

func TestRenderPagePNG(t *testing.T) {
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "test.png")

	img := image.NewRGBA(image.Rect(0, 0, 100, 80))
	for y := range 80 {
		for x := range 100 {
			img.Set(x, y, color.White)
		}
	}
	f, err := os.Create(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()

	result, err := RenderPage(pngPath, 1)
	if err != nil {
		t.Fatalf("RenderPage failed: %v", err)
	}
	if result.PageNum != 1 || result.PageCount != 1 {
		t.Errorf("expected page 1/1, got %d/%d", result.PageNum, result.PageCount)
	}
}

func TestRenderPageOutOfRange(t *testing.T) {
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "test.png")

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	f, err := os.Create(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()

	_, err = RenderPage(pngPath, 0)
	if !errors.Is(err, ErrPageOutOfRange) {
		t.Errorf("page 0: expected ErrPageOutOfRange, got %v", err)
	}

	_, err = RenderPage(pngPath, 2)
	if !errors.Is(err, ErrPageOutOfRange) {
		t.Errorf("page 2: expected ErrPageOutOfRange, got %v", err)
	}
}

func TestResizePage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 300))
	for y := range 300 {
		for x := range 200 {
			img.Set(x, y, color.RGBA{100, 100, 100, 255})
		}
	}

	result := ResizePage(img, 64)
	if result.Bounds().Dx() != 64 {
		t.Errorf("expected width 64, got %d", result.Bounds().Dx())
	}
	expectedH := int(pageHeight(64))
	if result.Bounds().Dy() != expectedH {
		t.Errorf("expected height %d, got %d", expectedH, result.Bounds().Dy())
	}
}

func TestRenderPagesPDF(t *testing.T) {
	if !hasTestdata() {
		t.Skip("testdata/ not found")
	}

	path := filepath.Join(testdataDir(), "5-twopage.pdf")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("5-twopage.pdf not found")
	}

	results, err := RenderPages(path)
	if err != nil {
		t.Fatalf("RenderPages failed: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected >= 2 pages, got %d", len(results))
	}
	for i, r := range results {
		if r.PageNum != i+1 {
			t.Errorf("page %d: expected PageNum %d, got %d", i, i+1, r.PageNum)
		}
		if r.PageCount != len(results) {
			t.Errorf("page %d: expected PageCount %d, got %d", i, len(results), r.PageCount)
		}
	}
}

func TestRenderPagePDF(t *testing.T) {
	if !hasTestdata() {
		t.Skip("testdata/ not found")
	}

	path := filepath.Join(testdataDir(), "5-twopage.pdf")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("5-twopage.pdf not found")
	}

	result, err := RenderPage(path, 2)
	if err != nil {
		t.Fatalf("RenderPage(2) failed: %v", err)
	}
	if result.PageNum != 2 {
		t.Errorf("expected PageNum 2, got %d", result.PageNum)
	}
}
