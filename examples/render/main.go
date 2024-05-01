package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/liuminhaw/renderer"
)

func main() {
	headless := flag.Bool("headless", true, "automation browser execution mode")
	browserWidth := flag.Int("bWidth", 1920, "width of browser window's size")
	browserHeight := flag.Int("bHeight", 1080, "height of browser window's size")
	timeout := flag.Int("timeout", 30, "seconds before timeout when rendering")
	imageLoad := flag.Bool("imageLoad", false, "indicate if load image when rendering")
	idleType := flag.String("idleType", "networkIdle",
		"how to determine loading idle and return, valid input: networkIdle, InteractiveTime")
	skipFrameCount := flag.Int("skipFrameCount", 0,
		"skip first n frames with same id as init frame, only valid with idleType=networkIdle")
	browserExecPath := flag.String("browserPath", "", "manually set browser executable path")
	container := flag.Bool(
		"container",
		false,
		"indicate if running in container (docker / lambda) environment",
	)
	debug := flag.Bool("debug", false, "turn on for outputing debug message")

	flag.Parse()

	if *browserWidth <= 0 || *browserHeight <= 0 {
		fmt.Println("Browser width / height value should be greater than 0")
		os.Exit(1)
	}
	if *idleType != "networkIdle" && *idleType != "InteractiveTime" {
		fmt.Println("Valid idleType value: networkIdle, InteractiveTime")
		os.Exit(1)
	}
	if *skipFrameCount < 0 {
		fmt.Println("skipFrameCount should be greater than or equal to 0")
		os.Exit(1)
	}
	if len(flag.Args()) != 1 {
		fmt.Printf("Usage: %s url\n", os.Args[0])
		os.Exit(1)
	}
	url := flag.Arg(0)

	// Explicit set browserContext if need to modify settings
	// otherwise no need to additional set it up
	browserContext := renderer.BrowserContext{
		IdleType:        *idleType,
		BrowserExecPath: *browserExecPath,
		Container:       *container,
		DebugMode:       *debug,
	}

	rendererContext := renderer.RendererContext{
		Headless:       *headless,
		WindowWidth:    *browserWidth,
		WindowHeight:   *browserHeight,
		Timeout:        *timeout,
		ImageLoad:      *imageLoad,
		SkipFrameCount: *skipFrameCount,
	}

	ctx := context.Background()
	ctx = renderer.WithBrowserContext(ctx, &browserContext)
	ctx = renderer.WithRendererContext(ctx, &rendererContext)

	context, err := renderer.RenderPage(ctx, url)
	if err != nil {
		log.Fatalf("Render test: %s", err)
	}

	if err := os.MkdirAll("result", 0775); err != nil {
		log.Fatalf("Render test: %s", err)
	}
	f, err := os.Create("result/result.out")
	if err != nil {
		log.Fatalf("Render test: %s", err)
	}
	defer f.Close()

	f.Write(context)
}
