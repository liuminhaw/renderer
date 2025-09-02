package renderer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type Renderer struct {
	logger           *slog.Logger
	idleCheck        *networkIdle     // use for network idle check
	interactiveCheck *interactiveTime // use for interactive time check
}

// NewRenderer create new renderer instance. Function options can be pass as argument
// to set for configuration (logger for now)
func NewRenderer(options ...WithOption) *Renderer {
	r := &Renderer{
		logger:           slog.Default(),
		interactiveCheck: newInteractiveTime(),
	}

	for _, option := range options {
		option(r)
	}

	if r.idleCheck == nil {
		r.idleCheck = newNetworkIdle(defaultNetworkIdleWait, defaultNetworkIdleMaxInflight)
	}

	return r
}

// RenderPage render the given url with automated chrome browser and return back the
// result html content. RendererOption is use for setting the behavior of the automated
// browser while rendering the page.
func (r *Renderer) RenderPage(urlStr string, opts *RendererOption) ([]byte, error) {
	if opts == nil {
		opts = &DefaultRendererOption
	}

	if !IsValidIdleType(opts.BrowserOpts.IdleType) {
		return nil, fmt.Errorf("invalid idleType %s", opts.BrowserOpts.IdleType)
	}

	chromeOpts := setChromeOpts(opts)

	start := time.Now()
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), chromeOpts...)
	defer cancel()
	if opts.BrowserOpts.ChromiumDebug {
		ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(r.logger.Debug))
	} else {
		ctx, cancel = chromedp.NewContext(ctx)
	}
	defer cancel()

	// Attach listener before navigating to the page
	r.Listen(ctx, opts.BrowserOpts)

	var resp string
	err := chromedp.Run(ctx,
		chromedp.Tasks{
			network.Enable(),
			r.navigateAndWaitFor(urlStr, *opts),
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
		opts = &DefaultPdfOption
	} else {
		pdfParams = opts.setParams()
	}

	if !IsValidIdleType(opts.BrowserOpts.IdleType) {
		return nil, fmt.Errorf("invalid idleType %s", opts.BrowserOpts.IdleType)
	}

	chromeOpts := setChromeOpts(opts)

	start := time.Now()
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), chromeOpts...)
	defer cancel()
	if opts.BrowserOpts.ChromiumDebug {
		ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(r.logger.Debug))
	} else {
		ctx, cancel = chromedp.NewContext(ctx)
	}
	defer cancel()

	// Attach listener before navigating to the page
	r.Listen(ctx, opts.BrowserOpts)

	var resp []byte
	err := chromedp.Run(ctx,
		network.Enable(),
		r.navigateAndWaitFor(urlStr, *opts),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := pdfParams.Do(ctx)
			if err != nil {
				return fmt.Errorf("renderPdf(%v): %w", urlStr, err)
			}
			resp = buf
			return nil
		}),
	)
	duration := time.Since(start)
	r.logger.Debug(fmt.Sprintf("Render time: %v", duration), slog.String("url", urlStr))
	if err != nil {
		r.logger.Error(fmt.Sprintf("chromedp run error: %s", err), slog.String("url", urlStr))
		return nil, err
	}

	return resp, nil
}

// navigateAndWaitFor is defined as task of chromedp for rendering step
func (r *Renderer) navigateAndWaitFor(url string, opts chromedpOption) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		_, _, _, _, err := page.Navigate(url).Do(ctx)
		if err != nil {
			return err
		}

		return r.waitFor(ctx, opts)
	}
}

func (r *Renderer) Listen(ctx context.Context, conf BrowserConf) {
	var mainFrame cdp.FrameID
	chromedp.ListenTarget(ctx, func(ev any) {
		switch e := ev.(type) {
		case *page.EventFrameNavigated:
			if e.Frame.ParentID == "" {
				mainFrame = e.Frame.ID
			}

		case *network.EventRequestWillBeSent:
			if !enabledIdleType([]string{"auto", "networkIdle"}, conf.IdleType) {
				return
			}
			if isNoisyRequest(e) {
				return
			}

			var activeLen int
			var msg string
			if e.Type == network.ResourceTypeDocument && e.FrameID == mainFrame {
				activeLen = r.idleCheck.addByLoader(e.LoaderID, e.RequestID, e.Type, e.Request.URL)
				msg = fmt.Sprintf(
					"Type: network.EventRequestWillBeSent - %s, mainFrameID: %s, FrameID: %s, LoaderID: %s, ID: %s, Active: %d",
					e.Type,
					mainFrame,
					e.FrameID,
					e.LoaderID,
					e.RequestID,
					activeLen,
				)
			} else if e.Type != network.ResourceTypeDocument {
				activeLen = r.idleCheck.add(e.RequestID, e.Type, e.Request.URL)
				msg = fmt.Sprintf(
					"Type: network.EventRequestWillBeSent - %s, mainFrameID: %s, FrameID: %s, ID: %s, Active: %d",
					e.Type,
					mainFrame,
					e.FrameID,
					e.RequestID,
					activeLen,
				)
			}
			r.logger.Debug(msg)
		case *network.EventLoadingFinished:
			if !enabledIdleType([]string{"auto", "networkIdle"}, conf.IdleType) {
				return
			}
			activeLen := r.idleCheck.remove(e.RequestID)
			msg := fmt.Sprintf(
				"Type: network.EventLoadingFinished, ID: %s, Active: %d",
				e.RequestID,
				activeLen,
			)
			r.logger.Debug(msg)
		case *network.EventLoadingFailed:
			if !enabledIdleType([]string{"auto", "networkIdle"}, conf.IdleType) {
				return
			}
			activeLen := r.idleCheck.remove(e.RequestID)
			msg := fmt.Sprintf(
				"Type: network.EventLoadingFailed, ID: %s, Active: %d",
				e.RequestID,
				activeLen,
			)
			r.logger.Debug(msg)
		case *page.EventLifecycleEvent:
			if (e.Name == "load" || e.Name == "networkIdle") && e.FrameID == mainFrame {
				if !enabledIdleType([]string{"auto", "networkIdle"}, conf.IdleType) {
					return
				}
				if requestID, ok := r.idleCheck.byLoader[e.LoaderID]; ok {
					activeLen := r.idleCheck.remove(requestID)
					msg := fmt.Sprintf(
						"Type: EventLifecycle, Name: %s, Loader ID: %s, Active: %d",
						e.Name,
						e.LoaderID,
						activeLen,
					)
					r.logger.Debug(msg)
				}
			}

			if e.Name == "InteractiveTime" {
				if !enabledIdleType([]string{"auto", "InteractiveTime"}, conf.IdleType) {
					return
				}
				r.logger.Debug(fmt.Sprintf("Event name: %s, Frame ID: %s", e.Name, e.FrameID))
				r.interactiveCheck.done <- struct{}{} // cancel the context when InteractiveTime is met
			}
		}
	})
}

// waitFor listens for events in chromedp and stop loading as soon as given event is matched
// or timeout is reached.
// Two events are currently supported: networkIdle and InteractiveTime.
func (r *Renderer) waitFor(ctx context.Context, opts chromedpOption) error {
	browserConf := opts.readBrowserConf()
	rendererConf := opts.readRendererConf()

	cctx, cancel := context.WithTimeout(ctx, time.Duration(rendererConf.Timeout)*time.Second)
	defer cancel()

	// Initial arm (in case we’re already quiet)
	if enabledIdleType([]string{"auto", "networkIdle"}, browserConf.IdleType) {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		go func() {
			for {
				select {
				case <-ticker.C:
					r.idleCheck.promoteLongPoll()
				case <-cctx.Done():
					return
				}
			}
		}()

		r.idleCheck.mu.Lock()
		r.idleCheck.startOrResetTimer()
		r.idleCheck.mu.Unlock()
	}

	select {
	case <-r.idleCheck.done:
		r.idleCheck.mu.Lock()
		r.idleCheck.stopped = true
		if r.idleCheck.idleTimer != nil {
			_ = r.idleCheck.idleTimer.Stop() // stop the timer if it is running
		}
		r.idleCheck.mu.Unlock()
		r.logger.Debug("waitFor: networkIdle done")
		return nil
	case <-r.interactiveCheck.done:
		r.idleCheck.mu.Lock()
		r.idleCheck.stopped = true
		if r.idleCheck.idleTimer != nil {
			_ = r.idleCheck.idleTimer.Stop() // stop the timer if it is running
		}
		r.idleCheck.mu.Unlock()
		r.logger.Debug("waitFor: InteractiveTime done")
		return nil
	case <-cctx.Done():
		r.idleCheck.mu.Lock()
		r.idleCheck.stopped = true
		if r.idleCheck.idleTimer != nil {
			_ = r.idleCheck.idleTimer.Stop() // stop the timer if it is running
		}
		r.idleCheck.mu.Unlock()
		if err := cctx.Err(); errors.Is(err, context.DeadlineExceeded) {
			r.logger.Debug(fmt.Sprintf("Remains active network request: %+v", r.idleCheck.active))
			for _, info := range r.idleCheck.active {
				r.logger.Debug(fmt.Sprintf("Remain active info: %+v", info))
			}
			return fmt.Errorf("waitFor err: %w", err)
		}
		r.logger.Debug("waitFor: cctx done")
		return nil
	}
}

func setChromeOpts(opts chromedpOption) []chromedp.ExecAllocatorOption {
	browserConf := opts.readBrowserConf()
	rendererConf := opts.readRendererConf()

	chromeOpts := chromedp.DefaultExecAllocatorOptions[:]
	if browserConf.BrowserExecPath != "" {
		chromeOpts = append(chromeOpts, chromedp.ExecPath(browserConf.BrowserExecPath))
	}
	if browserConf.Container {
		// r.logger.Debug("Set configuration for container environment", slog.String("url", urlStr))
		chromeOpts = append(chromeOpts,
			chromedp.Flag("disable-setuid-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("single-process", true),
			chromedp.Flag("no-zygote", true),
			chromedp.NoSandbox,
		)
	}

	if rendererConf.UserAgent != "" {
		chromeOpts = append(chromeOpts, chromedp.UserAgent(rendererConf.UserAgent))
	}

	chromeOpts = append(
		chromeOpts,
		chromedp.Flag("blink-settings", fmt.Sprintf("imagesEnabled=%t", rendererConf.ImageLoad)),
		chromedp.WindowSize(rendererConf.WindowWidth, rendererConf.WindowHeight),
		chromedp.Flag("headless", rendererConf.Headless),
		chromedp.WSURLReadTimeout(time.Duration(rendererConf.Timeout)*time.Second),
	)

	return chromeOpts
}

// cmToInch convert centimeter input to inch with two decimal precision
func cmToInch(cm float64) float64 {
	return math.Round((cm/2.54)*100) / 100
}

func enabledIdleType(idleTypes []string, idleType string) bool {
	if slices.Contains(idleTypes, idleType) {
		return true
	}
	return false
}
