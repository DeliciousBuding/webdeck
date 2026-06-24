package browser

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

const (
	GameW = 1280
	GameH = 720
)

// Browser manages a headless Chrome instance via chromedp/CDP.
type Browser struct {
	ctx    context.Context
	cancel context.CancelFunc
	alive  atomic.Bool
}

// NewLocal launches a headless Chrome with mobile emulation.
func NewLocal(authFile string) (*Browser, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.WindowSize(GameW, GameH),
		chromedp.UserAgent(mobileUA),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(string, ...interface{}) {}))

	if err := chromedp.Run(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("chrome launch: %w", err)
	}

	b := &Browser{ctx: ctx, cancel: cancel}
	b.alive.Store(true)
	b.init(authFile)
	return b, nil
}

// NewRemote connects to an existing Chrome via --remote-debugging-port.
func NewRemote(chromeURL, authFile string) (*Browser, error) {
	allocCtx, cancel := chromedp.NewRemoteAllocator(context.Background(), chromeURL)
	ctx, _ := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(string, ...interface{}) {}))

	if err := chromedp.Run(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("connect chrome: %w", err)
	}

	b := &Browser{ctx: ctx, cancel: cancel}
	b.alive.Store(true)
	b.init(authFile)
	return b, nil
}

// init runs common post-creation setup: viewport, UA, stealth, cookies.
func (b *Browser) init(authFile string) {
	chromedp.Run(b.ctx,
		chromedp.EmulateViewport(GameW, GameH, chromedp.EmulateMobile, chromedp.EmulateTouch),
		emulation.SetUserAgentOverride(mobileUA).WithAcceptLanguage("zh-CN"),
		chromedp.Evaluate(stealthJS, nil),
	)
	if authFile != "" {
		b.loadCookies(authFile)
	}
}

// IsAlive reports whether the browser is still running.
func (b *Browser) IsAlive() bool { return b.alive.Load() }

// Close shuts down the browser.
func (b *Browser) Close() {
	b.alive.Store(false)
	if b.cancel != nil {
		b.cancel()
		b.cancel = nil
	}
}

// Ctx returns the chromedp context for advanced use.
func (b *Browser) Ctx() context.Context { return b.ctx }

const mobileUA = "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Mobile Safari/537.36"

const stealthJS = `
	Object.defineProperty(navigator, "webdriver", {get: () => false});
	Object.defineProperty(navigator, "plugins", {get: () => [1,2,3,4,5]});
	Object.defineProperty(navigator, "languages", {get: () => ["zh-CN","en"]});
	window.chrome = {runtime: {}};
`
