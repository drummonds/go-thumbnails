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
		width    uint
		expected string
	}{
		{"doc.pdf", 64, "doc.tn_64.png"},
		{"/path/to/file.tiff", 128, "/path/to/file.tn_128.png"},
		{"image.jpg", 32, "image.tn_32.png"},
	}

	for _, tt := range tests {
		result := DefaultThumbnailPath(tt.docPath, tt.width)
		if result != tt.expected {
			t.Errorf("DefaultThumbnailPath(%q, %d) = %q, want %q", tt.docPath, tt.width, result, tt.expected)
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
		width    uint
		minPages int // minimum number of page tiles expected
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

			img, err := Generate(path, tt.width)
			if err != nil {
				t.Fatalf("Generate(%q, %d) failed: %v", path, tt.width, err)
			}

			bounds := img.Bounds()
			expectedHeight := int(pageHeight(tt.width))
			if bounds.Dy() != expectedHeight {
				t.Errorf("expected height %d, got %d", expectedHeight, bounds.Dy())
			}
			minWidth := tt.minPages * int(tt.width)
			if bounds.Dx() < minWidth {
				t.Errorf("expected width >= %d, got %d", minWidth, bounds.Dx())
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

	// width=50: image scaled to 50×40, padded to 50×pageHeight(50)=50×71
	thumb, err := Generate(pngPath, 50)
	if err != nil {
		t.Fatalf("Generate for PNG failed: %v", err)
	}

	expectedW := 50
	expectedH := int(pageHeight(50)) // 71
	if thumb.Bounds().Dx() != expectedW {
		t.Errorf("expected width %d, got %d", expectedW, thumb.Bounds().Dx())
	}
	if thumb.Bounds().Dy() != expectedH {
		t.Errorf("expected height %d, got %d", expectedH, thumb.Bounds().Dy())
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

func TestGenerateOrPlaceholderSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "test.png")

	// Create a 100x80 test PNG
	img := image.NewRGBA(image.Rect(0, 0, 100, 80))
	for y := range 80 {
		for x := range 100 {
			img.Set(x, y, color.RGBA{0, 128, 0, 255})
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

	thumb := GenerateOrPlaceholder(pngPath, 50)
	if thumb == nil {
		t.Fatal("GenerateOrPlaceholder returned nil for valid image")
	}
	bounds := thumb.Bounds()
	expectedW := 50
	expectedH := int(pageHeight(50))
	if bounds.Dx() != expectedW {
		t.Errorf("expected width %d, got %d", expectedW, bounds.Dx())
	}
	if bounds.Dy() != expectedH {
		t.Errorf("expected height %d, got %d", expectedH, bounds.Dy())
	}
}

func TestGenerateOrPlaceholderUnsupported(t *testing.T) {
	thumb := GenerateOrPlaceholder("test.xyz", 64)
	if thumb == nil {
		t.Fatal("GenerateOrPlaceholder returned nil for unsupported format")
	}
	bounds := thumb.Bounds()
	expectedW := 64
	expectedH := int(pageHeight(64))
	if bounds.Dx() != expectedW || bounds.Dy() != expectedH {
		t.Errorf("expected %dx%d placeholder, got %dx%d", expectedW, expectedH, bounds.Dx(), bounds.Dy())
	}
}

func TestErrorPlaceholder(t *testing.T) {
	tests := []struct {
		label string
		width uint
	}{
		{"Password Protected", 64},
		{"Unsupported Format", 128},
		{"File Not Found", 32},
		{"Error", 64},
	}
	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			img := ErrorPlaceholder(tt.label, tt.width)
			if img == nil {
				t.Fatal("ErrorPlaceholder returned nil")
			}
			bounds := img.Bounds()
			expectedW := int(tt.width)
			expectedH := int(pageHeight(tt.width))
			if bounds.Dx() != expectedW || bounds.Dy() != expectedH {
				t.Errorf("expected %dx%d, got %dx%d", expectedW, expectedH, bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestGenerateTestdataPNG(t *testing.T) {
	if !hasTestdata() {
		t.Skip("testdata/ not found, skipping image tests")
	}

	tests := []struct {
		name  string
		file  string
		width uint
	}{
		{"landscape PNG", "landscape_300x200.png", 64},
		{"portrait PNG", "portrait_200x400.png", 64},
		{"square PNG", "square_256x256.png", 64},
		{"tiny PNG", "tiny_8x8.png", 64},
		{"large PNG", "large_1200x800.png", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(testdataDir(), tt.file)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Skipf("test file %s not found", tt.file)
			}

			img, err := Generate(path, tt.width)
			if err != nil {
				t.Fatalf("Generate(%q, %d) failed: %v", tt.file, tt.width, err)
			}

			bounds := img.Bounds()
			expectedW := int(tt.width)
			expectedH := int(pageHeight(tt.width))
			if bounds.Dx() != expectedW {
				t.Errorf("expected width %d, got %d", expectedW, bounds.Dx())
			}
			if bounds.Dy() != expectedH {
				t.Errorf("expected height %d, got %d", expectedH, bounds.Dy())
			}
		})
	}
}

func TestGenerateTestdataJPEG(t *testing.T) {
	if !hasTestdata() {
		t.Skip("testdata/ not found, skipping JPEG tests")
	}

	tests := []struct {
		name  string
		file  string
		width uint
	}{
		{"photo JPEG", "photo_400x300.jpg", 64},
		{"portrait JPEG", "portrait_300x500.jpg", 64},
		{"small JPEG", "small_50x50.jpg", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(testdataDir(), tt.file)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Skipf("test file %s not found", tt.file)
			}

			img, err := Generate(path, tt.width)
			if err != nil {
				t.Fatalf("Generate(%q, %d) failed: %v", tt.file, tt.width, err)
			}

			bounds := img.Bounds()
			expectedW := int(tt.width)
			expectedH := int(pageHeight(tt.width))
			if bounds.Dx() != expectedW {
				t.Errorf("expected width %d, got %d", expectedW, bounds.Dx())
			}
			if bounds.Dy() != expectedH {
				t.Errorf("expected height %d, got %d", expectedH, bounds.Dy())
			}
		})
	}
}

func TestGenerateAndSaveTestdataImage(t *testing.T) {
	if !hasTestdata() {
		t.Skip("testdata/ not found, skipping image save tests")
	}

	files := []string{"landscape_300x200.png", "photo_400x300.jpg"}
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			path := filepath.Join(testdataDir(), file)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Skipf("test file %s not found", file)
			}

			outputPath := filepath.Join(t.TempDir(), "thumb.png")
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
		})
	}
}

func TestUniformHeight(t *testing.T) {
	tests := []struct {
		width    uint
		expected uint
	}{
		{64, 91},   // round(64 * 1.42) = round(90.88) = 91
		{100, 142}, // round(100 * 1.42) = 142
		{128, 182}, // round(128 * 1.42) = round(181.76) = 182
	}
	for _, tt := range tests {
		result := uniformHeight(tt.width)
		if result != tt.expected {
			t.Errorf("uniformHeight(%d) = %d, want %d", tt.width, result, tt.expected)
		}
	}
}

func TestGenerateStyledUniformPNG(t *testing.T) {
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

	thumb, err := GenerateStyled(pngPath, 50, StyleUniform)
	if err != nil {
		t.Fatalf("GenerateStyled for PNG failed: %v", err)
	}

	expectedW := 50
	expectedH := int(uniformHeight(50)) // 71
	if thumb.Bounds().Dx() != expectedW {
		t.Errorf("expected width %d, got %d", expectedW, thumb.Bounds().Dx())
	}
	if thumb.Bounds().Dy() != expectedH {
		t.Errorf("expected height %d, got %d", expectedH, thumb.Bounds().Dy())
	}
}

func TestGenerateStyledUniformPDF(t *testing.T) {
	if !hasTestdata() {
		t.Skip("testdata/ not found, skipping PDF tests")
	}

	tests := []struct {
		name  string
		file  string
		width uint
	}{
		{"single page PDF", "2-hello.pdf", 64},
		{"two page PDF", "5-twopage.pdf", 64},
		{"five page PDF", "6-fivepage.pdf", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(testdataDir(), tt.file)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Skipf("test file %s not found", tt.file)
			}

			img, err := GenerateStyled(path, tt.width, StyleUniform)
			if err != nil {
				t.Fatalf("GenerateStyled(%q, %d, StyleUniform) failed: %v", path, tt.width, err)
			}

			bounds := img.Bounds()
			expectedW := int(tt.width)
			expectedH := int(uniformHeight(tt.width))
			if bounds.Dx() != expectedW {
				t.Errorf("expected width %d, got %d", expectedW, bounds.Dx())
			}
			if bounds.Dy() != expectedH {
				t.Errorf("expected height %d, got %d", expectedH, bounds.Dy())
			}
		})
	}
}

func TestGenerateStyledCompositeBackcompat(t *testing.T) {
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "test.png")

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

	orig, err := Generate(pngPath, 50)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	styled, err := GenerateStyled(pngPath, 50, StyleComposite)
	if err != nil {
		t.Fatalf("GenerateStyled failed: %v", err)
	}

	if orig.Bounds() != styled.Bounds() {
		t.Errorf("bounds differ: Generate=%v, GenerateStyled=%v", orig.Bounds(), styled.Bounds())
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
