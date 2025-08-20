# Changelog

## [0.12.0] - 2025-08-20

### Changed

- **Breaking** Changed `RendererOption` and `PdfOption` structure, will not be compatible with previous version ([#51](https://github.com/liuminhaw/renderer/pull/51)) (Min-Haw, Liu)
- **Breaking** `SkipFrameCount` removed from `RendererOption` (now `RendererConf`) sturcture field ([#51](https://github.com/liuminhaw/renderer/pull/51)) (Min-Haw, Liu)
- Change implementation of network idle wait type by tracking in-flight requests which previous use frame counting approach ([#51](https://github.com/liuminhaw/renderer/pull/51)) (Min-Haw, Liu)

## [0.11.0] - 2025-05-01

### Added

- Introduce new `idleType` option `auto`, enabled by default ([#46](https://github.com/liuminhaw/renderer/pull/46)) (Min-Haw, Liu)
- Add render page and PDF usage examples to GoDoc ([#49](https://github.com/liuminhaw/renderer/pull/49)) (Min-Haw, Liu)

## [0.10.0] - 2025-01-28

### Changed

- **Breaking** Update renderer usage from using context to using options parameter ([#42](https://github.com/liuminhaw/renderer/pull/42)) (Min-Haw, Liu)
- Update logging with slog package and enable to use custom logger for slog ([#42](https://github.com/liuminhaw/renderer/pull/42)) (Min-Haw, Liu)

### Added

- Add User-Agent option to set custom user-agent value when using automated browser for rendering ([#43](https://github.com/liuminhaw/renderer/pull/43)) (Min-Haw, Liu)
- Set websocket url read timeout equal to the renderer timeout option ([#44](https://github.com/liuminhaw/renderer/pull/44)) (Min-Haw, Liu)

## [0.9.1] - 2024-12-27

### Changed

- Update dependencies for fixing message `ERROR: could not unmarshal event: parse error` when rendering page ([#37](https://github.com/liuminhaw/renderer/pull/37)) (Min-Haw, Liu)

## [0.9.0] - 2024-05-18

### Added

- Add context to show chromium execution message for debugging ([#34](https://github.com/liuminhaw/renderer/pull/34)) (Min-Haw, Liu)

## [0.8.0] - 2024-05-01

### Added

- Add context to run chrome in container environment (eg. docker / lambda) ([#30](https://github.com/liuminhaw/renderer/pull/30)) (Min-Haw, Liu)

### Removed

- **Breaking** Remove single-process from context option ([#30](https://github.com/liuminhaw/renderer/pull/30)) (Min-Haw, Liu)
- **Breaking** Remove no sandbox from context option ([#30](https://github.com/liuminhaw/renderer/pull/30)) (Min-Haw, Liu)

## [0.7.0] - 2024-03-07

### Added

- Add context to run chrome in single-process mode ([#28](https://github.com/liuminhaw/renderer/pull/28))

### Fixed

- Fix timeout default to 0 but not 30 seconds if not explicitly set ([#27](https://github.com/liuminhaw/renderer/pull/27))

## [0.6.0] - 2023-12-25

### Changed

- **Breaking:** New BrowserContext separated from RendererContext and PdfContext ([#24](https://github.com/liuminhaw/renderer/pull/24))

### Added

- Debug mode to print out debugging message ([#24](https://github.com/liuminhaw/renderer/pull/24))
- Add context to manually set chrome / chromium executable path ([#22](https://github.com/liuminhaw/renderer/pull/22))
- Add context to disable sandbox option from chromedp ([#23](https://github.com/liuminhaw/renderer/pull/23))

## [0.5.0] - 2023-10-11

### Changed

- Modify license from GNU General Public License v3.0 to MIT License ([#18](https://github.com/liuminhaw/renderer/pull/18))
- Upgrade chromedp from 0.8.6 to 0.9.2  ([#16](https://github.com/liuminhaw/renderer/pull/16))

### Added

- Add pdf renderer for site ([#17](https://github.com/liuminhaw/renderer/pull/17))

## [0.4.0] - 2023-09-27

### Changed

- **Breaking:** headless mode default set to `false`
- **Breaking:** imageLoad option default set to `false`
- Modify renderer context to use custom type ([#8](https://github.com/liuminhaw/renderer/pull/8))

_:seedling: Initial release._

[0.12.0]: https://github.com/liuminhaw/renderer/releases/tag/v0.12.0

[0.11.0]: https://github.com/liuminhaw/renderer/releases/tag/v0.11.0

[0.10.0]: https://github.com/liuminhaw/renderer/releases/tag/v0.10.0

[0.9.1]: https://github.com/liuminhaw/renderer/releases/tag/v0.9.1

[0.9.0]: https://github.com/liuminhaw/renderer/releases/tag/v0.9.0

[0.8.0]: https://github.com/liuminhaw/renderer/releases/tag/v0.8.0

[0.7.0]: https://github.com/liuminhaw/renderer/releases/tag/v0.7.0

[0.6.0]: https://github.com/liuminhaw/renderer/releases/tag/v0.6.0

[0.5.0]: https://github.com/liuminhaw/renderer/releases/tag/v0.5.0

[0.4.0]: https://github.com/liuminhaw/renderer/releases/tag/v0.4.0
