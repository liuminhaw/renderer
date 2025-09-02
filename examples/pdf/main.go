package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/liuminhaw/renderer"
)

func main() {
	landscape := flag.Bool("landscape", false, "create pdf in landscape layout")
	headerFooter := flag.Bool("headerFooter", false, "show header and footer")
	paperWidth := flag.Float64("paperWidth", 0, "paper width in centimeter")
	paperHeight := flag.Float64("paperHeight", 0, "paper height in centimeter")
	marginTop := flag.Float64("marginTop", 1, "top margin in centimeter")
	marginBottom := flag.Float64("marginBottom", 1, "bottom margin in centimeter")
	marginLeft := flag.Float64("marginLeft", 1, "left margin in centimeter")
	marginRight := flag.Float64("marginRight", 1, "right margin in centimeter")
	idleType := flag.String("idleType", "auto",
		"how to determine loading idle and return, valid input: auto, networkIdle, InteractiveTime")
	browserExecPath := flag.String("browserPath", "", "manually set browser executable path")
	container := flag.Bool(
		"container",
		false,
		"indicate if running in container (docker / lambda) environment",
	)
	debug := flag.Bool("debug", false, "turn on for outputing debug message")
	chromiumDebug := flag.Bool(
		"chromiumDebug",
		false,
		"turn on for chrome debug message output (must enable debug for output)",
	)

	flag.Parse()

	if *paperWidth < 0 || *paperHeight < 0 {
		fmt.Println("Paper width / height value should be greater than 0")
		os.Exit(1)
	}
	if *marginTop < 0 || *marginBottom < 0 || *marginLeft < 0 || *marginRight < 0 {
		fmt.Println("Margins value should be greater than 0")
		os.Exit(1)
	}
	if !renderer.IsValidIdleType(*idleType) {
		fmt.Println("Valid idleType value: networkIdle, InteractiveTime")
		os.Exit(1)
	}
	if len(flag.Args()) != 1 {
		fmt.Printf("Usage: %s url\n", os.Args[0])
		os.Exit(1)
	}
	url := flag.Arg(0)

	var logger *slog.Logger
	if *debug {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	r := renderer.NewRenderer(renderer.WithLogger(logger))

	context, err := r.RenderPdf(url, &renderer.PdfOption{
		BrowserOpts: renderer.BrowserConf{
			IdleType:        *idleType,
			BrowserExecPath: *browserExecPath,
			Container:       *container,
			ChromiumDebug:   *chromiumDebug,
			DebugMode:       *debug,
		},
		RendererOpts:        renderer.DefaultPdfOption.RendererOpts,
		Landscape:           *landscape,
		DisplayHeaderFooter: *headerFooter,
		PaperWidthCm:        *paperWidth,
		PaperHeightCm:       *paperHeight,
		MarginTopCm:         *marginTop,
		MarginBottomCm:      *marginBottom,
		MarginLeftCm:        *marginLeft,
		MarginRightCm:       *marginRight,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("RenderPDF test: %s", err))
		os.Exit(1)
	}

	if err := os.MkdirAll("result", 0775); err != nil {
		logger.Error(fmt.Sprintf("RenderPDF test: %s", err))
		os.Exit(1)
	}
	f, err := os.Create("result/result.pdf")
	if err != nil {
		logger.Error(fmt.Sprintf("RenderPDF test: %s", err))
		os.Exit(1)
	}
	defer f.Close()

	f.Write(context)
}
