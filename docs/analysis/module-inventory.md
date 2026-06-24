# Module Inventory — src-web-gateway

## Summary

| Module | Responsibility | Dependencies | Files | Lines | Complexity | S.U.P.E.R Score |
|:-------|:---------------|:-------------|------:|------:|:-----------|:----------------|
| `cmd/gateway` | Entry point, flag parsing, wire-up | device, server | 2 | ~50 | Low | S🟢 U🟢 P🟢 E🟢 R🟡 |
| `internal/device` | Device abstraction interface + ChromeDevice | browser | 2 | ~160 | Medium | S🟢 U🟢 P🟢 E🟢 R🟢 |
| `internal/server` | HTTP routing, v1 API, compat aliases, WS, MJPEG | device, stream | 1 | ~300 | Medium | S🟡 U🟢 P🟡 E🟢 R🟡 |
| `internal/browser` | Chrome/CDP management (launch, screenshot, input) | chromedp | 4 | ~230 | Medium | S🟡 U🟢 P🟡 E🟡 R🟢 |
| `internal/stream` | WebSocket hub for MJPEG frame broadcast | gorilla/websocket | 1 | ~80 | Low | S🟢 U🟢 P🟢 E🟢 R🟢 |

## Module Details

### cmd/gateway
- **Path**: `cmd/gateway/`
- **Responsibility**: CLI entry point. Parses flags, creates ChromeDevice, constructs Server, delegates to `server.Start()`.
- **Public API**: `main()` — the single entry point
- **Internal Dependencies**: `internal/device`, `internal/server`
- **External Dependencies**: None (only Go stdlib + project internal)
- **Complexity Rating**: Low
- **S.U.P.E.R Assessment**:
  - **S**: ✅ Single purpose — wire-up only
  - **U**: ✅ Unidirectional: flags → device → server
  - **P**: ✅ All I/O through Device interface
  - **E**: ✅ All env-specific config via CLI flags
  - **R**: 🟡 Replaceable with other main.go patterns, but embed.FS coupling

### internal/device
- **Path**: `internal/device/`
- **Responsibility**: Define the Device interface contract and provide ChromeDevice (CDP) implementation.
- **Public API**: `Device` interface (Info, Health, Screenshot, Tap, Swipe, Key, Start, Stop, Restart, Reset), `ChromeDevice`, `NewChrome()`
- **Internal Dependencies**: `internal/browser`
- **External Dependencies**: None
- **Complexity Rating**: Medium
- **S.U.P.E.R Assessment**:
  - **S**: ✅ Single purpose — device abstraction
  - **U**: ✅ Dependencies flow inward (device → browser, never reverse)
  - **P**: ✅ Interface contract defined before implementation — future backends (WebRTC, Moonlight) implement same interface
  - **E**: ✅ No hardcoded paths or platform assumptions
  - **R**: ✅ Full replacement test passes — swap ChromeDevice for another impl, nothing else changes

### internal/server
- **Path**: `internal/server/server.go`
- **Responsibility**: HTTP server with v1 stable API, compat aliases, WebSocket, and MJPEG streaming. All browser operations go through the Device interface.
- **Public API**: `Server`, `New(dev Device, cfg Config)`, `Start() error`
- **Internal Dependencies**: `internal/device`, `internal/stream`
- **External Dependencies**: None (embed, net/http, encoding/json)
- **Complexity Rating**: Medium
- **S.U.P.E.R Assessment**:
  - **S**: 🟡 HTTP routing + capture loop — two responsibilities in one file
  - **U**: ✅ Data flows: request → Device interface → response
  - **P**: 🟡 Compat dismiss handler type-asserts to ChromeDevice (bypasses Device interface)
  - **E**: ✅ All config via Config struct, no hardcoded paths
  - **R**: 🟡 Can swap Device backend, but compat dismiss currently couples to ChromeDevice

### internal/browser
- **Path**: `internal/browser/`
- **Responsibility**: Chrome lifecycle management via chromedp. Launch/connect, screenshot, input dispatch, cookie loading, navigation, stealth injection.
- **Public API**: `Browser` struct, `NewLocal()`, `NewRemote()`, `ScreenshotJPEG()`, `Click()`, `Swipe()`, `Key()`, `Navigate()`, `DismissHTML()`, `InspectWhite()`, `IsAlive()`, `Close()`
- **Internal Dependencies**: None (only chromedp)
- **External Dependencies**: chromedp/cdproto, chromedp/chromedp
- **Complexity Rating**: Medium
- **S.U.P.E.R Assessment**:
  - **S**: 🟡 CDP commands + navigation + cookie loading — could split into cdp.go, navigate.go, cookies.go (already partially done)
  - **U**: ✅ All calls flow Browser → chromedp → Chrome
  - **P**: 🟡 Hardcoded mobile UA string, hardcoded stealth JS, hardcoded game URL
  - **E**: 🟡 `chromedp.DefaultExecAllocatorOptions` — platform-specific Chrome flags
  - **R**: ✅ Future CDP library could replace chromedp without changing device/server

### internal/stream
- **Path**: `internal/stream/hub.go`
- **Responsibility**: WebSocket hub for broadcasting JPEG frames to connected WebUI clients.
- **Public API**: `Hub`, `NewHub()`, `SetFrame()`, `HandleWS()`
- **Internal Dependencies**: None
- **External Dependencies**: gorilla/websocket
- **Complexity Rating**: Low
- **S.U.P.E.R Assessment**:
  - **S**: ✅ Single purpose — WebSocket frame broadcast
  - **U**: ✅ One-directional: SetFrame → broadcast to clients
  - **P**: ✅ Cmd struct is JSON-serializable
  - **E**: ✅ No hardcoded paths or platform assumptions
  - **R**: ✅ Replaceable with any WebSocket library
