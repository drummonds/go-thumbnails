package thumbnails

import (
	"image"
	"image/color"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// placeholderInfo holds the display label and background colour for an error type.
type placeholderInfo struct {
	label string
	bg    color.RGBA
}

// classifyError inspects an error and returns the appropriate placeholder info.
func classifyError(err error) placeholderInfo {
	msg := err.Error()

	switch {
	case strings.Contains(msg, "invalid password"):
		return placeholderInfo{"Password Protected", color.RGBA{200, 150, 0, 255}} // amber
	case strings.Contains(msg, "unsupported file format"):
		return placeholderInfo{"Unsupported Format", color.RGBA{130, 130, 130, 255}} // grey
	case strings.Contains(msg, "no such file") || strings.Contains(msg, "not exist"):
		return placeholderInfo{"File Not Found", color.RGBA{80, 80, 80, 255}} // dark grey
	default:
		return placeholderInfo{"Error", color.RGBA{180, 40, 40, 255}} // red
	}
}

// ErrorPlaceholder generates a coloured placeholder image with the given label.
// The image is square (height × height) with white centred text.
func ErrorPlaceholder(label string, height uint) image.Image {
	size := int(height)
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Fill background — pick colour from label, or default red.
	bg := bgForLabel(label)
	for y := range size {
		for x := range size {
			img.Set(x, y, bg)
		}
	}

	// Draw text centred in the image.
	drawCentredText(img, label, size)

	return img
}

// bgForLabel returns the background colour associated with a known label.
func bgForLabel(label string) color.RGBA {
	switch label {
	case "Password Protected":
		return color.RGBA{200, 150, 0, 255}
	case "Unsupported Format":
		return color.RGBA{130, 130, 130, 255}
	case "File Not Found":
		return color.RGBA{80, 80, 80, 255}
	default:
		return color.RGBA{180, 40, 40, 255}
	}
}

// drawCentredText draws white text centred in the image.
// For small images the text may be clipped, which is acceptable for thumbnails.
func drawCentredText(img *image.RGBA, text string, size int) {
	face := basicfont.Face7x13
	textWidth := font.MeasureString(face, text).Ceil()
	x := (size - textWidth) / 2
	x = max(x, 2)
	y := size/2 + face.Metrics().Ascent.Ceil()/2

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.White),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(text)
}

// GenerateOrPlaceholder wraps Generate: on success it returns the real
// thumbnail; on any error it returns a placeholder image indicating the
// error type. It never returns nil.
func GenerateOrPlaceholder(filePath string, height uint) image.Image {
	img, err := Generate(filePath, height)
	if err == nil {
		return img
	}
	info := classifyError(err)
	return ErrorPlaceholder(info.label, height)
}
