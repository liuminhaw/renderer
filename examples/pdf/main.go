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
	landscape := flag.Bool("landscape", false, "create pdf in landscape layout")
	headerFooter := flag.Bool("headerFooter", false, "show header and footer")
	paperWidth := flag.Float64("paperWidth", 0, "paper width in centimeter")
	paperHeight := flag.Float64("paperHeight", 0, "paper height in centimeter")
	marginTop := flag.Float64("marginTop", 1, "top margin in centimeter")
	marginBottom := flag.Float64("marginBottom", 1, "bottom margin in centimeter")
	marginLeft := flag.Float64("marginLeft", 1, "left margin in centimeter")
	marginRight := flag.Float64("marginRight", 1, "right margin in centimeter")
	idleType := flag.String("idleType", "networkIdle",
		"how to determine loading idle and return, valid input: networkIdle, InteractiveTime")
	flag.Parse()

	if *paperWidth < 0 || *paperHeight < 0 {
		fmt.Println("Paper width / height value should be greater than 0")
		os.Exit(1)
	}
	if *marginTop < 0 || *marginBottom < 0 || *marginLeft < 0 || *marginRight < 0 {
		fmt.Println("Margins value should be greater than 0")
		os.Exit(1)
	}
	if *idleType != "networkIdle" && *idleType != "InteractiveTime" {
		fmt.Println("Valid idleType value: networkIdle, InteractiveTime")
		os.Exit(1)
	}
	if len(flag.Args()) != 1 {
		fmt.Printf("Usage: %s url\n", os.Args[0])
		os.Exit(1)
	}
	url := flag.Arg(0)

	pdfContext := renderer.PdfContext{
		Landscape:           *landscape,
		DisplayHeaderFooter: *headerFooter,
		PaperWidthCm:        *paperWidth,
		PaperHeightCm:       *paperHeight,
		MarginTopCm:         *marginTop,
		MarginBottomCm:      *marginBottom,
		MarginLeftCm:        *marginLeft,
		MarginRightCm:       *marginRight,
		IdleType:            *idleType,
	}

	ctx := context.Background()
	ctx = renderer.WithPdfContext(ctx, &pdfContext)

	context, err := renderer.RenderPdf(ctx, url)
	if err != nil {
		log.Fatalf("RenderPDf test: %s", err)
	}

	if err := os.MkdirAll("result", 0775); err != nil {
		log.Fatalf("RenderPDF test: %s", err)
	}
	f, err := os.Create("result/result.pdf")
	if err != nil {
		log.Fatalf("RenderPDF test: %s", err)
	}
	defer f.Close()

	f.Write(context)
}
