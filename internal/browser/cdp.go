package browser

import (
	"context"
	"strings"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ScreenshotJPEG captures a JPEG screenshot via CDP. Retries once on page detach.
func (b *Browser) ScreenshotJPEG(quality int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(b.ctx, 5*time.Second)
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		buf, err = page.CaptureScreenshot().
			WithFormat(page.CaptureScreenshotFormatJpeg).
			WithQuality(int64(quality)).
			Do(ctx)
		return err
	}))
	// Retry once on page detach
	if err != nil && strings.Contains(err.Error(), "Not attached") {
		time.Sleep(500 * time.Millisecond)
		ctx2, cancel2 := context.WithTimeout(b.ctx, 5*time.Second)
		defer cancel2()
		err = chromedp.Run(ctx2, chromedp.ActionFunc(func(ctx context.Context) error {
			var err2 error
			buf, err2 = page.CaptureScreenshot().
				WithFormat(page.CaptureScreenshotFormatJpeg).
				WithQuality(int64(quality)).
				Do(ctx)
			return err2
		}))
	}
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

// Eval runs JavaScript in the page and returns the result as JSON.
func (b *Browser) Eval(js string) (string, error) {
	var result string
	err := chromedp.Run(b.ctx, chromedp.Evaluate(js, &result))
	return result, err
}
