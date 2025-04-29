package renderer_test

import (
	"log/slog"
	"os"

	"github.com/liuminhaw/renderer"
)

func ExampleRenderer_RenderPage() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	r := renderer.NewRenderer(renderer.WithLogger(logger))
	content, err := r.RenderPage("https://www.example.com", nil)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	f, err := os.Create("result.html")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer f.Close()

	f.Write(content)
}

func ExampleRenderer_RenderPdf() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	r := renderer.NewRenderer(renderer.WithLogger(logger))

	content, err := r.RenderPdf("https://www.example.com", nil)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	f, err := os.Create("result.pdf")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer f.Close()

	f.Write(content)
}
