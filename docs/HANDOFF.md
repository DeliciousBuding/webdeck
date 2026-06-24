# StarRailCopilot-Web 交接文档

> 生成时间: 2026-06-24 | 给下一个 agent 的完整状态和任务

---

## 一、项目目标

将 SRC（StarRailCopilot Python 自动化脚本）从 **Android 模拟器 + ADB** 改为 **浏览器 + CDP**，通过云游戏 `sr.mihoyo.com/cloud/` 运行崩坏：星穹铁道。

**核心原则：对 SRC 本体做最小改动，只替换设备抽象层。**

---

## 二、最终架构（已确认）

```
SRC 调度器 (Python, 不改动)
  └── Device 抽象层
        ├── adb.py                 # 原版 Android / 模拟器
        └── web_gateway.py         # 新增 (替代旧 playwright_device.py)

web_gateway.py  (HTTP client, ~100行)
  └── Gateway HTTP API

══════════════════════════════════════════════════════════════
Gateway (Go, 单文件 <20MB, port 8090)
  ├── Device interface (internal/device/interface.go)
  │     └── ChromeDevice 实现 (internal/device/chrome_device.go)
  ├── Server layer (internal/server/server.go)
  │     ├── /api/v1/*    稳定 API (SRC 控制平面)
  │     ├── /api/*       兼容别名 (向后兼容旧端点)
  │     └── /stream, /ws 人工调试平面
  ├── Browser layer (internal/browser/*)
  │     └── Chrome/CDP 操作
  └── Stream (internal/stream/hub.go)
        └── WebSocket Hub
══════════════════════════════════════════════════════════════

Chrome (headless, mobile UA, 1280×720)
  └── 云·星穹铁道 web 版
```

---

## 三、仓库

| 仓库 | 职责 | 位置 | 分支 |
|:-----|:-----|:-----|:-----|
| **star-rail-copilot-web** | SRC fork + web_gateway.py adapter | `D:\Code\Projects\star-rail-copilot-web` | `feat/playwright` |
| **src-web-gateway** | Go Gateway (Chrome/CDP, HTTP API, WebUI) | `D:\Code\Projects\src-web-gateway` | `master` |

**原则**：Gateway 代码不进 SRC repo，SRC 代码不进 Gateway repo。

---

## 四、已验证的技术方案（无需重试）

| 项目 | 方案 | 状态 |
|:-----|:-----|:----:|
| **截图** | CDP `Page.captureScreenshot` → JPEG Q75, 1280×720, ~30KB | ✅ |
| **点击** | `bring_to_front` → `window.focus()` → CDP `mousePressed`(80ms)→`mouseReleased` | ✅ |
| **滑动** | CDP `touchStart` → `touchMove`×10 → `touchEnd` | ✅ |
| **手机 UA** | `EmulateViewport` + `EmulateMobile` + `EmulateTouch` + Android UA | ✅ |
| **Cookie 持久化** | Playwright 兼容 JSON → `network.SetCookie` → 14 cookies | ✅ |
| **隐身** | `navigator.webdriver=false`, `plugins=5`, `window.chrome={}` | ✅ |
| **MJPEG 流** | JPEG via `multipart/x-mixed-replace`, 30 FPS | ✅ |
| **Docker 构建** | `python:3.10-slim-bookworm` + `av==12.3.0` | ✅ |

---

## 五、当前代码结构（已部分重构）

```
src-web-gateway/
├── cmd/gateway/
│   ├── main.go              # 入口: load config → create device → start server
│   └── frontend/index.html  # 嵌入式 WebUI
├── internal/
│   ├── device/
│   │   ├── interface.go     # ✅ Device 接口已定义 (Go interface)
│   │   └── chrome_device.go # ✅ ChromeDevice 实现
│   ├── browser/
│   │   ├── browser.go       # ✅ Browser struct, NewLocal, NewRemote, IsAlive
│   │   ├── cdp.go           # ✅ ScreenshotJPEG, Click, Swipe, Key
│   │   ├── cookies.go       # ✅ loadCookies (Playwright JSON format)
│   │   └── navigate.go      # ✅ Navigate, DismissHTML, InspectWhite
│   ├── server/
│   │   └── server.go        # ✅ 完整 HTTP 路由: v1 + compat + WS + MJPEG
│   └── stream/
│       └── hub.go           # ✅ WebSocket Hub
├── docs/
│   ├── ARCHITECTURE.md      # ✅ 系统架构文档
│   ├── API.md               # ✅ API 参考
│   └── progress/MASTER.md   # ✅ 进度追踪
├── go.mod / go.sum
└── Dockerfile / docker-compose.yml  # ⬜ 待写
```

---

## 六、编译状态

```bash
# 目前有 2 个编译错误：
# 1. internal/browser/cookies.go — 缺少 "context" import (已修)
# 2. internal/browser/browser.go — 未使用的 "log" import (已修)
# 修复后应该能编译
cd D:/Code/Projects/src-web-gateway
go build -o gw.exe ./cmd/gateway/
```

---

## 七、Device 接口 (已定义，不可改)

```go
// internal/device/interface.go
type Device interface {
    Info(ctx context.Context) (*DeviceInfo, error)
    Health(ctx context.Context) HealthStatus
    Screenshot(ctx context.Context, opts ScreenshotOptions) ([]byte, error)
    Tap(ctx context.Context, x, y int) error
    Swipe(ctx context.Context, x1, y1, x2, y2 int, durationMs int) error
    Key(ctx context.Context, key string) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Restart(ctx context.Context) error
    Reset(ctx context.Context) error
}
```

**WebUI、WS、MJPEG 只能通过 Device 接口操作设备，不能直接调 browser。**

---

## 八、API 设计（已冻结，不可改）

### v1 Stable API (SRC Control Plane)

```
GET  /api/v1/health              → {"ok":true, "state":"RUNNING", ...}
GET  /api/v1/device/info         → {"width":1280, "height":720, ...}
GET  /api/v1/device/screenshot   → image/jpeg
POST /api/v1/input/tap           → {"x":640, "y":360}
POST /api/v1/input/swipe         → {"x1":..., "y1":..., "x2":..., "y2":..., "duration_ms":300}
POST /api/v1/input/key           → {"key":"Escape"}
POST /api/v1/app/start           → 导航进游戏
POST /api/v1/app/restart         → 停止再启动
POST /api/v1/session/reset       → 重置 Chrome
```

### Compat Aliases (向后兼容，已实现)

```
GET /api/click?x=&y=            → POST /api/v1/input/tap
GET /api/swipe?x1=&y1=&x2=&y2=  → POST /api/v1/input/swipe
GET /api/screenshot              → GET  /api/v1/device/screenshot
GET /api/navigate                → POST /api/v1/app/start
GET /api/dismiss                 → JS dismiss HTML dialogs
GET /api/health                  → "ok" or state name
```

### Human Debug Plane

```
GET  /         → 嵌入式 WebUI
WS   /ws       → 命令通道
GET  /stream   → MJPEG 视频流
```

---

## 九、坐标契约

- 截图分辨率：**1280×720**
- 所有坐标：**screenshot pixel space**
- `/api/v1/device/info` 返回坐标元信息
- Python adapter 启动时必须验证 `width==1280 && height==720`

---

## 十、剩余任务（按优先级）

### P0: 编译修复 (5min)

```bash
cd D:/Code/Projects/src-web-gateway
# 修复 cookies.go 和 browser.go 的 import 问题
go build -o src-web-gateway.exe ./cmd/gateway/
```

### P1: Docker + 端到端验证 (30min)

1. 写 `Dockerfile`（生产模式：Gateway 容器内启动 Chromium 子进程）
2. 写 `docker-compose.yml`（调试模式：Chrome 容器 + Gateway 容器）
3. `docker compose up --build`
4. 验证：screenshot 有画面、click 有响应、health 返回 RUNNING

### P2: SRC adapter (20min)

1. 在 `D:\Code\Projects\star-rail-copilot-web`：
   - 重命名 `module/device/method/playwright_device.py` → `web_gateway.py`
   - 删除 playwight 直接调用，改为 HTTP client 调 Gateway
   - 保留方法签名：`screenshot()`, `click()`, `swipe()`, 与 `adb.py` 一致
   - 不改 `screenshot.py` 和 `control.py` 的现有集成点
2. 验证：SRC 能通过 web_gateway.py 截图并识别按钮

### P3: 恢复机制 (后续)

Gateway 检测到以下情况自动触发 Reset：
- Chrome 进程消失
- CDP 断开 > 10s
- 截图连续相同 > 60s
- 云游戏断连弹窗

---

## 十一、SRC 侧 web_gateway.py 目标接口

```python
class WebGateway:
    """SRC Device method — HTTP client to Gateway. Same interface as adb.py."""

    def __init__(self, config):
        self.gateway_url = config.GatewayUrl  # e.g. http://gateway:8090

    def screenshot(self) -> np.ndarray:
        """GET /api/v1/device/screenshot → JPEG decode → BGR ndarray 720×1280×3"""

    def click(self, x: int, y: int):
        """POST /api/v1/input/tap"""

    def swipe(self, p1: tuple, p2: tuple, duration: float = 0.3):
        """POST /api/v1/input/swipe"""

    def app_start(self):
        """POST /api/v1/app/start"""

    def app_is_running(self) -> bool:
        """GET /api/v1/health → check state == RUNNING"""
```

**SRC 侧绝对不能出现**：`cdp`, `chrome`, `playwright`, `cloud_game`, `mjpeg`, `websocket`

---

## 十二、关键架构决策（不可推翻）

1. **Gateway 是 Virtual Device Runtime，不是 Playwright wrapper**
2. **SRC 侧只新增一个 `web_gateway.py`，不改任务逻辑**
3. **WebUI/WS/MJPEG 只能通过 Device 接口操作设备**
4. **坐标系统固定 1280×720 screenshot pixel**
5. **API 分 v1 (稳定) 和 compat (别名) 两层**
6. **Docker 双模式：production (单容器子进程) 和 debug (双容器 --remote)**
7. **不要实现完整 ADB wire protocol，做 ADB shell 语义兼容即可**
8. **不要 if is_web / is_chrome / is_cloud_game 污染 SRC 主逻辑**

---

## 十三、文件索引

| 文件 | 用途 | 状态 |
|:-----|:-----|:----:|
| `docs/ARCHITECTURE.md` | 完整系统架构说明 | ✅ |
| `docs/API.md` | API 参考 | ✅ |
| `docs/progress/MASTER.md` | 进度追踪 | ✅ |
| `internal/device/interface.go` | Device Go 接口 | ✅ |
| `internal/device/chrome_device.go` | ChromeDevice 实现 | ✅ |
| `internal/server/server.go` | HTTP 路由 + WS + MJPEG | ✅ |
| `internal/browser/browser.go` | Browser 结构体 | ✅ |
| `internal/browser/cdp.go` | CDP 命令封装 | ✅ |
| `cmd/gateway/main.go` | 入口 (精简版) | ✅ |
| `cmd/gateway/frontend/index.html` | 嵌入式 WebUI | ✅ |
| `Dockerfile` | 待写 | ⬜ |
| `docker-compose.yml` | 待写 | ⬜ |

---

## 十四、下一个 agent 应该做的第一件事

```bash
# 1. 读文档
cat docs/ARCHITECTURE.md
cat docs/API.md

# 2. 修复编译
cd D:/Code/Projects/src-web-gateway
go build -o gw.exe ./cmd/gateway/
# 预期：2 个 import 错误，修完即可编译通过

# 3. 写 Dockerfile + docker-compose.yml
# 4. docker compose up --build
# 5. 验证 /api/v1/health 返回 RUNNING
# 6. 写 SRC web_gateway.py
```

---

## 十五、Git 策略

- `star-rail-copilot-web`: `feat/web-gateway-device` 分支，只增 `web_gateway.py` + 集成修改
- `src-web-gateway`: `master` 分支，正常开发
- 两边独立 commit，不交叉
