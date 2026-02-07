package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/drummonds/go-thumbnails/pdfrenderer"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: diagnose <pdf> <output-dir>\n")
		os.Exit(1)
	}
	pdfPath := os.Args[1]
	outDir := os.Args[2]
	os.MkdirAll(outDir, 0755)

	renderer, err := pdfrenderer.NewPDFiumRenderer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "renderer error: %v\n", err)
		os.Exit(1)
	}
	defer renderer.Close()

	pages, err := renderer.RenderPDF(pdfPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Pages: %d\n", len(pages))
	for i, pg := range pages {
		b := pg.Bounds()
		fmt.Printf("\nPage %d: %dx%d type=%T\n", i, b.Dx(), b.Dy(), pg)

		rgba, ok := pg.(*image.RGBA)
		if !ok {
			fmt.Printf("  Not *image.RGBA, skipping deep analysis\n")
			continue
		}

		fmt.Printf("  Stride: %d  Pix len: %d  Expected pix: %d\n",
			rgba.Stride, len(rgba.Pix), b.Dx()*b.Dy()*4)

		// Find first corrupt row and dump raw bytes
		w := b.Dx()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			rowStart := (y-b.Min.Y)*rgba.Stride
			hasNonGray := false
			hasNon255Alpha := false
			for x := 0; x < w; x++ {
				off := rowStart + x*4
				r, g, bv, a := rgba.Pix[off], rgba.Pix[off+1], rgba.Pix[off+2], rgba.Pix[off+3]
				if r != g || g != bv {
					hasNonGray = true
				}
				if a != 255 {
					hasNon255Alpha = true
				}
			}
			if hasNonGray {
				fmt.Printf("  First corrupt row: y=%d (hasNon255Alpha=%v)\n", y, hasNon255Alpha)
				// Dump first 40 bytes of this row
				rowStart := (y-b.Min.Y)*rgba.Stride
				fmt.Printf("  Raw bytes[0:40]: ")
				for j := 0; j < 40 && j < len(rgba.Pix)-rowStart; j++ {
					fmt.Printf("%02x ", rgba.Pix[rowStart+j])
				}
				fmt.Println()

				// Also dump a known-good row before it
				if y > b.Min.Y {
					goodRow := (y-1-b.Min.Y)*rgba.Stride
					fmt.Printf("  Good row y=%d bytes[0:40]: ", y-1)
					for j := 0; j < 40 && j < len(rgba.Pix)-goodRow; j++ {
						fmt.Printf("%02x ", rgba.Pix[goodRow+j])
					}
					fmt.Println()
				}

				// Check: are the corrupt rows the ones with images/graphics?
				// Count non-255 alpha pixels in this row
				nonOpaqueCount := 0
				for x := 0; x < w; x++ {
					off := rowStart + x*4
					if rgba.Pix[off+3] != 255 {
						nonOpaqueCount++
					}
				}
				fmt.Printf("  Non-opaque pixels in this row: %d/%d\n", nonOpaqueCount, w)
				break
			}
		}

		// Count total corrupt vs clean rows
		corruptRows := 0
		alphaRows := 0
		for y := b.Min.Y; y < b.Max.Y; y++ {
			rowStart := (y-b.Min.Y)*rgba.Stride
			rowCorrupt := false
			rowHasAlpha := false
			for x := 0; x < w; x += max(1, w/100) {
				off := rowStart + x*4
				r, g, bv, a := rgba.Pix[off], rgba.Pix[off+1], rgba.Pix[off+2], rgba.Pix[off+3]
				if r != g || g != bv {
					rowCorrupt = true
				}
				if a != 255 {
					rowHasAlpha = true
				}
			}
			if rowCorrupt {
				corruptRows++
			}
			if rowHasAlpha {
				alphaRows++
			}
		}
		fmt.Printf("  Corrupt rows: %d/%d  Alpha rows: %d/%d\n", corruptRows, b.Dy(), alphaRows, b.Dy())

		// Save the raw page
		savePNG(pg, fmt.Sprintf("%s/page_%d_raw.png", outDir, i))
	}
}

func savePNG(img image.Image, path string) {
	f, _ := os.Create(path)
	defer f.Close()
	png.Encode(f, img)
}
