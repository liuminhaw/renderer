# renderer
Golang module for executing url rendering with chromedp

## Requirements
chrome browser installed on host

## Renderer
Context values:
- `headless`: Browser execution mode
    - Type: bool
    - Default: true
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
    - Default: true
- `idleType`: Method to detemine render is complete
    - Type: string (valid values: networkIdle, InteractiveTime)
    - Default: networkIdle
- `skipFrameCount`: Skip first n framces with same id as init frame, only valid with idleType=networkIdle (Use on page with protection like CloudFlare)
    - Type: Int
    - Default: 0

### Example
See usage example at [manualTests](manualTests/render/main.go)

#### Build Example / Test
```bash
go build -o render.test
```

#### Run Example / Test
```
Usage of ./render.test:
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