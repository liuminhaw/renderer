package renderer

import (
	"context"
	"errors"
)

type key string

const (
	rendererKey key = "renderContext"
	pdfKey      key = "pdfContext"
)

var (
	ErrRendererContextNotFound = errors.New("renderer context not found")
	ErrPdfContextNotFound      = errors.New("pdf context not found")
)

type RendererContext struct {
	Headless       bool
	WindowWidth    int
	WindowHeight   int
	Timeout        int
	ImageLoad      bool
	IdleType       string
	SkipFrameCount int
}

func WithRendererContext(ctx context.Context, rc *RendererContext) context.Context {
	return context.WithValue(ctx, rendererKey, rc)
}

func GetRendererContext(ctx context.Context) (*RendererContext, error) {
	val := ctx.Value(rendererKey)

	rendererContext, ok := val.(*RendererContext)
	if !ok {
		return nil, ErrRendererContextNotFound
	}

	return rendererContext, nil
}

type PdfContext struct {
	Landscape           bool
	DisplayHeaderFooter bool
	PaperWidthCm        float64
	PaperHeightCm       float64
	MarginTopCm         float64
	MarginBottomCm      float64
	MarginLeftCm        float64
	MarginRightCm       float64
	IdleType            string
}

func WithPdfContext(ctx context.Context, pc *PdfContext) context.Context {
	return context.WithValue(ctx, pdfKey, pc)
}

func GetPdfContext(ctx context.Context) (*PdfContext, error) {
	val := ctx.Value(pdfKey)

	pdfContext, ok := val.(*PdfContext)
	if !ok {
		return nil, ErrPdfContextNotFound
	}

	return pdfContext, nil
}
