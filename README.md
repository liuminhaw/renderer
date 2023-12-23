# renderer
Golang module for executing url rendering with chromedp

## Requirements
chrome browser installed on host

## Renderer
RendererContext values:
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
- `IdleType`: Method to detemine render is complete
    - Type: string (valid values: networkIdle, InteractiveTime)
    - Default: networkIdle
- `SkipFrameCount`: Skip first n framces with same id as init frame, only valid with idleType=networkIdle (Use on page with protection like CloudFlare)
    - Type: Int
    - Default: 0
- `BrowserExecPath`: Manually set chrome / chromium browser's executable path
    - Type: String
    - Default: Empty string (Auto detect)

### Example
See usage example at [examples](examples/render/main.go)

#### Build Example / Test
```bash
go build -o render.out
```

#### Run Example / Test
```
Usage of ./render:
  -bHeight int
      height of browser window's size (default 1080)
  -bWidth int
      width of browser window's size (default 1920)
  -browserPath string
      manually set browser executable path
  -headless
      automation browser execution mode (default true)
  -idleType string
      how to determine loading idle and return, valid input: networkIdle, InteractiveTime (default "networkIdle")
  -imageLoad
      indicate if load image when rendering
  -skipFrameCount int
      skip first n frames with same id as init frame, only valid with idleType=networkIdle
  -timeout int
      seconds before timeout when rendering (default 30)
```

## Render PDF
PdfContext values:
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
- `IdleType`: Method to detemine render is complete
    - Type: string (valid values: networkIdle, InteractiveTime)
    - Default: networkIdle
- `BrowserExecPath`: Manually set chrome / chromium browser's executable path
    - Type: String
    - Default: Empty string (Auto detect)

### Example
See usage example at [examples](examples/pdf/main.go)

#### Build Example / Test
```bash
go build -o pdf.out
```

#### Run Example / Test
```
Usage of ./pdf.out:
  -browserPath string
      manually set browser executable path
  -headerFooter
      show header and footer
  -idleType string
      how to determine loading idle and return, valid input: networkIdle, InteractiveTime (default "networkIdle")
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


