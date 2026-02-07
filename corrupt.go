package thumbnails

import (
	"image"
)

// CorruptionResult describes corruption detected in a rendered page.
type CorruptionResult struct {
	// Corrupt is true if the image appears corrupted.
	Corrupt bool
	// Reason describes the type of corruption detected.
	Reason string
	// CorruptRowFraction is the fraction of rows with non-grayscale artifacts (0.0–1.0).
	CorruptRowFraction float64
	// NonOpaqueRowFraction is the fraction of rows containing non-opaque (alpha != 255) pixels.
	NonOpaqueRowFraction float64
}

// CheckPageCorruption detects rendering corruption in a single page image.
//
// PDFium can produce corrupt RGBA buffers where pixel data contains garbage bytes
// with non-255 alpha values and spurious colour in what should be grayscale or
// clean colour content. The key signal is rows where a high fraction of pixels
// have alpha != 255 — legitimate document renders are fully opaque.
func CheckPageCorruption(img image.Image) CorruptionResult {
	rgba, ok := img.(*image.RGBA)
	if !ok {
		// Non-RGBA images: fall back to generic check
		return checkGenericCorruption(img)
	}

	b := rgba.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == 0 || h == 0 {
		return CorruptionResult{Corrupt: true, Reason: "zero dimensions"}
	}

	corruptRows := 0
	alphaRows := 0

	// Sample every pixel in each row but skip rows for large images
	rowStep := 1
	if h > 500 {
		rowStep = h / 500
	}
	rowsSampled := 0

	for y := b.Min.Y; y < b.Max.Y; y += rowStep {
		rowsSampled++
		rowStart := (y - b.Min.Y) * rgba.Stride
		nonOpaqueInRow := 0
		pixelsSampled := 0

		// Sample pixels across the row
		xStep := 1
		if w > 100 {
			xStep = w / 100
		}
		for x := 0; x < w; x += xStep {
			off := rowStart + x*4
			a := rgba.Pix[off+3]
			if a != 255 {
				nonOpaqueInRow++
			}
			pixelsSampled++
		}

		// A row is "alpha-corrupt" if >10% of sampled pixels are non-opaque
		if pixelsSampled > 0 && float64(nonOpaqueInRow)/float64(pixelsSampled) > 0.10 {
			alphaRows++
			corruptRows++
		}
	}

	if rowsSampled == 0 {
		return CorruptionResult{}
	}

	corruptFrac := float64(corruptRows) / float64(rowsSampled)
	alphaFrac := float64(alphaRows) / float64(rowsSampled)

	// If >5% of sampled rows are corrupt, flag the page
	if corruptFrac > 0.05 {
		return CorruptionResult{
			Corrupt:              true,
			Reason:               "non-opaque alpha rows indicating corrupt pixel buffer",
			CorruptRowFraction:   corruptFrac,
			NonOpaqueRowFraction: alphaFrac,
		}
	}

	return CorruptionResult{
		CorruptRowFraction:   corruptFrac,
		NonOpaqueRowFraction: alphaFrac,
	}
}

// CheckThumbnailCorruption checks a final composited thumbnail for corruption.
// This catches issues that survive resizing/compositing.
func CheckThumbnailCorruption(img image.Image) CorruptionResult {
	return CheckPageCorruption(img)
}

func checkGenericCorruption(img image.Image) CorruptionResult {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == 0 || h == 0 {
		return CorruptionResult{Corrupt: true, Reason: "zero dimensions"}
	}

	nonOpaqueRows := 0
	rowsSampled := 0
	rowStep := 1
	if h > 500 {
		rowStep = h / 500
	}

	for y := b.Min.Y; y < b.Max.Y; y += rowStep {
		rowsSampled++
		nonOpaque := 0
		sampled := 0
		xStep := 1
		if w > 100 {
			xStep = w / 100
		}
		for x := b.Min.X; x < b.Max.X; x += xStep {
			_, _, _, a := img.At(x, y).RGBA()
			if a != 0xffff {
				nonOpaque++
			}
			sampled++
		}
		if sampled > 0 && float64(nonOpaque)/float64(sampled) > 0.10 {
			nonOpaqueRows++
		}
	}

	if rowsSampled == 0 {
		return CorruptionResult{}
	}

	frac := float64(nonOpaqueRows) / float64(rowsSampled)
	if frac > 0.05 {
		return CorruptionResult{
			Corrupt:              true,
			Reason:               "non-opaque alpha rows indicating corrupt pixel buffer",
			NonOpaqueRowFraction: frac,
			CorruptRowFraction:   frac,
		}
	}

	return CorruptionResult{
		NonOpaqueRowFraction: frac,
		CorruptRowFraction:   frac,
	}
}
