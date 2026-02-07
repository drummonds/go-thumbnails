package thumbnails

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/nfnt/resize"
)

// compositePages creates a composite thumbnail from multiple page images.
// It shows up to 4 pages side-by-side, resized to the target height.
// If there are more than 4 pages, a "+" indicator is appended.
func compositePages(pages []image.Image, height uint) image.Image {
	numPagesToShow := len(pages)
	showPlusIndicator := false
	if numPagesToShow > 4 {
		numPagesToShow = 4
		showPlusIndicator = true
	}

	resizedPages := make([]image.Image, numPagesToShow)
	totalWidth := 0

	for i := 0; i < numPagesToShow; i++ {
		resized := resize.Resize(0, height, pages[i], resize.Lanczos3)
		resizedPages[i] = resized
		totalWidth += resized.Bounds().Dx()
	}

	plusWidth := 0
	if showPlusIndicator {
		plusWidth = int(height)
		totalWidth += plusWidth
	}

	composite := image.NewRGBA(image.Rect(0, 0, totalWidth, int(height)))

	// Fill with white background
	draw.Draw(composite, composite.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw each page thumbnail side by side
	currentX := 0
	for i := 0; i < numPagesToShow; i++ {
		bounds := resizedPages[i].Bounds()
		destRect := image.Rect(currentX, 0, currentX+bounds.Dx(), int(height))
		draw.Draw(composite, destRect, resizedPages[i], bounds.Min, draw.Src)
		currentX += bounds.Dx()
	}

	if showPlusIndicator {
		drawPlusIndicator(composite, currentX, int(height))
	}

	return composite
}

// drawPlusIndicator draws a simple "+" symbol in a square area.
func drawPlusIndicator(img *image.RGBA, startX, size int) {
	bgColor := color.RGBA{240, 240, 240, 255}
	plusColor := color.RGBA{100, 100, 100, 255}

	// Draw background
	for y := 0; y < size; y++ {
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
	centerY := size / 2
	for x := startX + size/4; x < startX+3*size/4; x++ {
		for dy := -lineWidth / 2; dy <= lineWidth/2; dy++ {
			img.Set(x, centerY+dy, plusColor)
		}
	}
}
