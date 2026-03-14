# go-thumbnails

Pure Go thumbnail generator for PDF, TIFF, JPEG, PNG, and GIF documents. Uses PDFium via WebAssembly — no CGo required.

## Features

- Multi-page composite thumbnails (up to 4 pages side-by-side with "+" indicator)
- Uniform fixed-size thumbnails with page-count badge
- Per-page thumbnail extraction via page-level API
- Error placeholder generation with colour-coded labels
- PDF rendering corruption detection

## Installation

### Library

```
go get github.com/drummonds/go-thumbnails
```

### CLI

```
go install github.com/drummonds/go-thumbnails/cmd/thumbnails@latest
```

## Usage

### Library

```go
import thumbnails "github.com/drummonds/go-thumbnails"

// Composite thumbnail (default)
img, err := thumbnails.Generate("doc.pdf", 128)

// Uniform style with page-count badge
img, err := thumbnails.GenerateStyled("doc.pdf", 128, thumbnails.StyleUniform)

// Generate and save to disk
err := thumbnails.GenerateAndSave("doc.pdf", "doc.tn_128.png", 128)

// Render individual pages
pages, err := thumbnails.RenderPages("doc.pdf")
for _, p := range pages {
    resized := thumbnails.ResizePage(p.Image, 128)
    // p.PageNum, p.PageCount available
}

// Render a single page
page, err := thumbnails.RenderPage("doc.pdf", 3)
```

### CLI

```
thumbnails [flags] file [file...]

Flags:
  -w int        Thumbnail width (default 128)
  -o string     Output path or directory (default: alongside input)
  -style string Rendering style: composite, uniform, or page (default "composite")
  -page int     Page number for -style page (0 = all pages)
```

Examples:

```sh
# Composite thumbnail at default width
thumbnails doc.pdf

# Uniform style, 256px wide
thumbnails -w 256 -style uniform doc.pdf

# Extract all pages as individual thumbnails
thumbnails -style page doc.pdf

# Extract page 3 only
thumbnails -style page -page 3 doc.pdf

# Process multiple files
thumbnails -w 64 *.pdf
```

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
| Documentation | https://h3-go-thumbnails.statichost.page/ |
| Source (Codeberg) | https://codeberg.org/hum3/go-thumbnails |
| Mirror (GitHub) | https://github.com/drummonds/go-thumbnails |
