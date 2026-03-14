// Command gendocimages generates sample thumbnail images for the documentation gallery.
//
// Usage:
//
//	go run ./cmd/gendocimages [-output docs/images]
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	thumbnails "github.com/drummonds/go-thumbnails"
)

func main() {
	outputDir := flag.String("output", "docs/images", "Output directory for gallery images")
	flag.Parse()

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}

	tmpDir, err := os.MkdirTemp("", "gendocimages-src-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "tmpdir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	// Source images: landscape, portrait, square
	sources := []struct {
		name string
		w, h int
		col  color.RGBA
	}{
		{"landscape", 300, 200, color.RGBA{60, 120, 200, 255}},
		{"portrait", 200, 400, color.RGBA{200, 80, 60, 255}},
		{"square", 256, 256, color.RGBA{60, 180, 80, 255}},
	}

	srcPaths := make(map[string]string)
	for _, s := range sources {
		img := gradientImage(s.w, s.h, s.col)
		p := filepath.Join(tmpDir, s.name+".png")
		mustSavePNG(img, p)
		srcPaths[s.name] = p

		// Save source image to docs for reference
		mustSavePNG(img, filepath.Join(*outputDir, "source_"+s.name+".png"))
	}

	fmt.Println("=== Composite style ===")
	for _, s := range sources {
		out := filepath.Join(*outputDir, "composite_"+s.name+".png")
		must(thumbnails.GenerateAndSave(srcPaths[s.name], out, 128))
		fmt.Printf("  %s\n", filepath.Base(out))
	}

	fmt.Println("=== Uniform style ===")
	for _, s := range sources {
		out := filepath.Join(*outputDir, "uniform_"+s.name+".png")
		must(thumbnails.GenerateStyledAndSave(srcPaths[s.name], out, 128, thumbnails.StyleUniform))
		fmt.Printf("  %s\n", filepath.Base(out))
	}

	fmt.Println("=== Width comparison ===")
	for _, w := range []uint{64, 128, 256} {
		out := filepath.Join(*outputDir, fmt.Sprintf("width_%d.png", w))
		must(thumbnails.GenerateAndSave(srcPaths["landscape"], out, w))
		fmt.Printf("  %s\n", filepath.Base(out))
	}

	fmt.Println("=== Per-page rendering ===")
	// Single-page image: show the page-level API output
	pr, err := thumbnails.RenderPage(srcPaths["landscape"], 1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "RenderPage: %v\n", err)
		os.Exit(1)
	}
	resized := thumbnails.ResizePage(pr.Image, 128)
	mustSavePNG(resized, filepath.Join(*outputDir, "page_single.png"))
	fmt.Println("  page_single.png")

	fmt.Println("=== Error placeholders ===")
	placeholders := []struct {
		label, file string
	}{
		{"Password Protected", "placeholder_password.png"},
		{"Unsupported Format", "placeholder_unsupported.png"},
		{"File Not Found", "placeholder_notfound.png"},
		{"Error", "placeholder_error.png"},
	}
	for _, p := range placeholders {
		img := thumbnails.ErrorPlaceholder(p.label, 128)
		mustSavePNG(img, filepath.Join(*outputDir, p.file))
		fmt.Printf("  %s\n", p.file)
	}

	// --- PDF examples (requires testdata/) ---
	testdata := filepath.Join(".", "testdata")
	if info, err := os.Stat(testdata); err == nil && info.IsDir() {
		fmt.Println("=== PDF examples ===")
		pdfFiles := []struct {
			name, file string
		}{
			{"singlepage", "hello.pdf"},
			{"twopage", "twopage.pdf"},
			{"fivepage", "fivepage.pdf"},
		}
		for _, pdf := range pdfFiles {
			src := filepath.Join(testdata, pdf.file)
			if _, err := os.Stat(src); err != nil {
				fmt.Printf("  skip %s (not found)\n", pdf.file)
				continue
			}

			// Composite
			out := filepath.Join(*outputDir, "pdf_composite_"+pdf.name+".png")
			if err := thumbnails.GenerateAndSave(src, out, 128); err != nil {
				fmt.Printf("  %s composite ERROR: %v\n", pdf.name, err)
			} else {
				fmt.Printf("  pdf_composite_%s.png\n", pdf.name)
			}

			// Uniform
			out = filepath.Join(*outputDir, "pdf_uniform_"+pdf.name+".png")
			if err := thumbnails.GenerateStyledAndSave(src, out, 128, thumbnails.StyleUniform); err != nil {
				fmt.Printf("  %s uniform ERROR: %v\n", pdf.name, err)
			} else {
				fmt.Printf("  pdf_uniform_%s.png\n", pdf.name)
			}
		}

		// Per-page: all pages of fivepage.pdf
		fivepage := filepath.Join(testdata, "fivepage.pdf")
		if _, err := os.Stat(fivepage); err == nil {
			fmt.Println("=== PDF per-page ===")
			pages, err := thumbnails.RenderPages(fivepage)
			if err != nil {
				fmt.Printf("  RenderPages ERROR: %v\n", err)
			} else {
				for _, p := range pages {
					resized := thumbnails.ResizePage(p.Image, 128)
					out := filepath.Join(*outputDir, fmt.Sprintf("pdf_page_%d_of_%d.png", p.PageNum, p.PageCount))
					mustSavePNG(resized, out)
					fmt.Printf("  pdf_page_%d_of_%d.png\n", p.PageNum, p.PageCount)
				}
			}

			// Single page extraction
			p, err := thumbnails.RenderPage(fivepage, 3)
			if err == nil {
				resized := thumbnails.ResizePage(p.Image, 128)
				mustSavePNG(resized, filepath.Join(*outputDir, "pdf_page3_only.png"))
				fmt.Println("  pdf_page3_only.png")
			}
		}
	} else {
		fmt.Println("=== testdata/ not found, skipping PDF examples ===")
	}

	fmt.Printf("\nDone. Images in %s\n", *outputDir)
}

func gradientImage(w, h int, base color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			r := uint8(min(255, int(base.R)+x*80/w))
			g := uint8(min(255, int(base.G)+y*80/h))
			b := base.B
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	return img
}

func mustSavePNG(img image.Image, path string) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create %s: %v\n", path, err)
		os.Exit(1)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		fmt.Fprintf(os.Stderr, "encode %s: %v\n", path, err)
		os.Exit(1)
	}
}

func must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
