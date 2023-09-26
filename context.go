package renderer

import (
	"context"
	"errors"
)

type key string

const (
	rendererKey key = "renderContext"
)

var (
	ErrRendererContextNotFound = errors.New("renderer context not found")
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
