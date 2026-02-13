package thumbnails

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// uniformHeight returns the height for a uniform-style thumbnail: round(1.42 × width).
func uniformHeight(width uint) uint {
	return uint(math.Round(float64(width) * 1.42))
}

// uniformPage creates a fixed-size width × uniformHeight(width) thumbnail.
// The first page is scaled to fill the width and cropped/padded to the uniform height.
// If pageCount > 1, a page-count badge is drawn in the bottom-right corner.
func uniformPage(firstPage image.Image, pageCount int, width uint) image.Image {
	w := int(width)
	h := int(uniformHeight(width))
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	// Light grey background to show padding
	draw.Draw(dst, dst.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Scale first page to fill width, preserving aspect ratio
	b := firstPage.Bounds()
	srcW, srcH := b.Dx(), b.Dy()
	if srcW > 0 && srcH > 0 {
		scaledH := int(float64(srcH) * float64(width) / float64(srcW))
		scaled := image.NewRGBA(image.Rect(0, 0, w, scaledH))
		draw.CatmullRom.Scale(scaled, scaled.Bounds(), firstPage, b, draw.Src, nil)

		if scaledH >= h {
			// Crop from top
			draw.Draw(dst, dst.Bounds(), scaled, image.Point{}, draw.Src)
		} else {
			// Place at top, white fills below
			dstRect := image.Rect(0, 0, w, scaledH)
			draw.Draw(dst, dstRect, scaled, scaled.Bounds().Min, draw.Src)
		}
	}

	if pageCount > 1 {
		drawPageCountBadge(dst, pageCount)
	}

	return dst
}

// drawPageCountBadge draws a page-count indicator in the bottom-right corner.
// Shows "2".."9" for 2–9 pages, "9+" for more than 9 pages.
func drawPageCountBadge(img *image.RGBA, pageCount int) {
	var label string
	if pageCount > 9 {
		label = "9+"
	} else {
		label = fmt.Sprintf("%d", pageCount)
	}

	face := basicfont.Face7x13
	textWidth := font.MeasureString(face, label).Ceil()
	ascent := face.Metrics().Ascent.Ceil()

	padding := 3
	badgeW := textWidth + padding*2
	badgeH := ascent + padding*2

	imgW := img.Bounds().Dx()
	imgH := img.Bounds().Dy()

	margin := 2
	badgeX := imgW - badgeW - margin
	badgeY := imgH - badgeH - margin

	// Draw semi-transparent dark background (70% black overlay)
	for y := badgeY; y < badgeY+badgeH; y++ {
		for x := badgeX; x < badgeX+badgeW; x++ {
			if x >= 0 && y >= 0 && x < imgW && y < imgH {
				existing := img.RGBAAt(x, y)
				r := uint8(float64(existing.R) * 0.3)
				g := uint8(float64(existing.G) * 0.3)
				b := uint8(float64(existing.B) * 0.3)
				img.Set(x, y, color.RGBA{r, g, b, 255})
			}
		}
	}

	// Draw white text
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.White),
		Face: face,
		Dot:  fixed.P(badgeX+padding, badgeY+padding+ascent),
	}
	d.DrawString(label)
}
