# Gateway API Reference

## v1 Stable API (SRC Control Plane)

### Health & Info

#### `GET /api/v1/health`
```json
{
  "ok": true,
  "state": "RUNNING",
  "chrome_alive": true,
  "cdp_connected": true,
  "page_ready": true,
  "last_frame_age_ms": 120,
  "last_input_age_ms": 3400,
  "recover_count": 1
}
```

#### `GET /api/v1/device/info`
```json
{
  "width": 1280,
  "height": 720,
  "dpr": 1,
  "orientation": "landscape",
  "screenshot_format": "jpeg",
  "input_coordinate": "screenshot_pixel",
  "backend": "chrome-cdp",
  "ready": true
}
```

### Screenshot

#### `GET /api/v1/device/screenshot`
- Query: `?format=jpeg&quality=75`
- Response: `image/jpeg` binary (1280×720)

### Input

#### `POST /api/v1/input/tap`
```json
{"x": 640, "y": 360}
```

#### `POST /api/v1/input/swipe`
```json
{"x1": 100, "y1": 200, "x2": 300, "y2": 400, "duration_ms": 300}
```

#### `POST /api/v1/input/key`
```json
{"key": "Escape"}
```

### App Control

#### `POST /api/v1/app/start`
Navigates to cloud game, dismisses dialogs, waits for video.

#### `POST /api/v1/app/stop`
Stops game session.

#### `POST /api/v1/app/restart`
Stops then starts.

### Session

#### `POST /api/v1/session/reset`
Kills Chrome, restarts, re-navigates, re-enters game.

---

## Compat Aliases

```
GET /api/click?x=&y=           → POST /api/v1/input/tap
GET /api/swipe?x1=&y1=&x2=&y2= → POST /api/v1/input/swipe
GET /api/screenshot             → GET  /api/v1/device/screenshot
GET /api/navigate               → POST /api/v1/app/start
GET /api/dismiss                → JS dismiss HTML dialogs
```

---

## Human Debug Plane

```
GET  /              → Embedded WebUI (remote desktop)
WS   /ws            → Command channel
GET  /stream        → MJPEG video stream
GET  /api/v1/debug/inspect → DOM background inspection
```

---

## Coordinate Contract

- Screenshot resolution: **1280×720**
- All tap/swipe coordinates: **screenshot pixel space**
- `/api/v1/device/info` returns coordinate metadata
- Client should verify `width==1280 && height==720` at startup
