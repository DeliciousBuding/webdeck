package browser

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ScreenshotJPEG captures a JPEG screenshot via CDP Page.captureScreenshot.
func (b *Browser) ScreenshotJPEG(quality int) ([]byte, error) {
	var buf []byte
	err := chromedp.Run(b.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		buf, err = page.CaptureScreenshot().
			WithFormat(page.CaptureScreenshotFormatJpeg).
			WithQuality(int64(quality)).
			WithClip(&page.Viewport{X: 0, Y: 0, Width: GameW, Height: GameH, Scale: 1}).
			Do(ctx)
		return err
	}))
	return buf, err
}

// Click sends a trusted CDP mouse click. Proven sequence: focus → press → 80ms → release.
func (b *Browser) Click(x, y int) error {
	return chromedp.Run(b.ctx,
		chromedp.Evaluate(`window.focus()`, nil),
		chromedp.Sleep(50*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return input.DispatchMouseEvent(input.MousePressed, float64(x), float64(y)).
				WithButton(input.Left).WithClickCount(1).Do(ctx)
		}),
		chromedp.Sleep(80*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return input.DispatchMouseEvent(input.MouseReleased, float64(x), float64(y)).
				WithButton(input.Left).WithClickCount(1).Do(ctx)
		}),
	)
}

// Swipe performs a CDP touch swipe sequence.
// durationMs controls the total gesture time (0 = default 300ms).
func (b *Browser) Swipe(x1, y1, x2, y2 int, durationMs int) error {
	if durationMs <= 0 {
		durationMs = 300
	}
	steps := 10
	stepSleep := time.Duration(durationMs/steps) * time.Millisecond

	return chromedp.Run(b.ctx,
		chromedp.Evaluate(`window.focus()`, nil),
		chromedp.Sleep(30*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return input.DispatchTouchEvent(input.TouchStart, []*input.TouchPoint{
				{X: float64(x1), Y: float64(y1), ID: 1, RadiusX: 5, RadiusY: 5, Force: 1},
			}).Do(ctx)
		}),
		chromedp.Sleep(30*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			for i := 1; i <= steps; i++ {
				t := float64(i) / float64(steps)
				cx := float64(x1) + (float64(x2)-float64(x1))*t
				cy := float64(y1) + (float64(y2)-float64(y1))*t
				input.DispatchTouchEvent(input.TouchMove, []*input.TouchPoint{
					{X: cx, Y: cy, ID: 1, RadiusX: 5, RadiusY: 5, Force: 1},
				}).Do(ctx)
				time.Sleep(stepSleep)
			}
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return input.DispatchTouchEvent(input.TouchEnd, []*input.TouchPoint{
				{X: float64(x2), Y: float64(y2), ID: 1},
			}).Do(ctx)
		}),
	)
}

// Key sends a keyboard event.
func (b *Browser) Key(k string) error {
	return chromedp.Run(b.ctx, chromedp.KeyEvent(k))
}
