package renderer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"math"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// networkIdle is a struct to keep track of networkIdle state
type networkIdle struct {
	navigateFrame  bool
	frameId        string
	skipFrameCount int
	frameCount     int
}

type Renderer struct {
	logger *slog.Logger
}

// NewRenderer create new renderer instance. Function options can be pass as argument
// to set for configuration (logger for now)
func NewRenderer(options ...WithOption) *Renderer {
	r := &Renderer{
		logger: slog.Default(),
	}

	for _, option := range options {
		option(r)
	}

	return r
}

// RenderPage render the given url with automated chrome browser and return back the
// result html content. RendererOption is use for setting the behavior of the automated
// browser while rendering the page.
func (r *Renderer) RenderPage(urlStr string, opts *RendererOption) ([]byte, error) {
	if opts == nil {
		opts = &defaultRendererOption
	}

	if opts.BrowserOpts.IdleType != "networkIdle" &&
		opts.BrowserOpts.IdleType != "InteractiveTime" {
		return nil, fmt.Errorf("invalid idleType %s", opts.BrowserOpts.IdleType)
	}

	chromeOpts := chromedp.DefaultExecAllocatorOptions[:]
	if opts.BrowserOpts.BrowserExecPath != "" {
		chromeOpts = append(chromeOpts, chromedp.ExecPath(opts.BrowserOpts.BrowserExecPath))
	}
	if opts.BrowserOpts.Container {
		r.logger.Debug("Set configuration for container environment", slog.String("url", urlStr))
		chromeOpts = append(chromeOpts,
			chromedp.Flag("disable-setuid-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("single-process", true),
			chromedp.Flag("no-zygote", true),
			chromedp.NoSandbox,
		)
	}

	if opts.UserAgent != "" {
		chromeOpts = append(chromeOpts, chromedp.UserAgent(opts.UserAgent))
	}

	chromeOpts = append(
		chromeOpts,
		chromedp.Flag("blink-settings", fmt.Sprintf("imagesEnbled=%t", opts.ImageLoad)),
	)
	chromeOpts = append(chromeOpts, chromedp.WindowSize(opts.WindowWidth, opts.WindowHeight))
	chromeOpts = append(chromeOpts, chromedp.Flag("headless", opts.Headless))

	start := time.Now()
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), chromeOpts...)
	defer cancel()
	if opts.BrowserOpts.ChromiumDebug {
		ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(r.logger.Debug))
	} else {
		ctx, cancel = chromedp.NewContext(ctx)
	}
	defer cancel()

	var resp string
	err := chromedp.Run(ctx,
		chromedp.Tasks{
			r.navigateAndWaitFor(urlStr, opts),
			chromedp.ActionFunc(func(ctx context.Context) error {
				node, err := dom.GetDocument().Do(ctx)
				if err != nil {
					r.logger.Error(err.Error(), slog.String("url", urlStr))
					return fmt.Errorf("renderPage(%v): %w", urlStr, err)
				}
				resp, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
				if err != nil {
					r.logger.Error(err.Error(), slog.String("url", urlStr))
					return fmt.Errorf("renderPage(%v): %w", urlStr, err)
				}
				return nil
			}),
		},
	)
	duration := time.Since(start)
	r.logger.Debug(fmt.Sprintf("Render time: %v", duration), slog.String("url", urlStr))
	if err != nil {
		r.logger.Error(fmt.Sprintf("chromedp run error: %s", err), slog.String("url", urlStr))
		return nil, err
	}

	return []byte(resp), nil
}

// RenderPdf generate and return the pdf as byte array from the given url using
// automated chrome browser. PdfOption is for setting the style of the generated pdf.
func (r *Renderer) RenderPdf(urlStr string, opts *PdfOption) ([]byte, error) {
	pdfParams := page.PrintToPDF()
	if opts == nil {
		opts = &defaultPdfOption
	} else {
		pdfParams = opts.setParams()
	}

	if opts.BrowserOpts.IdleType != "networkIdle" &&
		opts.BrowserOpts.IdleType != "InteractiveTime" {
		return nil, fmt.Errorf("invalid idleType %s", opts.BrowserOpts.IdleType)
	}

	chromeOpts := chromedp.DefaultExecAllocatorOptions[:]
	if opts.BrowserOpts.BrowserExecPath != "" {
		chromeOpts = append(chromeOpts, chromedp.ExecPath(opts.BrowserOpts.BrowserExecPath))
	}

	if opts.BrowserOpts.Container {
		r.logger.Debug("Set configuration for container environment", slog.String("url", urlStr))
		chromeOpts = append(chromeOpts,
			chromedp.Flag("disable-setuid-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("single-process", true),
			chromedp.Flag("no-zygote", true),
			chromedp.NoSandbox,
		)
	}

	start := time.Now()
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), chromeOpts...)
	defer cancel()
	if opts.BrowserOpts.ChromiumDebug {
		ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(log.Printf))
	} else {
		ctx, cancel = chromedp.NewContext(ctx)
	}
	defer cancel()

	var resp []byte
	err := chromedp.Run(ctx,
		r.navigateAndWaitFor(urlStr, &defaultRendererOption),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := pdfParams.Do(ctx)
			if err != nil {
				return fmt.Errorf("renderPdf(%v): %w", urlStr, err)
			}
			resp = buf
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp run: %w", err)
	}

	duration := time.Since(start)
	r.logger.Debug(fmt.Sprintf("Render time: %v", duration), slog.String("url", urlStr))

	return resp, nil
}

// navigateAndWaitFor is defined as task of chromedp for rendering step
func (r *Renderer) navigateAndWaitFor(url string, opts *RendererOption) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		_, _, _, err := page.Navigate(url).Do(ctx)
		if err != nil {
			return err
		}

		return r.waitFor(ctx, opts)
	}
}

// waitFor listens for events in chromedp and stop loading as soon as given event is matched
// or timeout is reached. Two events are supported currently: networkIdle and InteractiveTime.
func (r *Renderer) waitFor(ctx context.Context, opts *RendererOption) error {
	cctx, cancel := context.WithTimeout(ctx, time.Duration(opts.Timeout)*time.Second)

	idleCheck := networkIdle{
		navigateFrame:  false,
		frameCount:     0,
		skipFrameCount: opts.SkipFrameCount,
	}
	chromedp.ListenTarget(cctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *page.EventFrameNavigated:
			msg := fmt.Sprintf("Navigate ID: %s, Frame ID: %s", e.Type, e.Frame.ID)
			r.logger.Debug(msg)
			if !idleCheck.navigateFrame {
				idleCheck.frameId = e.Frame.ID.String()
			}
			idleCheck.navigateFrame = true
		case *page.EventLifecycleEvent:
			switch opts.BrowserOpts.IdleType {
			case "networkIdle":
				if r.isNetworkIdle(&idleCheck, e) {
					cancel()
				}
			case "InteractiveTime":
				if r.isInteractiveTime(e) {
					cancel()
				}
			}
		}
	})

	select {
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			r.logger.Info(fmt.Sprintf("waitFor err: %s", err))
		}
		r.logger.Debug("waitFor: ctx done")
		return ctx.Err()
	case <-cctx.Done():
		if err := cctx.Err(); errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("waitFor err: %w", err)
		}
		r.logger.Debug("waitFor: cctx done")
		return nil
	}
}

// isNetworkIdle check if networkIdle met complete state.
// Complete state is met if input event frame id is same
// as the first frame id from EventFrameNavigated
func (r *Renderer) isNetworkIdle(n *networkIdle, e *page.EventLifecycleEvent) bool {
	if e.Name == "networkIdle" && n.navigateFrame {
		r.logger.Debug(fmt.Sprintf("Idle count: %d, Frame id: %s", n.frameCount, n.frameId))
		r.logger.Debug(fmt.Sprintf("Event name: %s, Frame ID: %s", e.Name, e.FrameID))
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
func (r *Renderer) isInteractiveTime(e *page.EventLifecycleEvent) bool {
	if e.Name == "InteractiveTime" {
		r.logger.Debug(fmt.Sprintf("Event name: %s, Frame ID: %s", e.Name, e.FrameID))
	}
	return e.Name == "InteractiveTime"
}

// cmToInch convert centimeter input to inch with two decimal precision
func cmToInch(cm float64) float64 {
	return math.Round((cm/2.54)*100) / 100
}
