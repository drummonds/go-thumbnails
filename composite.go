package thumbnails

import (
	"image"
	"image/color"

	"golang.org/x/image/draw"
)

// resizeToPage scales img to the given width, then crops or pads vertically
// to produce a fixed width × pageHeight(width) output (A4 aspect ratio).
// Tall images are cropped from the top; short images are placed at the top
// on a white background.
func resizeToPage(img image.Image, width uint) *image.RGBA {
	b := img.Bounds()
	srcW, srcH := b.Dx(), b.Dy()
	if srcW == 0 || srcH == 0 {
		return image.NewRGBA(image.Rect(0, 0, int(width), int(pageHeight(width))))
	}

	// Scale so image width == width, preserving aspect ratio.
	scaledH := int(float64(srcH) * float64(width) / float64(srcW))
	scaled := image.NewRGBA(image.Rect(0, 0, int(width), scaledH))
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), img, b, draw.Src, nil)

	ph := int(pageHeight(width))
	dst := image.NewRGBA(image.Rect(0, 0, int(width), ph))

	// White background
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	if scaledH >= ph {
		// Crop from top: take the top ph rows.
		srcRect := image.Rect(0, 0, int(width), ph)
		draw.Draw(dst, dst.Bounds(), scaled, srcRect.Min, draw.Src)
	} else {
		// Place at top, white fills the rest.
		dstRect := image.Rect(0, 0, int(width), scaledH)
		draw.Draw(dst, dstRect, scaled, scaled.Bounds().Min, draw.Src)
	}

	return dst
}

// compositePages creates a composite thumbnail from multiple page images.
// Each page is resized to width × pageHeight(width). Up to 4 pages are shown
// side-by-side. If there are more than 4 pages, a "+" indicator is appended.
func compositePages(pages []image.Image, width uint) image.Image {
	numPagesToShow := len(pages)
	showPlusIndicator := false
	if numPagesToShow > 4 {
		numPagesToShow = 4
		showPlusIndicator = true
	}

	ph := int(pageHeight(width))
	resizedPages := make([]*image.RGBA, numPagesToShow)

	for i := 0; i < numPagesToShow; i++ {
		resizedPages[i] = resizeToPage(pages[i], width)
	}

	totalWidth := numPagesToShow * int(width)
	plusWidth := 0
	if showPlusIndicator {
		plusWidth = int(width)
		totalWidth += plusWidth
	}

	composite := image.NewRGBA(image.Rect(0, 0, totalWidth, ph))

	// Fill with white background
	draw.Draw(composite, composite.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw each page thumbnail side by side
	currentX := 0
	for i := 0; i < numPagesToShow; i++ {
		bounds := resizedPages[i].Bounds()
		destRect := image.Rect(currentX, 0, currentX+bounds.Dx(), ph)
		draw.Draw(composite, destRect, resizedPages[i], bounds.Min, draw.Src)
		currentX += int(width)
	}

	if showPlusIndicator {
		drawPlusIndicator(composite, currentX, ph)
	}

	return composite
}

// drawPlusIndicator draws a simple "+" symbol in a rectangular area.
func drawPlusIndicator(img *image.RGBA, startX, height int) {
	bgColor := color.RGBA{240, 240, 240, 255}
	plusColor := color.RGBA{100, 100, 100, 255}

	// Use height as the indicator width for a square-ish plus area
	size := height

	// Draw background
	for y := 0; y < height; y++ {
		for x := startX; x < startX+size; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Draw "+" symbol - vertical line
	centerX := startX + size/2
	lineWidth := size / 8
	if lineWidth < 2 {
		lineWidth = 2
	}

	for y := size / 4; y < 3*size/4; y++ {
		for dx := -lineWidth / 2; dx <= lineWidth/2; dx++ {
			img.Set(centerX+dx, y, plusColor)
		}
	}

	// Horizontal line
	centerY := height / 2
	for x := startX + size/4; x < startX+3*size/4; x++ {
		for dy := -lineWidth / 2; dy <= lineWidth/2; dy++ {
			img.Set(x, centerY+dy, plusColor)
		}
	}
}
