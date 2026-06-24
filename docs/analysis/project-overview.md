# Project Overview — src-web-gateway

## Preliminary Direction
Replace the Python Playwright-based cloud gaming bridge (`live_preview.py`) with a Go single-binary Gateway that manages Chrome via CDP, exposing a stable HTTP API for the SRC scheduler.

## Current Architecture

```
SRC Scheduler (Python) ──HTTP──► Gateway (Go, :8090) ──CDP──► Chrome (headless, 1280×720)
                                       │
                                  Embedded WebUI (MJPEG + WebSocket)
```

- **SRC (upstream)**: Talks to Gateway via `/api/v1/*` HTTP endpoints — same interface as ADB device methods
- **Gateway**: Go single binary (~16 MB), manages Chrome lifecycle, captures frames, dispatches input
- **Chrome**: Headless, mobile UA (Pixel 7, Android 13), 1280×720 fixed viewport, cloud game at `sr.mihoyo.com/cloud/`

## Technology Stack

| Layer        | Current              |
|:-------------|:---------------------|
| Language     | Go 1.22              |
| CDP Library  | chromedp v0.11       |
| WebSocket    | gorilla/websocket    |
| Frontend     | Embedded HTML/CSS/JS |
| Deployment   | Docker (Debian Bookworm + Chromium) |
| Build        | `go build -o src-web-gateway.exe ./cmd/gateway/` |

## Entry Points

| Entry | Description |
|:------|:------------|
| `GET /` | Embedded WebUI (remote desktop) |
| `WS /ws` | WebSocket command channel (click/swipe/key/dismiss) |
| `GET /stream` | MJPEG video stream |
| `GET /api/v1/health` | Health check with state machine status |
| `GET /api/v1/device/info` | Device metadata (resolution, backend, readiness) |
| `GET /api/v1/device/screenshot` | JPEG screenshot at 1280×720 |
| `POST /api/v1/input/tap` | CDP trusted click at (x, y) |
| `POST /api/v1/input/swipe` | CDP touch swipe gesture |
| `POST /api/v1/input/key` | Keyboard event |
| `POST /api/v1/app/start` | Navigate to cloud game + dismiss dialogs |
| `POST /api/v1/app/stop` | Stop game session |
| `POST /api/v1/app/restart` | Stop then start |
| `POST /api/v1/session/reset` | Kill Chrome + restart |

## Build & Run

```bash
go build -o src-web-gateway ./cmd/gateway/

# Standalone (launches Chrome subprocess):
./src-web-gateway --auth cloud_auth.json --port 8090 --fps 30

# Debug (connect to existing Chrome):
./src-web-gateway --remote ws://127.0.0.1:9222 --auth cloud_auth.json
```

## Testing Baseline
No automated tests yet. Verification is manual:
- `go vet ./...` — static analysis
- `go build` — compilation check
- End-to-end: Docker Compose with Chrome container

## Project Governance Baseline
- No AGENTS.md or CLAUDE.md in this repo
- No platform rule files
- docs/HANDOFF.md: handoff notes from initial development session

## External Integrations
- **Chrome DevTools Protocol**: via chromedp — page capture, input dispatch, cookie management
- **Cloud Game**: `sr.mihoyo.com/cloud/` (MiHoYo cloud gaming service)
- **SRC Scheduler**: Communicates over HTTP via `/api/v1/*` endpoints
