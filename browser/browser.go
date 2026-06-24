package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

const (
	GameW = 1280
	GameH = 720
)

type Browser struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type CookieFile struct {
	Cookies []Cookie `json:"cookies"`
}
type Cookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	HTTPOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
	SameSite string  `json:"sameSite"`
}

func New(authFile string) (*Browser, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.WindowSize(1280, 720),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(string, ...interface{}) {}))

	if err := chromedp.Run(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("chromedp init: %w", err)
	}

	// Stealth init script
	if err := chromedp.Run(ctx, chromedp.Evaluate(`
		Object.defineProperty(navigator, "webdriver", {get: () => false});
		Object.defineProperty(navigator, "plugins", {get: () => [1,2,3,4,5]});
		Object.defineProperty(navigator, "languages", {get: () => ["zh-CN","en"]});
		window.chrome = {runtime: {}};
	`, nil)); err != nil {
		log.Printf("[browser] stealth warn: %v", err)
	}

	// Set cookies
	if authFile != "" {
		setCookies(ctx, authFile)
	}

	return &Browser{ctx: ctx, cancel: cancel}, nil
}

func setCookies(ctx context.Context, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[browser] cookie read: %v", err)
		return
	}
	var cf CookieFile
	if err := json.Unmarshal(data, &cf); err != nil {
		log.Printf("[browser] cookie parse: %v", err)
		return
	}
	for _, ck := range cf.Cookies {
		expr := cdp.TimeSinceEpoch(time.Unix(int64(ck.Expires), 0))
		sameSite := network.CookieSameSiteLax
		switch ck.SameSite {
		case "Strict":
			sameSite = network.CookieSameSiteStrict
		case "None":
			sameSite = network.CookieSameSiteNone
		}
		chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			return network.SetCookie(ck.Name, ck.Value).
				WithDomain(ck.Domain).WithPath(ck.Path).
				WithSecure(ck.Secure).WithHTTPOnly(ck.HTTPOnly).
				WithSameSite(sameSite).WithExpires(&expr).
				Do(ctx)
		}))
	}
	log.Printf("[browser] loaded %d cookies", len(cf.Cookies))

	// Need to navigate to set cookies on the domain
	chromedp.Run(ctx, chromedp.Navigate("https://sr.mihoyo.com/cloud/"))
}

func (b *Browser) Navigate() error {
	return chromedp.Run(b.ctx,
		chromedp.Navigate("https://sr.mihoyo.com/cloud/"),
		chromedp.Sleep(3*time.Second),
		chromedp.WaitVisible(`.wel-card__content--start`, chromedp.ByQuery),
		chromedp.Click(`.wel-card__content--start`, chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			for i := 0; i < 60; i++ {
				var hasVideo bool
				chromedp.Evaluate(`!!document.querySelector("video.game-player__video")`, &hasVideo).Do(ctx)
				if hasVideo {
					log.Printf("[browser] game video ready (%ds)", i*2)
					return nil
				}
				chromedp.Evaluate(`document.querySelector('.van-button--danger')?.click()`, nil).Do(ctx)
				chromedp.Evaluate(`document.querySelector('.van-button--default')?.click()`, nil).Do(ctx)
				time.Sleep(2 * time.Second)
			}
			return fmt.Errorf("timeout waiting for game video")
		}),
	)
}

func (b *Browser) Close() {
	if b.cancel != nil {
		b.cancel()
	}
}

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

func (b *Browser) Swipe(x1, y1, x2, y2 int) error {
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
			steps := 10
			for i := 1; i <= steps; i++ {
				t := float64(i) / float64(steps)
				cx := float64(x1) + (float64(x2)-float64(x1))*t
				cy := float64(y1) + (float64(y2)-float64(y1))*t
				input.DispatchTouchEvent(input.TouchMove, []*input.TouchPoint{
					{X: cx, Y: cy, ID: 1, RadiusX: 5, RadiusY: 5, Force: 1},
				}).Do(ctx)
				time.Sleep(20 * time.Millisecond)
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

func (b *Browser) Key(k string) error {
	return chromedp.Run(b.ctx, chromedp.KeyEvent(k))
}

func (b *Browser) DismissHTML() error {
	return chromedp.Run(b.ctx, chromedp.Evaluate(`
		document.querySelector('.van-button--default')?.click();
		document.querySelector('.van-button--danger')?.click();
		document.querySelector('.van-overlay')?.click();
	`, nil))
}
