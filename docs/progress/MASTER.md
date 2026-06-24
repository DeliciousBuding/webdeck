# MASTER.md — StarRailCopilot-Web 重构计划

> **核心理念**：SRC 侧把 Gateway 当设备；Gateway 侧把 Chrome 当显示和输入后端；WebUI 只是远程桌面调试器。

---

## 一、架构分层

```
SRC 原版任务逻辑 (100% 不动)
  └── Device 抽象层
        ├── adb.py                 # 原版 Android / 模拟器
        └── web_gateway.py         # 新增 (原 playwright_device.py 改名)

web_gateway.py (Python HTTP client)
  └── Gateway API (HTTP)

═══════════════════════════════════════════════════╗
║  Gateway (Go, 单文件, port 8090)                 ║
║                                                   ║
║  Device Interface ──→ ChromeDevice (CDP 实现)     ║
║  v1 Stable API ────→ SRC Control Plane            ║
║  Compat Aliases ───→ 向后兼容旧端点               ║
║  WebUI ────────────→ Human Debug Plane            ║
║  Health State ─────→ 状态机 + 恢复                ║
║                                                   ║
║  ┌─ production: Gateway 启动 Chromium 子进程      ║
║  └─ debug:      Gateway 连接外部 Chrome (--remote)║
╚═══════════════════════════════════════════════════╝

Chrome (headless, mobile UA, 1280×720)
  └── 云·星穹铁道 web 版
```

## 二、仓库规划

| 仓库 | 职责 | 位置 |
|:-----|:-----|:-----|
| **star-rail-copilot-web** | SRC fork + `web_gateway.py` adapter | `D:\Code\Projects\star-rail-copilot-web` |
| **src-web-gateway** | Go Gateway (Chrome/CDP, HTTP API, WebUI) | `D:\Code\Projects\src-web-gateway` |

**原则**：Gateway 代码不进 SRC repo，SRC 代码不进 Gateway repo。

### star-rail-copilot-web 改动清单

```
仅改/增 3 文件:
  ➕ module/device/method/web_gateway.py    (重命名自 playwright_device.py)
  ✏️ module/device/screenshot.py            (+1 行 import)
  ✏️ module/device/control.py               (+2 行 dispatch)
  ✏️ module/config/                          (+1 配置项 GatewayUrl)

不动:
  ✅ module/base/
  ✅ module/tasks/
  ✅ module/ocr/
  ✅ module/combat/
  ✅ module/assets/
  ✅ module/alas.py
```

### src-web-gateway 目标结构

```
src-web-gateway/
├── cmd/gateway/
│   ├── main.go
│   └── frontend/index.html       # 嵌入式 WebUI
├── internal/
│   ├── device/
│   │   ├── interface.go          # Go Device 接口定义
│   │   └── chrome.go             # ChromeDevice 实现 (CDP)
│   ├── server/
│   │   ├── server.go             # HTTP 路由注册
│   │   ├── api_v1.go             # /api/v1/* 稳定 API
│   │   ├── api_compat.go         # /api/click 等兼容别名
│   │   └── webui.go              # MJPEG + WebSocket
│   ├── protocol/
│   │   └── types.go              # DTOs / request-response types
│   └── stream/
│       └── hub.go                # WebSocket Hub
├── Dockerfile                    # 生产模式 (单容器)
├── docker-compose.yml            # 调试模式 (双容器)
└── docs/
    ├── ARCHITECTURE.md
    ├── API.md
    └── progress/MASTER.md
```

## 三、Device 接口 (Go)

```go
type Device interface {
    Info(ctx context.Context) (*DeviceInfo, error)
    Screenshot(ctx context.Context, opts ScreenshotOptions) ([]byte, error)
    Tap(ctx context.Context, x, y int) error
    Swipe(ctx context.Context, x1, y1, x2, y2 int, durationMs int) error
    Key(ctx context.Context, key string) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Restart(ctx context.Context) error
    Reset(ctx context.Context) error
    Health(ctx context.Context) HealthStatus
}

type ChromeDevice struct { ... } // CDP 实现
```

## 四、v1 Stable API

### SRC Control Plane

```
GET  /api/v1/health              → {"ok":true, "state":"RUNNING", ...}
GET  /api/v1/device/info         → {"width":1280, "height":720, ...}
GET  /api/v1/device/screenshot   → image/jpeg (Q75, 1280×720)
POST /api/v1/input/tap           → {"x":640, "y":360}
POST /api/v1/input/swipe         → {"x1":100,"y1":200,"x2":300,"y2":400,"duration_ms":300}
POST /api/v1/input/key           → {"key":"Escape"}
POST /api/v1/app/start           → 导航进游戏
POST /api/v1/app/stop            → 停止
POST /api/v1/session/reset       → 重置 Chrome + 重新登录
```

### Compat Aliases (向后兼容)

```
GET  /api/click?x=&y=        → POST /api/v1/input/tap
GET  /api/swipe?x1=&y1=&x2=&y2= → POST /api/v1/input/swipe
GET  /api/screenshot           → GET /api/v1/device/screenshot
GET  /api/navigate             → POST /api/v1/app/start
GET  /api/dismiss              → JS dismiss HTML dialogs
```

### Human Debug Plane

```
GET  /                          → 嵌入式 WebUI
WS   /ws                        → 命令通道 (click/swipe/key/dismiss)
GET  /stream                    → MJPEG 视频流
GET  /api/v1/debug/inspect      → DOM 检查
```

## 五、坐标系统

- **固定分辨率**：1280×720
- **坐标系**：screenshot pixel
- **端到端保证**：screenshot 返回的像素坐标 = input/tap 使用的坐标
- `/api/v1/device/info` 必须返回坐标元信息，Python adapter 启动时验证

## 六、Chrome 模式

### 生产模式（默认）

```
Gateway 容器
  ├── src-web-gateway
  └── chromium --remote-debugging-port=9222 (子进程)

Gateway 管理 Chrome 生命周期（启动/重启/销毁）
```

### 调试模式

```
Chrome 容器 (port 9222)
Gateway 容器 (--remote ws://chrome:9222)

双容器，可单独看 Chrome 日志
```

## 七、Health 状态机

```
INIT → CHROME_STARTING → PAGE_LOADING → GAME_READY → RUNNING
                                                          ↓
                                                     DEGRADED
                                                          ↓
                                                     RECOVERING
                                                          ↓
                                                       RUNNING
```

`/api/v1/health` 返回：

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

## 八、执行计划

| Phase | Task | Est. |
|:------|:-----|:----:|
| **P1** | Gateway 目录重构 + Device 接口 + ChromeDevice | 30min |
| **P2** | v1 API + compat aliases + WebUI 通过 Device interface | 30min |
| **P3** | Docker 生产/调试双模式 | 15min |
| **P4** | SRC `web_gateway.py` adapter (重命名 + HTTP client) | 20min |
| **P5** | 端到端验证: 截图→点击→截图对比 | 15min |

## 九、SRC 侧 web_gateway.py 接口

```python
class WebGateway:
    """SRC Device method — mimics adb.py's interface."""

    def screenshot(self) -> np.ndarray:
        """Return 720×1280×3 BGR image."""

    def click(self, x: int, y: int):
        """Tap at screenshot coordinates."""

    def swipe(self, p1: tuple, p2: tuple, duration: float = 0.3):
        """Swipe from p1 to p2."""

    def long_click(self, x: int, y: int, duration: float = 1.0):
        """Long press."""

    def app_start(self):
        """Navigate to cloud game + enter."""

    def app_stop(self):
        """Stop game session."""

    def app_restart(self):
        """Restart game."""

    def app_is_running(self) -> bool:
        """Check if game video is playing."""

    def release(self):
        """Clean up."""
```

## 十、Git 策略

```bash
# SRC repo
git checkout -b feat/web-gateway-device
# Commits:
#   1. add module/device/method/web_gateway.py
#   2. integrate into screenshot.py + control.py
#   3. add config option
#   4. docs

# Gateway repo
# 主分支开发，Docker 驱动
```

## 十一、恢复机制

Gateway 检测到以下情况自动触发 Reset：

| 条件 | 动作 |
|:-----|:-----|
| Chrome 进程消失 | 重启 Chrome |
| CDP 断开 > 10s | 重连 |
| 截图连续相同 > 60s | 刷新页面 |
| 云游戏断连弹窗 | Dismiss + 重新进入 |
| Cookie 过期 | 标记 DEGRADED，等人工 |

## 十二、MVP 最小闭环

```
1. Gateway 启动 → Chrome 启动 → 导航云游戏 → 等待 video → RUNNING
2. SRC web_gateway.py 调 /screenshot → 返回 JPEG → 转 np.ndarray
3. SRC 模板匹配 → 找到按钮坐标 (x, y)
4. SRC web_gateway.py 调 /input/tap → Gateway CDP click → 游戏响应
5. 下一帧截图验证变化
```

## Adaptive Control

| Metric | Value |
|:-------|:------|
| Phases | 5 |
| Drift Score | 0 |
| Annotate Threshold | 1 |
| Replan Threshold | 2 |
| Rescope Threshold | 3 |
