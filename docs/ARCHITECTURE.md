# StarRailCopilot-Web 架构文档

## 一、目标

将 SRC（StarRailCopilot Python 自动化脚本）从 **Android 模拟器 + ADB** 改为 **浏览器 + CDP**，通过云游戏（`sr.mihoyo.com/cloud/`）运行崩坏：星穹铁道。

**核心原则：对 SRC 本体做最小改动，只替换设备抽象层。**

## 二、三层架构

```
┌──────────────────────────────────────────────────┐
│  SRC 调度器 (Python, 不改动)                      │
│  module/device/method/playwright_device.py        │
│                                                   │
│  截图: GET /api/screenshot  → np.ndarray (BGR)   │
│  点击: GET /api/click?x=&y=                       │
│  滑动: GET /api/swipe?x1=&y1=&x2=&y2=             │
│  启动: POST /api/navigate                         │
└──────────────────┬───────────────────────────────┘
                   │ HTTP (docker compose 内部网络)
┌──────────────────▼───────────────────────────────┐
│  Gateway (Go, 16MB 单文件, port 8090)             │
│                                                   │
│  • 管理 Chrome 生命周期（启动/重启/销毁）         │
│  • HTTP API → SRC 调用                            │
│  • WebSocket + MJPEG → WebUI 远程桌面             │
│  • CDP 客户端 → 连接 Chrome                       │
└──────────────────┬───────────────────────────────┘
                   │ CDP (ws://chrome:9222)
┌──────────────────▼───────────────────────────────┐
│  Chrome (headless, Docker 容器)                   │
│                                                   │
│  • 手机 UA (Android Pixel 7)                      │
│  • 手机触摸模拟 (EmulateMobile + EmulateTouch)     │
│  • 1280×720 固定分辨率（匹配 SRC 模板）           │
│  • 隐身脚本 (navigator.webdriver=false)           │
│  • 加载 cloud_auth.json cookie                    │
└──────────────────────────────────────────────────┘
```

## 三、和原版 SRC 的对照

| 原版 SRC | 我们 |
|:---------|:-----|
| Android 模拟器 | Chrome + 云游戏网页 |
| ADB (`adb shell input tap`) | CDP `Input.dispatchMouseEvent` (可信事件) |
| `adb shell screencap` | `Page.captureScreenshot` (JPEG Q75, ~30KB) |
| `scrcpy` / `minitouch` | 手机 UA + 触摸模拟 |
| 多设备并行 | Docker Compose 多实例 |

## 四、已验证的技术方案

| 项目 | 方案 | 验证 |
|:-----|:-----|:----:|
| **截图** | CDP `Page.captureScreenshot` → JPEG Q75, 1280×720 | ✅ |
| **点击** | `bring_to_front` → `window.focus()` → CDP `mousePressed`(80ms)→`mouseReleased` | ✅ |
| **滑动** | CDP `touchStart` → `touchMove`×10 → `touchEnd` | ✅ |
| **手机 UA** | `EmulateViewport` + `EmulateMobile` + `EmulateTouch` + Android UA string | ✅ |
| **Cookie 持久化** | Playwright 兼容 JSON → `network.SetCookie` → 14 cookies | ✅ |
| **隐身** | `navigator.webdriver=false`, `plugins=5`, `window.chrome={}` | ✅ |
| **MJPEG 流** | JPEG 帧 via `multipart/x-mixed-replace`, 30 FPS | ✅ |

## 五、仓库

| 仓库 | 用途 | 位置 |
|:-----|:-----|:-----|
| `star-rail-copilot-web` | SRC fork + Playwright device | `D:\Code\Projects\star-rail-copilot-web` — `feat/playwright` |
| `src-web-gateway` | Go Gateway | `D:\Code\Projects\src-web-gateway` |

## 六、Gateway 代码结构

```
src-web-gateway/
├── cmd/gateway/
│   ├── main.go              # 入口 (flag解析, HTTP路由, 截图循环)
│   └── frontend/index.html  # 嵌入式 WebUI
├── internal/
│   ├── browser/browser.go   # Chrome/CDP 管理
│   └── stream/stream.go     # WebSocket Hub
├── go.mod / go.sum
├── Dockerfile               # Gateway 容器
└── docker-compose.yml       # Chrome + Gateway
```

## 七、Gateway API

| 端点 | 方法 | 说明 |
|:-----|:-----|:-----|
| `/api/health` | GET | 健康检查 |
| `/api/screenshot` | GET | JPEG 截图 (1280×720) |
| `/api/click?x=&y=` | GET | CDP 可信点击 |
| `/api/swipe?x1=&y1=&x2=&y2=` | GET | CDP 触摸滑动 |
| `/api/dismiss` | GET | 关闭 HTML 弹窗 |
| `/api/navigate` | POST | 导航到云游戏 |
| `/api/inspect` | GET | 调试：列出 DOM 白色背景元素 |
| `/stream` | GET | MJPEG 视频流 |
| `/ws` | WebSocket | 命令通道 (click/swipe/key/dismiss) |
| `/` | GET | 嵌入式 WebUI 远程桌面 |

## 八、已完成 vs 待完成

| # | 任务 | 状态 | 说明 |
|:--|:-----|:----:|:-----|
| 1 | Python Playwright device 实现 | ✅ | `playwright_device.py`，SRC mixin 模式 |
| 2 | CDP 点击方案验证 | ✅ | `bring_to_front + focus + mousePressed/released` |
| 3 | Cookie 加载 + 隐身 | ✅ | 14 cookies, stealth JS |
| 4 | Go Gateway 编译 | ✅ | 16MB 单文件, chromedp |
| 5 | WebUI 基础 | ✅ | MJPEG + WS + 坐标映射 |
| 6 | Gateway 启动测试 | ⬜ | 本地无 Chrome，阻塞 |
| 7 | Docker Compose | ⬜ | Chrome + Gateway 双容器 |
| 8 | 手机 UA + 触摸 | ⬜ | 编译通过，待实测 |
| 9 | SRC adapter (HTTP) | ⬜ | playwright_device.py 改 HTTP |
| 10 | 端到端闭环 | ⬜ | SRC → Gateway → Chrome → 自动化 |

## 九、阻塞 & 下一步

**当前阻塞**：Windows 本地无 Chrome，Gateway 启动失败。需在 Docker 中运行。

**下一步**：
```bash
# 1. 写 docker-compose.yml (Chrome + Gateway)
# 2. Gateway 用 --remote ws://chrome:9222 连接
# 3. docker compose up --build
# 4. 截图验证手机 UA 效果
# 5. SRC playwright_device.py 改为 HTTP 调 Gateway
```
