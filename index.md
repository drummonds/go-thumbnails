# go-thumbnails

Pure Go thumbnail generator for PDF, TIFF, JPEG, PNG, and GIF documents. Uses PDFium via WebAssembly — no CGo required.

## Features

- Multi-page composite thumbnails (up to 4 pages + "+" indicator)
- Uniform fixed-size thumbnails with page-count badge
- Per-page thumbnail extraction
- Error placeholder generation with colour-coded labels
- PDF corruption detection (PDFium WASM alpha artefacts)

## Supported formats

| Format | Multi-page | Notes |
|--------|-----------|-------|
| PDF    | Yes       | Via PDFium WebAssembly |
| TIFF   | First page only | golang.org/x/image/tiff lacks DecodeAll |
| JPEG   | No        | Simple resize |
| PNG    | No        | Simple resize |
| GIF    | No        | Simple resize |

## Links

| | |
|---|---|
| Source (Codeberg) | https://codeberg.org/hum3/go-thumbnails |
| Mirror (GitHub) | https://github.com/drummonds/go-thumbnails |
