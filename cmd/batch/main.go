package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"time"

	thumbnails "github.com/drummonds/go-thumbnails"
)

type Result struct {
	File            string  `json:"file"`
	Status          string  `json:"status"` // "ok", "error", "corrupt"
	Error           string  `json:"error,omitempty"`
	Width           int     `json:"width,omitempty"`
	Height          int     `json:"height,omitempty"`
	Elapsed         float64 `json:"elapsed_ms"`
	FileSize        int64   `json:"file_size_bytes,omitempty"`
	OutPath         string  `json:"out_path,omitempty"`
	CorruptRowPct   float64 `json:"corrupt_row_pct,omitempty"`
	NonOpaqueRowPct float64 `json:"non_opaque_row_pct,omitempty"`
}

func main() {
	inputDir := flag.String("input", "", "Directory containing PDF files")
	outputDir := flag.String("output", "", "Directory for thumbnail output")
	width := flag.Uint("width", 64, "Thumbnail width in pixels")
	reportPath := flag.String("report", "", "Path for JSON report (default: stdout)")
	flag.Parse()

	if *inputDir == "" || *outputDir == "" {
		fmt.Fprintf(os.Stderr, "Usage: batch -input <dir> -output <dir> [-width N] [-report file.json]\n")
		os.Exit(1)
	}

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output dir: %v\n", err)
		os.Exit(1)
	}

	pdfs, err := filepath.Glob(filepath.Join(*inputDir, "*.pdf"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to glob PDFs: %v\n", err)
		os.Exit(1)
	}
	sort.Strings(pdfs)

	fmt.Fprintf(os.Stderr, "Processing %d PDFs from %s\n", len(pdfs), *inputDir)
	fmt.Fprintf(os.Stderr, "Output to %s, width=%d\n\n", *outputDir, *width)

	var results []Result
	okCount, errCount, corruptCount := 0, 0, 0

	for i, pdfPath := range pdfs {
		baseName := filepath.Base(pdfPath)
		outName := baseName[:len(baseName)-len(".pdf")] + ".tn.png"
		outPath := filepath.Join(*outputDir, outName)

		info, _ := os.Stat(pdfPath)
		var fileSize int64
		if info != nil {
			fileSize = info.Size()
		}

		start := time.Now()
		img, genErr := thumbnails.Generate(pdfPath, *width)
		elapsed := time.Since(start).Milliseconds()

		r := Result{
			File:     baseName,
			Elapsed:  float64(elapsed),
			FileSize: fileSize,
		}

		if genErr != nil {
			r.Status = "error"
			r.Error = genErr.Error()
			errCount++
			fmt.Fprintf(os.Stderr, "[%3d/%d] ERROR   %s: %v (%.0fms)\n", i+1, len(pdfs), baseName, genErr, r.Elapsed)
		} else {
			bounds := img.Bounds()
			r.Width = bounds.Dx()
			r.Height = bounds.Dy()
			r.OutPath = outName

			cr := thumbnails.CheckThumbnailCorruption(img)
			r.CorruptRowPct = cr.CorruptRowFraction * 100
			r.NonOpaqueRowPct = cr.NonOpaqueRowFraction * 100

			if cr.Corrupt {
				r.Status = "corrupt"
				r.Error = fmt.Sprintf("%s (%.1f%% corrupt rows)", cr.Reason, r.CorruptRowPct)
				corruptCount++
				fmt.Fprintf(os.Stderr, "[%3d/%d] CORRUPT %s: %.1f%% corrupt rows (%dx%d, %.0fms)\n",
					i+1, len(pdfs), baseName, r.CorruptRowPct, r.Width, r.Height, r.Elapsed)
			} else {
				r.Status = "ok"
				okCount++
				fmt.Fprintf(os.Stderr, "[%3d/%d] OK      %s (%dx%d, %.0fms)\n",
					i+1, len(pdfs), baseName, r.Width, r.Height, r.Elapsed)
			}

			// Save the thumbnail regardless (so we can eyeball corrupt ones)
			if saveErr := savePNG(img, outPath); saveErr != nil {
				fmt.Fprintf(os.Stderr, "  WARNING: failed to save %s: %v\n", outPath, saveErr)
			}
		}

		results = append(results, r)
	}

	fmt.Fprintf(os.Stderr, "\n=== Summary ===\n")
	fmt.Fprintf(os.Stderr, "Total: %d  OK: %d  Error: %d  Corrupt: %d\n", len(pdfs), okCount, errCount, corruptCount)

	// Output JSON report
	reportData, _ := json.MarshalIndent(results, "", "  ")
	if *reportPath != "" {
		if err := os.WriteFile(*reportPath, reportData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write report: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Report written to %s\n", *reportPath)
		}
	} else {
		fmt.Println(string(reportData))
	}
}

func savePNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
