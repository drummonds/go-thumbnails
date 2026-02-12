package thumbnails

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// generateImageThumbnail creates a thumbnail from a JPG or PNG file.
// It resizes to the target width and crops/pads to A4 aspect ratio.
func generateImageThumbnail(path string, width uint) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return resizeToPage(img, width), nil
}
