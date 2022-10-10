package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/liuminhaw/renderer"
)

const url = "https://lmhaw.dev"

func main() {
	headless := flag.Bool("headless", true, "automation browser execution mode")
	browserWidth := flag.Int("bWidth", 1920, "width of browser window's size")
	browserHeight := flag.Int("bHeight", 1080, "height of browser window's size")
	timeout := flag.Int("timeout", 30, "seconds before timeout when rendering")
	imageLoad := flag.Bool("imageLoad", false, "indicate if load image when rendering")
	flag.Parse()

	if *browserWidth <= 0 || *browserHeight <= 0 {
		fmt.Println("Browser width / height value should be greater than 0")
		os.Exit(1)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "headless", *headless)
	ctx = context.WithValue(ctx, "windowWidth", *browserWidth)
	ctx = context.WithValue(ctx, "windowHeight", *browserHeight)
	ctx = context.WithValue(ctx, "timeout", *timeout)
	ctx = context.WithValue(ctx, "imageLoad", *imageLoad)

	ret, err := renderer.RenderPage(ctx, url)
	if err != nil {
		log.Fatalf("Render test: %s", err)
	}

	fmt.Println(string(ret))
}
