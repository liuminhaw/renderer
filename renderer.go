package renderer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

const (
	defaultIdleType string = "networkIdle"
	defaultTimeout  int    = 30
)

type networkIdle struct {
	navigateFrame  bool
	frameId        string
	skipFrameCount int
	frameCount     int
}

// RenderPdf turns given url page into pdf and return result
func RenderPdf(ctx context.Context, urlStr string) ([]byte, error) {
	idleType := defaultIdleType
	pdfParams := page.PrintToPDF()

	browserContext, err := GetBrowserContext(ctx)
	if err != nil {
		if errors.Is(err, ErrBrowserContextNotFound) {
			browserContext = &BrowserContext{}
			ctx = WithBrowserContext(ctx, browserContext)
		} else {
			return nil, fmt.Errorf("render pdf; %w", err)
		}
	}

	pdfContext, _ := GetPdfContext(ctx)
	if pdfContext != nil {
		pdfParams = setPdfParams(pdfContext)
		switch browserContext.IdleType {
		case "":
			idleType = defaultIdleType
		case "networkIdle", "InteractiveTime":
			idleType = browserContext.IdleType
		default:
			return nil, fmt.Errorf("render pdf: invalid idleType %s", browserContext.IdleType)
		}
	}

	opts := chromedp.DefaultExecAllocatorOptions[:]
	if browserContext.BrowserExecPath != "" {
		opts = append(opts, chromedp.ExecPath(browserContext.BrowserExecPath))
	}

	if browserContext.Container {
		debugMessage(browserContext.DebugMode, "Set configuration for container environment")
		opts = append(opts,
			chromedp.Flag("disable-setuid-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("single-process", true),
			chromedp.Flag("no-zygote", true),
			chromedp.NoSandbox,
		)
	}

	start := time.Now()
	ctx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()
	if browserContext.ChromiumDebug {
		ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(log.Printf))
	} else {
		ctx, cancel = chromedp.NewContext(ctx)
	}
	defer cancel()

	var res []byte
	err = chromedp.Run(ctx,
		navigateAndWaitFor(urlStr, idleType),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := pdfParams.Do(ctx)
			if err != nil {
				return fmt.Errorf("renderPdf(%v): %w", urlStr, err)
			}
			res = buf
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp run: %w", err)
	}

	duration := time.Since(start)
	debugMessage(browserContext.DebugMode, fmt.Sprintf("Render time: %v", duration))
	return res, nil
}

// RenderPage rendered given url in browser and returns result html content
func RenderPage(ctx context.Context, urlStr string) ([]byte, error) {
	browserContext, err := GetBrowserContext(ctx)
	if err != nil {
		if errors.Is(err, ErrBrowserContextNotFound) {
			browserContext = &BrowserContext{}
			ctx = WithBrowserContext(ctx, browserContext)
		} else {
			return nil, fmt.Errorf("render page; %w", err)
		}
	}

	rendererContext, err := GetRendererContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("render page: %w", err)
	}

	var idleType string
	switch browserContext.IdleType {
	case "":
		idleType = "networkIdle"
	case "networkIdle", "InteractiveTime":
		idleType = browserContext.IdleType
	default:
		return nil, fmt.Errorf("render page: invalid idleType %s", browserContext.IdleType)
	}

	var windowWidth, windowHeight int = 1000, 1000
	if rendererContext.WindowWidth != 0 {
		windowWidth = rendererContext.WindowWidth
	}
	if rendererContext.WindowHeight != 0 {
		windowHeight = rendererContext.WindowHeight
	}

	var headless bool = rendererContext.Headless
	opts := chromedp.DefaultExecAllocatorOptions[:]
	opts = append(
		opts,
		chromedp.Flag("blink-settings", fmt.Sprintf("imagesEnbled=%v", rendererContext.ImageLoad)),
	)
	opts = append(opts, chromedp.WindowSize(windowWidth, windowHeight))

	if browserContext.BrowserExecPath != "" {
		opts = append(opts, chromedp.ExecPath(browserContext.BrowserExecPath))
	}

	if browserContext.Container {
		debugMessage(browserContext.DebugMode, "Set configuration for container environment")
		opts = append(opts,
			chromedp.Flag("disable-setuid-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("single-process", true),
			chromedp.Flag("no-zygote", true),
			chromedp.NoSandbox,
		)
	}

	opts = append(opts, chromedp.Flag("headless", headless))

	start := time.Now()
	ctx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()
	if browserContext.ChromiumDebug {
		ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(log.Printf))
	} else {
		ctx, cancel = chromedp.NewContext(ctx)
	}
	defer cancel()

	var res string
	err = chromedp.Run(ctx,
		chromedp.Tasks{
			navigateAndWaitFor(urlStr, idleType),
			chromedp.ActionFunc(func(ctx context.Context) error {
				node, err := dom.GetDocument().Do(ctx)
				if err != nil {
					fmt.Printf("renderPage(%v): %v", urlStr, err)
					return fmt.Errorf("renderPage(%v): %w", urlStr, err)
				}
				res, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
				if err != nil {
					fmt.Printf("renderPage(%v): %v", urlStr, err)
					return fmt.Errorf("renderPage(%v): %w", urlStr, err)
				}
				return nil
			}),
		},
	)
	if err != nil {
		fmt.Printf("chromedp run error: %s\n", err)
	}

	duration := time.Since(start)
	debugMessage(browserContext.DebugMode, fmt.Sprintf("Render time: %v", duration))
	return []byte(res), nil
}

// navigateAndWaitFor is defined as task of chromedp for rendering step
func navigateAndWaitFor(url string, waitType string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		_, _, _, err := page.Navigate(url).Do(ctx)
		if err != nil {
			return err
		}

		return waitFor(ctx, waitType)
	}
}

// waitFor listens for events in chromedp and stop loading as soon as given event is match
func waitFor(ctx context.Context, waitType string) error {
	var skipFrameCount int
	var timeout int = defaultTimeout

	browserContext, err := GetBrowserContext(ctx)
	if err != nil {
		return fmt.Errorf("wait for: browser context not set")
	}

	rendererContext, err := GetRendererContext(ctx)
	if errors.Is(err, ErrRendererContextNotFound) {
		debugMessage(
			browserContext.DebugMode,
			"wait for: renderer context not set, use default value",
		)
	} else if err != nil {
		return fmt.Errorf("wait for: get renderer context; %w", err)
	} else {
		if rendererContext.Timeout != 0 {
			timeout = rendererContext.Timeout
		} else {
			timeout = defaultTimeout
		}
		skipFrameCount = rendererContext.SkipFrameCount
	}

	cctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)

	idleCheck := networkIdle{
		navigateFrame:  false,
		frameCount:     0,
		skipFrameCount: skipFrameCount,
	}
	chromedp.ListenTarget(cctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *page.EventFrameNavigated:
			msg := fmt.Sprintf("Navigate ID: %s, Frame ID: %s", e.Type, e.Frame.ID)
			debugMessage(browserContext.DebugMode, msg)
			if !idleCheck.navigateFrame {
				idleCheck.frameId = e.Frame.ID.String()
			}
			idleCheck.navigateFrame = true
		case *page.EventLifecycleEvent:
			switch waitType {
			case "networkIdle":
				if isNetworkIdle(&idleCheck, e) {
					cancel()
				}
			case "InteractiveTime":
				if isInteractiveTime(e) {
					cancel()
				}
			}
		}
	})

	select {
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			fmt.Printf("waitFor err: %s\n", err)
		}
		// fmt.Println("ctx done")
		return ctx.Err()
	case <-cctx.Done():
		if err := cctx.Err(); errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("waitFor err: %w", err)
		}
		// fmt.Println("waitFor: cctx done")
		return nil
	}
}

// isNetworkIdle check if networkIdle met complete state.
// Complete state is met if input event frame id is same
// as the first frame id from EventFrameNavigated
func isNetworkIdle(n *networkIdle, e *page.EventLifecycleEvent) bool {
	if e.Name == "networkIdle" && n.navigateFrame {
		// fmt.Printf("Idle count: %d, Frame id: %s\n", n.idleCount, n.frameId)
		// fmt.Printf("Event name: %s, Frame ID: %s\n", e.Name, e.FrameID)
		frameCountExit := false
		if n.frameId == e.FrameID.String() {
			switch n.frameCount < n.skipFrameCount {
			case true:
				n.frameCount++
			case false:
				frameCountExit = true
			}
		}
		return frameCountExit
	}

	return false
}

// isInteractiveTime check if life cycle have met InteractiveTime event.
func isInteractiveTime(e *page.EventLifecycleEvent) bool {
	// if e.Name == "InteractiveTime" {
	// 	fmt.Printf("Event name: %s, Frame ID: %s\n", e.Name, e.FrameID)
	// }
	return e.Name == "InteractiveTime"
}

// setPdfParams read PDF context input and output PrintToPDFParams
// according to context settings
func setPdfParams(pc *PdfContext) *page.PrintToPDFParams {
	// Default value for parameters if not set
	if pc.PaperWidthCm == 0 {
		pc.PaperWidthCm = 21
	}
	if pc.PaperHeightCm == 0 {
		pc.PaperHeightCm = 29.7
	}
	if pc.DisplayHeaderFooter {
		if pc.MarginTopCm == 0 {
			pc.MarginTopCm = 1
		}
		if pc.MarginBottomCm == 0 {
			pc.MarginBottomCm = 1
		}
	}

	return &page.PrintToPDFParams{
		Landscape:           pc.Landscape,
		DisplayHeaderFooter: pc.DisplayHeaderFooter,
		PaperWidth:          cmToInch(pc.PaperWidthCm),
		PaperHeight:         cmToInch(pc.PaperHeightCm),
		MarginTop:           cmToInch(pc.MarginTopCm),
		MarginBottom:        cmToInch(pc.MarginBottomCm),
		MarginLeft:          cmToInch(pc.MarginLeftCm),
		MarginRight:         cmToInch(pc.MarginRightCm),
	}
}

// cmToInch convert centimeter input to inch with two decimal precision
func cmToInch(cm float64) float64 {
	return math.Round((cm/2.54)*100) / 100
}

// debugMessage print out msg if debugMode is true
func debugMessage(debugMode bool, msg string) {
	if debugMode {
		log.Printf("%s\n", msg)
	}
}
