# renderer

Golang module for executing url rendering with chromedp

## Requirements

chromium / chrome browser installed on host

## Options configuration

### Browser

Browser config values:

- `IdleType`: Method to detemine render is complete
  - Type: string (valid values: auto, networkIdle, InteractiveTime)
  - Default: networkIdle
- `BrowserExecPath`: Manually set chrome / chromium browser's executable path
  - Type: String
  - Default: Empty string (Auto detect)
- `Container`: Use this option to execute chrome / chromium browser in container
  environment
  - Type: bool
  - Default: false
- `DebugMode`: Output debug message if set to true
  - Type: bool
  - Default: false
- `ChromiumDebug`: Output debug message on chromium if set to true
  - Type: bool
  - Default: false

### Renderer

Renderer options values:

- `BrowserOpts`: Browser configuration
  - Type: BrowserConf
- `Headless`: Browser execution mode
  - Type: bool
  - Default: false
- `WindowWidth`: Width of browser's window size
  - Type: Int
  - Default: 1000
- `WindowHeight`: Height of browser's window size
  - Type: Int
  - Default: 1000
- `Timeout`: Seconds before rendering timeout
  - Type: Int
  - Default: 30
- `ImageLoad`: Load image when rendering
  - Type: bool
  - Default: false
- `SkipFrameCount`: Skip first n framces with same id as init frame, only valid
  with idleType=networkIdle (Use on page with protection like CloudFlare)
  - Type: Int
  - Default: 0
- `UserAgent`: Set custom user agent value
  - Type: string
  - Default: Empty string (Will user default user agent of the browser)

**Note:** Default renderer option `defaultRendererOption` will be used if not set no explicit option is set.

#### Example

See usage example at [examples](examples/render/main.go)

**Build Example / Test**

```bash
cd examples/render
go build
```

**Run Example / Test**

```
Usage: ./render <url>
  -bHeight int
        height of browser window's size (default 1080)
  -bWidth int
        width of browser window's size (default 1920)
  -browserPath string
        manually set browser executable path
  -chromiumDebug
        turn on for chromium debug message output
  -container
        indicate if running in container (docker / lambda) environment
  -debug
        turn on for outputing debug message
  -headless
        automation browser execution mode (default true)
  -idleType string
        how to determine loading idle and return, valid input: auto, networkIdle, InteractiveTime (default "auto")
  -imageLoad
        indicate if load image when rendering
  -skipFrameCount int
        skip first n frames with same id as init frame, only valid with idleType=networkIdle
  -timeout int
        seconds before timeout when rendering (default 30)
  -userAgent string
        set custom user agent for sending request in automation browser
```

### Render PDF

Pdf options values:

- `BrowserOpts`: Browser configuration
  - Type: BrowserConf
- `Landscape`: Set paper orientation to landscape
  - Type: bool
  - Default: false
- `DisplayHeaderFooter`: Display header and footer
  - Type: bool
  - Default: false
- `PaperWidthCm`: Paper width in centimeter
  - Type: float64
  - Default: 21
- `PaperHeightCm`: Paper height in centimeter
  - Type: float64
  - Default: 29.7
- `MarginTopCm`: Top margin in centimeter
  - Type: float64
  - Default: 1
- `MarginBottomCm`: Bottom margin in centimeter
  - Type: float64
  - Default: 1
- `MarginLeftCm`: Left margin in centimeter
  - Type: float64
  - Default: 1
- `MarginRigthCm`: Right margin in centimeter
  - Type: float64
  - Default: 1

#### Example

See usage example at [examples](examples/pdf/main.go)

**Build Example / Test**

```bash
cd examples/pdf
go build
```

**Run Example / Test**

```
Usage: ./pdf <url>
  -browserPath string
        manually set browser executable path
  -chromiumDebug
        turn on for chrome debug message output
  -container
        indicate if running in container (docker / lambda) environment
  -debug
        turn on for outputing debug message
  -headerFooter
        show header and footer
  -idleType string
        how to determine loading idle and return, valid input: auto, networkIdle, InteractiveTime (default "auto")
  -landscape
        create pdf in landscape layout
  -marginBottom float
        bottom margin in centimeter (default 1)
  -marginLeft float
        left margin in centimeter (default 1)
  -marginRight float
        right margin in centimeter (default 1)
  -marginTop float
        top margin in centimeter (default 1)
  -paperHeight float
        paper height in centimeter
  -paperWidth float
        paper width in centimeter
```
