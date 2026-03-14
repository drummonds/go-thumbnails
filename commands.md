# Command-line Tools

go-thumbnails includes several command-line tools for development, testing, and batch processing.

## showcase

Visual validation tool that generates sample thumbnails across all formats and styles.

```sh
go run ./cmd/showcase [-width 64] [-output /tmp/thumbnails]
```

Processes testdata files (if present) and synthetic images. Generates composite and uniform style thumbnails, plus error placeholders. Useful for visually checking rendering after code changes.

## batch

Batch thumbnail generation with corruption detection and JSON reporting.

```sh
go run ./cmd/batch -input pdfs/ -output thumbs/ [-width 64] [-report report.json]
```

Flags:
- `-input` — directory of PDF files to process
- `-output` — directory for generated thumbnails
- `-width` — thumbnail width (default 64)
- `-report` — JSON report output path

The JSON report includes per-file status (ok/error/corrupt), dimensions, elapsed time, and corruption fractions from `CheckThumbnailCorruption`.

## diagnose

Low-level PDF rendering diagnostics for investigating corruption.

```sh
go run ./cmd/diagnose <pdf-path> <output-dir>
```

Outputs raw page PNGs, alpha corruption analysis, row-level statistics, and pixel dumps. Used for debugging PDFium WebAssembly rendering issues.

## gentestimages

Creates synthetic test images for the test suite.

```sh
go run ./cmd/gentestimages [-output testdata]
```

Generates 8 test images:
- PNG: landscape (300x200), portrait (200x400), square (256x256), tiny (8x8), large (1200x800)
- JPG: photo (400x300), portrait (300x500), small (50x50)

## gendocimages

Generates gallery images for documentation.

```sh
go run ./cmd/gendocimages [-output docs/images]
```

Creates thumbnail examples for the [gallery](gallery.md) page. Generates synthetic image thumbnails unconditionally; PDF examples require `testdata/` to be present.
