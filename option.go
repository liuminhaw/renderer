package renderer

import (
	"log/slog"

	"github.com/chromedp/cdproto/page"
)

const (
	defaultIdleType     string = "networkIdle"
	defaultTimeout      int    = 30
	defaultWindowWidth  int    = 1920
	defaultWindowHeight int    = 1080
)

type BrowserConf struct {
	IdleType        string
	BrowserExecPath string
	ChromiumDebug   bool
	DebugMode       bool
	Container       bool
}

type RendererOption struct {
	BrowserOpts    BrowserConf
	Headless       bool
	WindowWidth    int
	WindowHeight   int
	Timeout        int
	ImageLoad      bool
	SkipFrameCount int
	UserAgent      string
}

var defaultRendererOption = RendererOption{
	BrowserOpts: BrowserConf{
		IdleType: defaultIdleType,
	},
	WindowWidth:  defaultWindowWidth,
	WindowHeight: defaultWindowHeight,
	Timeout:      defaultTimeout,
}

type PdfOption struct {
	BrowserOpts         BrowserConf
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
var defaultPdfOption = PdfOption{
	BrowserOpts: BrowserConf{
		IdleType: defaultIdleType,
	},
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
