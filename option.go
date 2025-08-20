package renderer

import (
	"log/slog"
	"slices"
	"time"

	"github.com/chromedp/cdproto/page"
)

const (
	defaultIdleType               string        = "auto"
	defaultTimeout                int           = 30
	defaultWindowWidth            int           = 1920
	defaultWindowHeight           int           = 1080
	defaultNetworkIdleWait        time.Duration = 500 * time.Millisecond
	defaultNetworkIdleMaxInflight int           = 0
)

type chromedpOption interface {
	readBrowserConf() BrowserConf
	readRendererConf() RendererConf
}

// IsValidIdleType checks if the given idleType is valid
func IsValidIdleType(idleType string) bool {
	validTypes := []string{"auto", "networkIdle", "InteractiveTime"}

	return slices.Contains(validTypes, idleType)
}

type BrowserConf struct {
	IdleType        string
	BrowserExecPath string
	ChromiumDebug   bool
	DebugMode       bool
	Container       bool
}

type RendererConf struct {
	Headless     bool
	WindowWidth  int
	WindowHeight int
	Timeout      int
	ImageLoad    bool
	UserAgent    string
}

var DefaultRendererConf = RendererConf{
	Headless:     true,
	WindowWidth:  defaultWindowWidth,
	WindowHeight: defaultWindowHeight,
	Timeout:      defaultTimeout,
}

type RendererOption struct {
	BrowserOpts BrowserConf
	Opts        RendererConf
}

func (opts RendererOption) readBrowserConf() BrowserConf {
	return opts.BrowserOpts
}

func (opts RendererOption) readRendererConf() RendererConf {
	return opts.Opts
}

var DefaultRendererOption = RendererOption{
	BrowserOpts: BrowserConf{
		IdleType: defaultIdleType,
	},
	Opts: DefaultRendererConf,
}

type PdfOption struct {
	BrowserOpts         BrowserConf
	RendererOpts        RendererConf
	Landscape           bool
	DisplayHeaderFooter bool
	PaperWidthCm        float64
	PaperHeightCm       float64
	MarginTopCm         float64
	MarginBottomCm      float64
	MarginLeftCm        float64
	MarginRightCm       float64
}

// Chrome DevTools Protocol reference for default values:
// https://chromedevtools.github.io/devtools-protocol/tot/Page/#method-printToPDF
var DefaultPdfOption = PdfOption{
	BrowserOpts: BrowserConf{
		IdleType: defaultIdleType,
	},
	RendererOpts: DefaultRendererConf,
}

func (opts PdfOption) readBrowserConf() BrowserConf {
	return opts.BrowserOpts
}

func (opts PdfOption) readRendererConf() RendererConf {
	return opts.RendererOpts
}

// setPdfParams read PDF context input and output PrintToPDFParams
// according to context settings
func (opts *PdfOption) setParams() *page.PrintToPDFParams {
	params := *page.PrintToPDF()

	// Default value for parameters if not set
	params.Landscape = opts.Landscape
	params.DisplayHeaderFooter = opts.DisplayHeaderFooter

	if opts.PaperWidthCm != 0 {
		params.PaperWidth = cmToInch(opts.PaperWidthCm)
	}
	if opts.PaperHeightCm != 0 {
		params.PaperHeight = cmToInch(opts.PaperHeightCm)
	}
	if opts.MarginTopCm != 0 {
		params.MarginTop = cmToInch(opts.MarginTopCm)
	}
	if opts.MarginBottomCm != 0 {
		params.MarginBottom = cmToInch(opts.MarginBottomCm)
	}
	if opts.MarginLeftCm != 0 {
		params.MarginLeft = cmToInch(opts.MarginLeftCm)
	}
	if opts.MarginRightCm != 0 {
		params.MarginRight = cmToInch(opts.MarginRightCm)
	}

	return &params
}

// Function options
type WithOption func(*Renderer)

// WithLogger can be provided to NewRenderer function to set logger for the Renderer
func WithLogger(logger *slog.Logger) WithOption {
	return func(r *Renderer) {
		r.logger = logger
	}
}

// WithIdleCheck can be provided to NewRenderer function to determine how to check
// if the network is idle before returning the result.
// network is considered idle when there are less than or equal to maxInflight
// requests in action in the last `idleWait` time duration window.
func WithIdleCheck(idleWait time.Duration, maxInflight int) WithOption {
	return func(r *Renderer) {
		r.idleCheck = newNetworkIdle(idleWait, maxInflight)
	}
}
