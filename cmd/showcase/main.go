// Command showcase generates sample thumbnails into a temporary directory
// so you can visually inspect the output of every supported path.
//
// Usage:
//
//	go run ./cmd/showcase [-width 64] [-output /tmp/thumbnails]
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	thumbnails "github.com/drummonds/go-thumbnails"
)

func main() {
	width := flag.Uint("width", 64, "Thumbnail width in pixels")
	outputDir := flag.String("output", "", "Output directory (default: auto-created temp dir)")
	flag.Parse()

	dir := *outputDir
	if dir == "" {
		var err error
		dir, err = os.MkdirTemp("", "go-thumbnails-showcase-")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create temp dir: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create output dir: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Showcase output: %s  (width=%d)\n\n", dir, *width)

	// --- 1. All supported files from testdata/ ---
	testdata := filepath.Join(".", "testdata")
	if info, err := os.Stat(testdata); err == nil && info.IsDir() {
		var allFiles []string
		for _, pattern := range []string{"*.pdf", "*.tif", "*.tiff", "*.png", "*.jpg", "*.jpeg"} {
			matches, _ := filepath.Glob(filepath.Join(testdata, pattern))
			allFiles = append(allFiles, matches...)
		}
		sort.Strings(allFiles)

		if len(allFiles) > 0 {
			fmt.Println("=== Test data files ===")
			for _, f := range allFiles {
				base := filepath.Base(f)
				outName := strings.TrimSuffix(base, filepath.Ext(base)) + "_thumb.png"
				outPath := filepath.Join(dir, outName)
				err := thumbnails.GenerateAndSave(f, outPath, *width)
				if err != nil {
					fmt.Printf("  %-30s ERROR: %v\n", base, err)
				} else {
					info, _ := os.Stat(outPath)
					fmt.Printf("  %-30s -> %s (%d bytes)\n", base, outName, info.Size())
				}
			}
			fmt.Println()
		}
	} else {
		fmt.Println("=== testdata/ not found, skipping real document thumbnails ===")
	}

	// --- 2. Synthetic images (PNG, JPG) ---
	fmt.Println("=== Synthetic images ===")
	synthDir := filepath.Join(dir, "_synthetic_sources")
	os.MkdirAll(synthDir, 0755)

	// Red 200x100 PNG
	saveSynthThumb(synthDir, dir, "red_200x100.png", createSolidPNG(200, 100, color.RGBA{220, 40, 40, 255}), *width)

	// Green 80x120 PNG (portrait)
	saveSynthThumb(synthDir, dir, "green_80x120.png", createSolidPNG(80, 120, color.RGBA{40, 180, 40, 255}), *width)

	// Blue 500x500 PNG (square)
	saveSynthThumb(synthDir, dir, "blue_500x500.png", createSolidPNG(500, 500, color.RGBA{40, 40, 220, 255}), *width)

	// Gradient 300x200 PNG
	saveSynthThumb(synthDir, dir, "gradient_300x200.png", createGradientPNG(300, 200), *width)

	// Tiny 10x10 PNG
	saveSynthThumb(synthDir, dir, "tiny_10x10.png", createSolidPNG(10, 10, color.RGBA{255, 165, 0, 255}), *width)

	fmt.Println()

	// --- 3. Error placeholders ---
	fmt.Println("=== Error placeholders ===")
	placeholders := []struct {
		label    string
		filename string
	}{
		{"Password Protected", "placeholder_password.png"},
		{"Unsupported Format", "placeholder_unsupported.png"},
		{"File Not Found", "placeholder_notfound.png"},
		{"Error", "placeholder_error.png"},
	}
	for _, p := range placeholders {
		img := thumbnails.ErrorPlaceholder(p.label, *width)
		outPath := filepath.Join(dir, p.filename)
		if err := savePNG(img, outPath); err != nil {
			fmt.Printf("  %-30s ERROR: %v\n", p.label, err)
		} else {
			info, _ := os.Stat(outPath)
			fmt.Printf("  %-30s -> %s (%d bytes)\n", p.label, p.filename, info.Size())
		}
	}

	fmt.Printf("\nDone. View results in:\n  %s\n", dir)
}

// createSolidPNG creates a solid-colour PNG image in memory and returns its path after saving.
func createSolidPNG(w, h int, c color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}
	return img
}

// createGradientPNG creates a diagonal gradient image.
func createGradientPNG(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			r := uint8(255 * x / w)
			g := uint8(255 * y / h)
			b := uint8(128)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	return img
}

// saveSynthThumb saves a synthetic source image, then generates and saves its thumbnail.
func saveSynthThumb(synthDir, outDir, name string, img image.Image, width uint) {
	srcPath := filepath.Join(synthDir, name)
	if err := savePNG(img, srcPath); err != nil {
		fmt.Printf("  %-30s ERROR creating source: %v\n", name, err)
		return
	}

	outName := strings.TrimSuffix(name, filepath.Ext(name)) + "_thumb.png"
	outPath := filepath.Join(outDir, outName)
	if err := thumbnails.GenerateAndSave(srcPath, outPath, width); err != nil {
		fmt.Printf("  %-30s ERROR: %v\n", name, err)
		return
	}
	info, _ := os.Stat(outPath)
	fmt.Printf("  %-30s -> %s (%d bytes)\n", name, outName, info.Size())
}

func savePNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
