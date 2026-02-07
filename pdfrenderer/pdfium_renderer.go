package pdfrenderer

import (
	"fmt"
	"image"
	"os"
	"time"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"
)

// PDFiumRenderer implements PDF rendering using go-pdfium with WebAssembly (pure Go, no CGo).
type PDFiumRenderer struct {
	pool     pdfium.Pool
	instance pdfium.Pdfium
}

// NewPDFiumRenderer creates a new PDFium-based PDF renderer using WebAssembly.
func NewPDFiumRenderer() (*PDFiumRenderer, error) {
	pool, err := webassembly.Init(webassembly.Config{
		MinIdle:  1,
		MaxIdle:  1,
		MaxTotal: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PDFium WebAssembly: %w", err)
	}

	instance, err := pool.GetInstance(time.Second * 30)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to get PDFium instance: %w", err)
	}

	return &PDFiumRenderer{
		pool:     pool,
		instance: instance,
	}, nil
}

// RenderPDF converts all pages of a PDF file to images using go-pdfium WebAssembly.
func (r *PDFiumRenderer) RenderPDF(filename string) ([]image.Image, error) {
	pdfBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read PDF file: %w", err)
	}

	doc, err := r.instance.OpenDocument(&requests.OpenDocument{
		File: &pdfBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to open PDF document: %w", err)
	}
	defer r.instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{
		Document: doc.Document,
	})

	pageCountResp, err := r.instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{
		Document: doc.Document,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get page count: %w", err)
	}

	numPages := pageCountResp.PageCount
	images := make([]image.Image, 0, numPages)

	for pageIndex := 0; pageIndex < numPages; pageIndex++ {
		pageRender, err := r.instance.RenderPageInDPI(&requests.RenderPageInDPI{
			DPI: 150,
			Page: requests.Page{
				ByIndex: &requests.PageByIndex{
					Document: doc.Document,
					Index:    pageIndex,
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("unable to render page %d: %w", pageIndex, err)
		}

		// Copy pixel data into a Go-owned buffer and fix corrupt alpha.
		// PDFium WASM produces RGBA buffers with garbage alpha channels,
		// and Cleanup() may invalidate the underlying WASM memory.
		src := pageRender.Result.Image
		pix := make([]byte, len(src.Pix))
		copy(pix, src.Pix)
		for i := 3; i < len(pix); i += 4 {
			pix[i] = 255
		}
		img := &image.RGBA{
			Pix:    pix,
			Stride: src.Stride,
			Rect:   src.Rect,
		}
		pageRender.Cleanup()

		images = append(images, img)
	}

	return images, nil
}

// Close cleans up resources used by the PDFium renderer.
func (r *PDFiumRenderer) Close() error {
	if r.pool != nil {
		r.pool.Close()
		r.pool = nil
	}
	r.instance = nil
	return nil
}
