package device

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"src-web-gateway/internal/browser"
)

// ChromeDevice implements Device over Chrome DevTools Protocol.
type ChromeDevice struct {
	mu      sync.Mutex
	browser *browser.Browser
	state   string
	recover int

	lastFrameTime time.Time
	lastInputTime time.Time
}

// NewChrome creates a ChromeDevice. In production mode it launches Chrome
// as a subprocess. In debug mode (chromeURL != "") it connects to an
// existing Chrome instance.
func NewChrome(chromeURL, authFile string) (*ChromeDevice, error) {
	var br *browser.Browser
	var err error

	if chromeURL != "" {
		br, err = browser.NewRemote(chromeURL, authFile)
	} else {
		br, err = browser.NewLocal(authFile)
	}
	if err != nil {
		return nil, fmt.Errorf("chrome device: %w", err)
	}

	return &ChromeDevice{
		browser: br,
		state:   "CHROME_STARTING",
	}, nil
}

// Info returns fixed coordinate contract.
func (d *ChromeDevice) Info(ctx context.Context) (*DeviceInfo, error) {
	d.mu.Lock()
	ready := d.state == "RUNNING"
	d.mu.Unlock()
	return &DeviceInfo{
		Width:            1280,
		Height:           720,
		DPR:              1,
		Orientation:      "landscape",
		ScreenshotFormat: "jpeg",
		InputCoordinate:  "screenshot_pixel",
		Backend:          "chrome-cdp",
		Ready:            ready,
	}, nil
}

// Health returns live status.
func (d *ChromeDevice) Health(ctx context.Context) HealthStatus {
	d.mu.Lock()
	state := d.state
	recover := d.recover
	frameAge := time.Since(d.lastFrameTime)
	inputAge := time.Since(d.lastInputTime)
	d.mu.Unlock()
	return HealthStatus{
		OK:           state == "RUNNING",
		State:        state,
		ChromeAlive:  d.browser.IsAlive(),
		CDPConnected: d.browser.IsAlive(),
		PageReady:    state == "RUNNING",
		LastFrameAge: frameAge,
		LastInputAge: inputAge,
		RecoverCount: recover,
	}
}

// Screenshot captures a JPEG frame at 1280×720.
func (d *ChromeDevice) Screenshot(ctx context.Context, opts ScreenshotOptions) ([]byte, error) {
	if opts.Quality == 0 {
		opts.Quality = 75
	}
	jpeg, err := d.browser.ScreenshotJPEG(opts.Quality)
	if err == nil {
		d.mu.Lock()
		d.lastFrameTime = time.Now()
		d.mu.Unlock()
	}
	return jpeg, err
}

// Tap executes a trusted CDP click.
func (d *ChromeDevice) Tap(ctx context.Context, x, y int) error {
	err := d.browser.Click(x, y)
	if err == nil {
		d.mu.Lock()
		d.lastInputTime = time.Now()
		d.mu.Unlock()
	}
	return err
}

// Swipe executes a CDP touch swipe.
func (d *ChromeDevice) Swipe(ctx context.Context, x1, y1, x2, y2, durationMs int) error {
	if durationMs <= 0 {
		durationMs = 300
	}
	err := d.browser.Swipe(x1, y1, x2, y2, durationMs)
	if err == nil {
		d.mu.Lock()
		d.lastInputTime = time.Now()
		d.mu.Unlock()
	}
	return err
}

// Key sends a keyboard event.
func (d *ChromeDevice) Key(ctx context.Context, key string) error {
	err := d.browser.Key(key)
	if err == nil {
		d.mu.Lock()
		d.lastInputTime = time.Now()
		d.mu.Unlock()
	}
	return err
}

// Start navigates to cloud game and enters gameplay.
func (d *ChromeDevice) Start(ctx context.Context) error {
	d.mu.Lock()
	d.state = "PAGE_LOADING"
	d.mu.Unlock()
	if err := d.browser.Navigate(); err != nil {
		d.mu.Lock()
		d.state = "DEGRADED"
		d.mu.Unlock()
		return fmt.Errorf("start: %w", err)
	}
	d.mu.Lock()
	d.state = "RUNNING"
	d.mu.Unlock()
	log.Printf("[device] state → RUNNING")
	return nil
}

// Stop closes the game session.
func (d *ChromeDevice) Stop(ctx context.Context) error {
	d.mu.Lock()
	d.state = "STOPPED"
	d.mu.Unlock()
	d.browser.Close()
	return nil
}

// Restart stops then starts.
func (d *ChromeDevice) Restart(ctx context.Context) error {
	d.Stop(ctx)
	return d.Start(ctx)
}

// Reset destroys and recreates the browser runtime.
func (d *ChromeDevice) Reset(ctx context.Context) error {
	d.mu.Lock()
	d.state = "RECOVERING"
	d.recover++
	d.mu.Unlock()
	d.browser.Close()
	// Re-create browser — for now, caller must restart the whole process
	// TODO: full subprocess restart
	d.mu.Lock()
	d.state = "INIT"
	d.mu.Unlock()
	log.Printf("[device] reset #%d complete", d.recover)
	return nil
}

// DismissHTML closes overlay dialogs (vanilla Vant UI).
// This is a compat helper — not part of the Device interface.
func (d *ChromeDevice) DismissHTML() {
	d.browser.DismissHTML()
}
