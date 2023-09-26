package renderer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type networkIdle struct {
	navigateFrame  bool
	frameId        string
	skipFrameCount int
	frameCount     int
}

// RenderPage rendered given url in browser and returns result html content
func RenderPage(ctx context.Context, urlStr string) ([]byte, error) {
	// fmt.Printf("ctx headless: %v\n", ctx.Value("headless"))
	// windowWidth, windowHeight := 1000, 1000
	// idleType := "networkIdle"

	rendererContext, err := GetRendererContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("render page: %w", err)
	}

	var idleType string
	switch rendererContext.IdleType {
	case "":
		idleType = "networkIdle"
	case "networkIdle", "InteractiveTime":
		idleType = rendererContext.IdleType
	default:
		return nil, fmt.Errorf("render page: invalid idleType %s", rendererContext.IdleType)
	}

	var windowWidth, windowHeight int = 1000, 1000
	if rendererContext.WindowWidth != 0 {
		windowWidth = rendererContext.WindowWidth
	}
	if rendererContext.WindowHeight != 0 {
		windowHeight = rendererContext.WindowHeight
	}

	var opts = chromedp.DefaultExecAllocatorOptions[:]
	opts = append(opts, chromedp.Flag("headless", rendererContext.Headless))
	opts = append(
		opts,
		chromedp.Flag("blink-settings", fmt.Sprintf("imagesEnbled=%v", rendererContext.ImageLoad)),
	)
	opts = append(opts, chromedp.WindowSize(windowWidth, windowHeight))

	// fmt.Printf("Rendering: %s\n", urlStr)

	start := time.Now()
	ctx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
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
	fmt.Printf("Render time: %v\n", duration)
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
	rendererContext, err := GetRendererContext(ctx)
	if err != nil {
		return fmt.Errorf("wait for: %w", err)
	}

	var timeout int = 30
	if rendererContext.Timeout != 0 {
		timeout = rendererContext.Timeout
	}
	cctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)

	skipFrameCount := rendererContext.SkipFrameCount

	idleCheck := networkIdle{
		navigateFrame:  false,
		frameCount:     0,
		skipFrameCount: skipFrameCount,
	}
	chromedp.ListenTarget(cctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *page.EventFrameNavigated:
			fmt.Printf("Navigate ID: %s, Frame ID: %s\n", e.Type, e.Frame.ID)
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
		var frameCountExit = false
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
