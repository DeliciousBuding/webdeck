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
	alloc   context.Context // allocator context (lives forever)
	ctx     context.Context // current chromedp context (replaced after navigation)
	cancel  context.CancelFunc
	alive   atomic.Bool
	authFile string
}

// NewLocal launches a headless Chrome with mobile emulation.
func NewLocal(authFile string) (*Browser, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("headless", "new"),
			chromedp.Flag("enable-gpu", true),
			chromedp.Flag("use-gl", "swiftshader"),
			chromedp.Flag("enable-webgl", true),
			chromedp.Flag("enable-accelerated-video-decode", true),
			chromedp.Flag("ignore-gpu-blocklist", true),
		chromedp.WindowSize(GameW, GameH),
		chromedp.UserAgent(mobileUA),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	return newBrowser(allocCtx, authFile)
}

// NewRemote connects to an existing Chrome via --remote-debugging-port.
func NewRemote(chromeURL, authFile string) (*Browser, error) {
	allocCtx, _ := chromedp.NewRemoteAllocator(context.Background(), chromeURL)
	// NOTE: do NOT cancel allocCtx — it must outlive all chromedp contexts
	return newBrowser(allocCtx, authFile)
}

func newBrowser(allocCtx context.Context, authFile string) (*Browser, error) {
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(string, ...interface{}) {}))

	if err := chromedp.Run(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("chrome init: %w", err)
	}

	b := &Browser{alloc: allocCtx, ctx: ctx, cancel: cancel, authFile: authFile}
	b.alive.Store(true)
	b.init()
	return b, nil
}

// init runs common post-creation setup: viewport, UA, stealth, cookies.
func (b *Browser) init() {
	chromedp.Run(b.ctx,
		chromedp.EmulateViewport(GameW, GameH, chromedp.EmulateMobile, chromedp.EmulateTouch),
		emulation.SetUserAgentOverride(mobileUA).WithAcceptLanguage("zh-CN"),
		chromedp.Evaluate(stealthJS, nil),
	)
	if b.authFile != "" {
		b.loadCookies(b.authFile)
	}
}

// refreshCtx creates a new chromedp context from the allocator.
// Called after navigation without cancelling the old one.
func (b *Browser) refreshCtx() {
	b.ctx, b.cancel = chromedp.NewContext(b.alloc, chromedp.WithLogf(func(string, ...interface{}) {}))
	b.init()
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

const mobileUA = "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Mobile Safari/537.36"

const stealthJS = `
	Object.defineProperty(navigator, "webdriver", {get: () => false});
	Object.defineProperty(navigator, "plugins", {get: () => [1,2,3,4,5]});
	Object.defineProperty(navigator, "languages", {get: () => ["zh-CN","en"]});
	window.chrome = {runtime: {}};
`
