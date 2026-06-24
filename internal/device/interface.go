// Package device defines the stable Device interface that SRC talks to.
// Chrome/CDP is one implementation; future backends (WebRTC, Moonlight, etc.)
// implement the same interface without changing SRC.
package device

import (
	"context"
)

// DeviceInfo describes the virtual device's coordinate system and capabilities.
type DeviceInfo struct {
	Width             int    `json:"width"`
	Height            int    `json:"height"`
	DPR               int    `json:"dpr"`
	Orientation       string `json:"orientation"`
	ScreenshotFormat  string `json:"screenshot_format"`
	InputCoordinate   string `json:"input_coordinate"`
	Backend           string `json:"backend"`
	Ready             bool   `json:"ready"`
}

// HealthStatus reports runtime health for the SRC adapter and recovery logic.
type HealthStatus struct {
	OK              bool          `json:"ok"`
	State           string        `json:"state"`
	ChromeAlive     bool          `json:"chrome_alive"`
	CDPConnected    bool          `json:"cdp_connected"`
	PageReady       bool          `json:"page_ready"`
	LastFrameAge    int64 `json:"last_frame_age_ms"`
	LastInputAge    int64 `json:"last_input_age_ms"`
	RecoverCount    int           `json:"recover_count"`
}

// ScreenshotOptions controls screenshot format and quality.
type ScreenshotOptions struct {
	Format  string // "jpeg" or "png"
	Quality int    // 1-100, JPEG only
}

// Device is the stable interface for all device backends.
// SRC and WebUI MUST only call methods on this interface.
// They MUST NOT call browser or CDP directly.
type Device interface {
	// Info returns static device metadata (coordinates, backend).
	Info(ctx context.Context) (*DeviceInfo, error)

	// Health returns live health/runtime status.
	Health(ctx context.Context) HealthStatus

	// Screenshot captures the current frame.
	// Returns JPEG or PNG bytes at exactly Width×Height.
	Screenshot(ctx context.Context, opts ScreenshotOptions) ([]byte, error)

	// Tap sends a click at logical coordinates.
	// (x, y) are in screenshot pixel space.
	Tap(ctx context.Context, x, y int) error

	// Swipe performs a touch swipe gesture.
	// Coordinates in screenshot pixel space.
	Swipe(ctx context.Context, x1, y1, x2, y2 int, durationMs int) error

	// Key sends a keyboard key event.
	Key(ctx context.Context, key string) error

	// Start navigates to the cloud game and enters gameplay.
	Start(ctx context.Context) error

	// Stop ends the game session.
	Stop(ctx context.Context) error

	// Restart stops then starts.
	Restart(ctx context.Context) error

	// Reset kills and restarts the entire browser runtime.
	// Used for recovering from degraded states.
	Reset(ctx context.Context) error
}
