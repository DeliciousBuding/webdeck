# SRC-Web Gateway — 云·星穹铁道 浏览器控制工具

Go 单文件二进制，自包含前后端。通过 Chrome CDP 控制云游戏网页。

## 对比

| | Python live_preview.py | Go webdeck |
|:--|:--|:--|
| 二进制 | Docker 4GB | 单文件 16MB |
| 帧率 | 8 FPS PNG | 30 FPS JPEG Q75 |
| 单帧 | ~500KB | ~30KB |
| 点击 | HTTP 250ms | WebSocket <10ms |
| 启动 | 15s | 2s |
| 前端 | Python render | embed 内嵌 |
| 依赖 | Python + Playwright | 仅系统 Chrome |

## 用法

```bash
# 构建
go build -o webdeck.exe .

# 运行（需要 cloud_auth.json）
./webdeck.exe --auth cloud_auth.json --port 8090 --fps 30

# 打开
http://localhost:8090
```

## 命令行参数

| 参数 | 默认值 | 说明 |
|:--|:--|:--|
| `--auth` | `config/cloud_auth.json` | Playwright cookie JSON |
| `--port` | `8090` | HTTP 端口 |
| `--fps` | `30` | 目标帧率 |
| `--jpeg-quality` | `75` | JPEG 质量 (1-100) |

## API

| 端点 | 说明 |
|:--|:--|
| `GET /` | 嵌入式 HTML 前端 |
| `WS /ws` | WebSocket 命令 (click/swipe/key/dismiss) |
| `GET /stream` | MJPEG 视频流 |
| `GET /api/click?x=&y=` | HTTP 点击 |
| `GET /api/dismiss` | 关闭 HTML 弹窗 |
| `GET /api/screenshot` | 单帧 JPEG |
| `GET /api/health` | 健康检查 |

## 前端操作

- **点击** — 直接点画面 (CDP 可信事件)
- **滑动** — 点 Slide 按钮进入滑动模式
- **Dismiss** — 关闭 HTML 弹窗（用户协议/添加到桌面）
- **键盘** — 直接按键
