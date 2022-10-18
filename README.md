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

### Example
See usage example at [manualTests](manualTests/render/main.go)