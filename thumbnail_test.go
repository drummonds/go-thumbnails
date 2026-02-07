package thumbnails

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func testdataDir() string {
	return filepath.Join(".", "testdata")
}

func hasTestdata() bool {
	_, err := os.Stat(testdataDir())
	return err == nil
}

func TestDefaultThumbnailPath(t *testing.T) {
	tests := []struct {
		docPath  string
		height   uint
		expected string
	}{
		{"doc.pdf", 64, "doc.tn_64.png"},
		{"/path/to/file.tiff", 128, "/path/to/file.tn_128.png"},
		{"image.jpg", 32, "image.tn_32.png"},
	}

	for _, tt := range tests {
		result := DefaultThumbnailPath(tt.docPath, tt.height)
		if result != tt.expected {
			t.Errorf("DefaultThumbnailPath(%q, %d) = %q, want %q", tt.docPath, tt.height, result, tt.expected)
		}
	}
}

func TestGenerateUnsupportedFormat(t *testing.T) {
	_, err := Generate("test.xyz", 64)
	if err == nil {
		t.Error("expected error for unsupported format, got nil")
	}
}

func TestGeneratePDF(t *testing.T) {
	if !hasTestdata() {
		t.Skip("testdata/ not found, skipping PDF tests")
	}

	tests := []struct {
		name     string
		file     string
		height   uint
		minWidth int
	}{
		{"empty PDF", "1-empty.pdf", 64, 1},
		{"hello PDF", "2-hello.pdf", 64, 1},
		{"diagram PDF", "3-diagram.pdf", 64, 1},
		{"long text PDF", "4-longtext.pdf", 64, 1},
		{"two page PDF", "5-twopage.pdf", 64, 2},
		{"five page PDF (4+plus)", "6-fivepage.pdf", 64, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(testdataDir(), tt.file)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Skipf("test file %s not found", tt.file)
			}

			img, err := Generate(path, tt.height)
			if err != nil {
				t.Fatalf("Generate(%q, %d) failed: %v", path, tt.height, err)
			}

			bounds := img.Bounds()
			if uint(bounds.Dy()) != tt.height {
				t.Errorf("expected height %d, got %d", tt.height, bounds.Dy())
			}
			if bounds.Dx() < tt.minWidth {
				t.Errorf("expected width >= %d, got %d", tt.minWidth, bounds.Dx())
			}
		})
	}
}

func TestGenerateAndSavePDF(t *testing.T) {
	if !hasTestdata() {
		t.Skip("testdata/ not found, skipping PDF save tests")
	}

	path := filepath.Join(testdataDir(), "2-hello.pdf")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test file 2-hello.pdf not found")
	}

	outputPath := filepath.Join(t.TempDir(), "hello_thumb.png")
	err := GenerateAndSave(path, outputPath, 64)
	if err != nil {
		t.Fatalf("GenerateAndSave failed: %v", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
}

func TestGenerateImagePNG(t *testing.T) {
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "test.png")

	// Create a 100x80 white test PNG
	img := image.NewRGBA(image.Rect(0, 0, 100, 80))
	for y := 0; y < 80; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.White)
		}
	}
	f, err := os.Create(pngPath)
	if err != nil {
		t.Fatalf("failed to create test PNG: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("failed to encode test PNG: %v", err)
	}
	f.Close()

	thumb, err := Generate(pngPath, 40)
	if err != nil {
		t.Fatalf("Generate for PNG failed: %v", err)
	}

	if thumb.Bounds().Dy() != 40 {
		t.Errorf("expected height 40, got %d", thumb.Bounds().Dy())
	}
	// Width should scale proportionally: 100 * 40/80 = 50
	if thumb.Bounds().Dx() != 50 {
		t.Errorf("expected width 50, got %d", thumb.Bounds().Dx())
	}
}

func TestCheckPageCorruptionClean(t *testing.T) {
	// A normal opaque RGBA image should not be flagged as corrupt
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{50, 50, 50, 255})
		}
	}
	result := CheckPageCorruption(img)
	if result.Corrupt {
		t.Errorf("clean image flagged as corrupt: %s", result.Reason)
	}
}

func TestCheckPageCorruptionBadAlpha(t *testing.T) {
	// Simulate the PDFium corruption: rows with random non-255 alpha
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			if y < 20 { // 20% of rows are corrupt
				img.Set(x, y, color.RGBA{0x26, 0xa0, 0x3a, 0x07})
			} else {
				img.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}
	result := CheckPageCorruption(img)
	if !result.Corrupt {
		t.Error("corrupt image not detected")
	}
	if result.CorruptRowFraction < 0.10 {
		t.Errorf("expected >=10%% corrupt rows, got %.1f%%", result.CorruptRowFraction*100)
	}
}

func TestGenerateAndSaveImage(t *testing.T) {
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "test.png")
	outputPath := filepath.Join(tmpDir, "thumb.png")

	// Create a test PNG
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 200; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	f, err := os.Create(pngPath)
	if err != nil {
		t.Fatalf("failed to create test PNG: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("failed to encode test PNG: %v", err)
	}
	f.Close()

	err = GenerateAndSave(pngPath, outputPath, 50)
	if err != nil {
		t.Fatalf("GenerateAndSave failed: %v", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
}
