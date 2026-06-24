# webdeck Design System

> Virtual device runtime — single Go binary that turns Chrome into a game-controllable device.

## Principles

1. **Dark-first, always dark.** This is a gaming remote-desktop tool. Light mode is not a use case.
2. **Content is the game.** The video stream is the hero. Chrome should fade away — controls appear on-demand, status is glanceable.
3. **Precision over decoration.** Every pixel matters at 1280×720. Controls must be crisp, coordinates exact.
4. **No motion for motion's sake.** Transitions are 150ms max, snappy, and functional. Respect `prefers-reduced-motion`.
5. **Readable at a distance.** Status indicators and FPS counters must be legible on a secondary monitor.

---

## Color Tokens

Dark theme. Green accent for go/play/active states. Red for danger/stop.

| Token | Value | Usage |
|-------|-------|-------|
| `--gray-0` | `#f8fafc` | Primary text on dark |
| `--gray-1` | `#94a3b8` | Secondary text |
| `--gray-2` | `#64748b` | Tertiary text, disabled |
| `--gray-3` | `#334155` | Borders |
| `--gray-4` | `#1e293b` | Raised surfaces |
| `--gray-5` | `#0f172a` | Default background |
| `--gray-6` | `#020617` | Deepest background (stream area) |
| `--accent` | `#22c55e` | Primary action, connected, live |
| `--accent-dim` | `#166534` | Accent backgrounds, pressed |
| `--danger` | `#ef4444` | Error, disconnect, destructive |
| `--danger-dim` | `#7f1d1d` | Danger backgrounds |

---

## Typography

**Font**: Inter, system-ui fallback. Single weight scale.

| Token | Size / Line | Weight | Usage |
|-------|-------------|--------|-------|
| `--text-xs` | 11px / 16px | 400 | FPS counter, coordinates, timestamps |
| `--text-sm` | 12px / 16px | 500 | Labels, button text, secondary info |
| `--text-base` | 13px / 20px | 400 | Body, status messages |
| `--text-lg` | 14px / 20px | 600 | Brand, headings |

```css
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&display=swap');
```

---

## Spacing (4px grid)

| Token | Value | Usage |
|-------|-------|-------|
| 1 | 4px | Icon gaps, inline spacing |
| 2 | 8px | Button gaps, group internals |
| 3 | 12px | Card padding, bar padding |
| 4 | 16px | Between groups |
| 6 | 24px | Section separation |

---

## Border Radius

| Token | Value | Usage |
|-------|-------|-------|
| `--radius-sm` | 4px | Buttons, inputs, badges |
| `--radius-md` | 6px | Cards, panels |
| `--radius-full` | 9999px | Status dot, pills |

---

## Elevation

- **Status bar**: `0 1px 0 var(--gray-3)` — subtle bottom border
- **Controls overlay**: `0 -1px 0 var(--gray-3)` with `rgba(2,6,23,0.8)` backdrop

---

## Motion

- **Duration**: 150ms for state changes, 0ms default (snappy)
- **Easing**: `cubic-bezier(0.16, 1, 0.3, 1)` — quick out, gentle settle
- **Respect**: `@media (prefers-reduced-motion: reduce) { * { animation-duration: 0s !important } }`

---

## Component Tokens

### Status Dot
```css
.status-dot { width:8px; height:8px; border-radius:var(--radius-full); transition:background 150ms,box-shadow 150ms; }
.status-dot.live { background:var(--accent); box-shadow:0 0 6px var(--accent-dim); }
.status-dot.dead { background:var(--danger); }
```

### Button
```css
.btn { height:28px; padding:0 10px; border-radius:var(--radius-sm); font:var(--text-sm); border:1px solid var(--gray-3); background:var(--gray-4); color:var(--gray-1); cursor:pointer; transition:150ms; }
.btn:hover { border-color:var(--gray-2); color:var(--gray-0); }
.btn:active { background:var(--gray-3); }
.btn.active { background:var(--accent-dim); border-color:var(--accent); color:var(--accent); }
.btn.danger { background:var(--danger-dim); border-color:var(--danger); color:var(--danger); }
```

### Coordinate Display
```css
.coords { font:var(--text-xs); color:var(--accent); background:rgba(2,6,23,0.85); padding:2px 8px; border-radius:var(--radius-sm); }
```

### Stream Viewport
```css
.viewport { background:var(--gray-6); }
.viewport img { object-fit:contain; cursor:crosshair; }
```

---

## Layout

```
┌─ Status Bar (40px, fixed top) ────────────────────────────────┐
│ ● webdeck   30 fps   1280×720              click to control   │
├─ Stream Viewport (flex-1) ────────────────────────────────────┤
│                                                                │
│                    [1280×720 MJPEG]                            │
│                                                                │
│              ┌──────────────────────────┐                      │
│              │ (640,360)  [↔Swipe] [✕] │  ← controls overlay │
│              └──────────────────────────┘                      │
└────────────────────────────────────────────────────────────────┘
```

- **StatusBar**: 40px height, border-bottom, `--gray-4` background
- **StreamView**: fills remaining space, black background, centered image
- **Controls**: absolute bottom, centered, semi-transparent backdrop

---

## Anti-patterns

- No emoji as icons — use text labels or SVG
- No light mode
- No animations longer than 150ms
- No rounded corners on the stream viewport
- No color-only state indicators (always pair with text)
- No horizontal scroll
