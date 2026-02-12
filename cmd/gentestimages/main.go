// Command gentestimages creates PNG and JPEG test images in testdata/.
//
// Usage:
//
//	go run ./cmd/gentestimages
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	dir := filepath.Join(".", "testdata")
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create testdata/: %v\n", err)
		os.Exit(1)
	}

	images := []struct {
		name string
		img  image.Image
	}{
		{"landscape_300x200.png", solidImg(300, 200, color.RGBA{50, 120, 200, 255})},
		{"portrait_200x400.png", gradientImg(200, 400)},
		{"square_256x256.png", checkerboard(256, 256, 32)},
		{"tiny_8x8.png", solidImg(8, 8, color.RGBA{255, 100, 0, 255})},
		{"large_1200x800.png", gradientImg(1200, 800)},
		{"photo_400x300.jpg", sceneImg(400, 300)},
		{"portrait_300x500.jpg", sceneImg(300, 500)},
		{"small_50x50.jpg", solidImg(50, 50, color.RGBA{200, 50, 50, 255})},
	}

	for _, entry := range images {
		path := filepath.Join(dir, entry.name)
		if err := saveImage(path, entry.img); err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR %s: %v\n", entry.name, err)
			continue
		}
		info, _ := os.Stat(path)
		fmt.Printf("  %-30s %d bytes\n", entry.name, info.Size())
	}

	fmt.Println("\nDone.")
}

func saveImage(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Encode(f, img, &jpeg.Options{Quality: 85})
	case ".png":
		return png.Encode(f, img)
	default:
		return fmt.Errorf("unsupported extension: %s", ext)
	}
}

func solidImg(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}
	return img
}

func gradientImg(w, h int) *image.RGBA {
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

func checkerboard(w, h, size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			if ((x/size)+(y/size))%2 == 0 {
				img.Set(x, y, color.RGBA{240, 240, 240, 255})
			} else {
				img.Set(x, y, color.RGBA{60, 60, 60, 255})
			}
		}
	}
	return img
}

// sceneImg creates a simple "scene" with sky, ground, and a shape to give
// JPEG compression something interesting to work with.
func sceneImg(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	horizon := h * 2 / 3

	for y := range h {
		for x := range w {
			if y < horizon {
				// Sky gradient
				t := float64(y) / float64(horizon)
				r := uint8(100 + 100*t)
				g := uint8(150 + 60*t)
				b := uint8(255 - 30*t)
				img.Set(x, y, color.RGBA{r, g, b, 255})
			} else {
				// Ground
				t := float64(y-horizon) / float64(h-horizon)
				g := uint8(120 - 40*t)
				img.Set(x, y, color.RGBA{80, g, 40, 255})
			}
		}
	}

	// Draw a rectangle in the centre (building/shape)
	rx, ry := w/3, horizon-h/4
	rw, rh := w/4, h/4
	for y := ry; y < ry+rh; y++ {
		for x := rx; x < rx+rw; x++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				img.Set(x, y, color.RGBA{180, 160, 140, 255})
			}
		}
	}

	return img
}
