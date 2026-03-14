# Changelog

## [Unreleased]

### Added
- Page-level API: `RenderPages`, `RenderPage`, `ResizePage`, `DefaultPageThumbnailPath`
- `PageResult` type with page metadata (page number, page count)
- `ErrPageOutOfRange` sentinel error
- CLI tool (`cmd/thumbnails`) for generating thumbnails from the command line
- `ROADMAP.md`
- Documentation build pipeline (`docs:build` task, statichost deployment)

## [0.6.2] - 2026-03-02

- Adding gif to thumbnail generation

## [0.6.0]

- Change thumbnail background from white to light grey (RGB 240,240,240) so padding is visible when images are fitted into portrait thumbnails

## [0.5.0]

- Add rendering style choices (composite vs uniform page layout)

## [0.4.1]

- Fix width of icons and 1.4 height ratio

## [0.4.0]

- Add `GenerateOrPlaceholder` and `ErrorPlaceholder` for error placeholder thumbnails

## [0.3.0]

- Add compiled binaries to .gitignore
- Fix PDFium WASM alpha corruption and replace `nfnt/resize` with `x/image/draw`

## [0.1.0]

- Extract thumbnail generation into standalone `go-thumbnails` module
