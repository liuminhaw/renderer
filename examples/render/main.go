package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/liuminhaw/renderer"
)

func main() {
	headless := flag.Bool("headless", true, "automation browser execution mode")
	browserWidth := flag.Int("bWidth", 1920, "width of browser window's size")
	browserHeight := flag.Int("bHeight", 1080, "height of browser window's size")
	timeout := flag.Int("timeout", 30, "seconds before timeout when rendering")
	imageLoad := flag.Bool("imageLoad", false, "indicate if load image when rendering")
	idleType := flag.String("idleType", "auto",
		"how to determine loading idle and return, valid input: auto, networkIdle, InteractiveTime")
	skipFrameCount := flag.Int("skipFrameCount", 0,
		"skip first n frames with same id as init frame, only valid with idleType=networkIdle")
	browserExecPath := flag.String("browserPath", "", "manually set browser executable path")
	container := flag.Bool(
		"container",
		false,
		"indicate if running in container (docker / lambda) environment",
	)
	debug := flag.Bool("debug", false, "turn on for outputing debug message")
	chromiumDebug := flag.Bool("chromiumDebug", false, "turn on for chromium debug message output")
	userAgent := flag.String(
		"userAgent",
		"",
		"set custom user agent for sending request in automation browser",
	)

	flag.Parse()

	if *browserWidth <= 0 || *browserHeight <= 0 {
		fmt.Println("Browser width / height value should be greater than 0")
		os.Exit(1)
	}
	if !renderer.IsValidIdleType(*idleType) {
		fmt.Println("Valid idleType value: auto, networkIdle, InteractiveTime")
		os.Exit(1)
	}
	if *skipFrameCount < 0 {
		fmt.Println("skipFrameCount should be greater than or equal to 0")
		os.Exit(1)
	}
	flag.Parse()
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
	context, err := r.RenderPage(url, &renderer.RendererOption{
		BrowserOpts: renderer.BrowserConf{
			IdleType:        *idleType,
			BrowserExecPath: *browserExecPath,
			Container:       *container,
			ChromiumDebug:   *chromiumDebug,
			DebugMode:       *debug,
		},
		Headless:       *headless,
		WindowWidth:    *browserWidth,
		WindowHeight:   *browserHeight,
		Timeout:        *timeout,
		ImageLoad:      *imageLoad,
		SkipFrameCount: *skipFrameCount,
		UserAgent:      *userAgent,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Render test: %s", err))
		os.Exit(1)
	}

	if err := os.MkdirAll("result", 0775); err != nil {
		logger.Error(fmt.Sprintf("Render test: %s", err))
		os.Exit(1)
	}
	f, err := os.Create("result/result.out")
	if err != nil {
		logger.Error(fmt.Sprintf("Render test: %s", err))
		os.Exit(1)
	}
	defer f.Close()

	f.Write(context)
}
