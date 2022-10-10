package renderer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// RenderPage rendered given url in browser and returns result html content
func RenderPage(ctx context.Context, urlStr string) ([]byte, error) {
	fmt.Printf("ctx headless: %v\n", ctx.Value("headless"))

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", ctx.Value("headless")),
		chromedp.Flag("blink-settings", fmt.Sprintf("imagesEnbled=%v", ctx.Value("imageLoad"))),
		chromedp.WindowSize(ctx.Value("windowWidth").(int), ctx.Value("windowHeight").(int)),
	)
	fmt.Printf("Rendering: %s\n", urlStr)

	start := time.Now()
	ctx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var res string
	err := chromedp.Run(ctx,
		chromedp.Tasks{
			navigateAndWaitFor(urlStr, "InteractiveTime"),
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
	fmt.Println(duration)
	return []byte(res), nil
}

// navigateAndWaitFor is defined as task of chromedp for rendering step
func navigateAndWaitFor(url string, eventName string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		_, _, _, err := page.Navigate(url).Do(ctx)
		if err != nil {
			return err
		}

		return waitFor(ctx, eventName)
	}
}

// waitFor listens for events in chromedp and stop loading as soon as given event is match
func waitFor(ctx context.Context, eventName string) error {
	timeout := ctx.Value("timeout").(int)
	cctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	chromedp.ListenTarget(cctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *page.EventLifecycleEvent:
			// fmt.Printf("Event name: %s\n", e.Name)
			if e.Name == eventName {
				cancel()
			}
		case *fetch.EventRequestPaused:
			go func() {
				c := chromedp.FromContext(ctx)
				ctx := cdp.WithExecutor(ctx, c.Target)

				if e.ResourceType == network.ResourceTypeImage {
					fetch.FailRequest(e.RequestID, network.ErrorReasonBlockedByClient).Do(ctx)
				} else {
					fetch.ContinueRequest(e.RequestID).Do(ctx)
				}
			}()
		}
	})

	select {
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			fmt.Printf("waitFor err: %s\n", err)
		}
		fmt.Println("ctx done")
		return ctx.Err()
	case <-cctx.Done():
		if err := cctx.Err(); errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("waitFor err: %w", err)
		}
		fmt.Println("waitFor: cctx done")
		return nil
	}
}
