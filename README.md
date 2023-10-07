# renderer
Golang module for executing url rendering with chromedp

## Requirements
chrome browser installed on host

## Renderer
RendererContext values:
- `headless`: Browser execution mode
    - Type: bool
    - Default: false
- `windowWidth`: Width of browser's window size
    - Type: Int
    - Default: 1000
- `windowHeight`: Height of browser's window size
    - Type: Int
    - Default: 1000
- `timeout`: Seconds before rendering timeout
    - Type: Int
    - Default: 30
- `imageLoad`: Load image when rendering 
    - Type: bool
    - Default: false
- `idleType`: Method to detemine render is complete
    - Type: string (valid values: networkIdle, InteractiveTime)
    - Default: networkIdle
- `skipFrameCount`: Skip first n framces with same id as init frame, only valid with idleType=networkIdle (Use on page with protection like CloudFlare)
    - Type: Int
    - Default: 0

### Example
See usage example at [examples](examples/render/main.go)

#### Build Example / Test
```bash
go build -o render.out
```

#### Run Example / Test
```
Usage of ./render.out:
  -bHeight int
      height of browser window's size (default 1080)
  -bWidth int
      width of browser window's size (default 1920)
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
- `idleType`: Method to detemine render is complete
    - Type: string (valid values: networkIdle, InteractiveTime)
    - Default: networkIdle

### Example
See usage example at [examples](examples/pdf/main.go)

#### Build Example / Test
```bash
go build -o pdf.out
```

#### Run Example / Test
```
Usage of ./pdf.out:
  -landscape bool
      create pdf in landscape layout (default false)
  -headerFooter bool
      show header and footer (default false)
  -paperWidth float
      paper width in centimeter (default 21, A4 size)
  -paperHeight float
      paper height in centimeter (default 29.7, A4 size)
  -marginTop float
      top margin in centimeter (default 1)
  -marginBottom float
      bottom margin in centimeter (default 1)
  -marginLeft float
      left margin in centimeter (default 1)
  -marginRigth float
      right margin in centimeter (default 1)
  -idleType networkIdle|InteractiveTime
      how to determiine loading idle and return (default networkIdle)
```


