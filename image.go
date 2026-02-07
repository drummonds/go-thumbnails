package thumbnails

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/nfnt/resize"
)

// generateImageThumbnail creates a thumbnail from a JPG or PNG file.
// It simply resizes the image to the target height, maintaining aspect ratio.
func generateImageThumbnail(path string, height uint) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	resized := resize.Resize(0, height, img, resize.Lanczos3)
	return resized, nil
}
