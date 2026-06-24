// Package logx provides structured logging for webdeck.
// Wraps log/slog with convenience helpers and periodic health checks.
package logx

import (
	"log/slog"
	"os"
	"sync/atomic"
	"time"
)

var (
	frameCount atomic.Int64
	frameBytes atomic.Int64
	lastFrame  atomic.Int64 // unix milli
)

// Init configures structured JSON logging.
func Init() {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(h))
}

// Frame records a captured frame for telemetry.
func Frame(size int) {
	frameCount.Add(1)
	frameBytes.Add(int64(size))
	lastFrame.Store(time.Now().UnixMilli())
}

// StartHealth starts a goroutine that periodically logs health telemetry.
func StartHealth(stop <-chan struct{}) {
	go func() {
		tick := time.NewTicker(10 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-stop:
				return
			case <-tick.C:
				count := frameCount.Swap(0)
				bytes := frameBytes.Swap(0)
				slog.Info("capture health",
					"fps", count/10,
					"avg_bytes", safeDiv(bytes, count),
					"last_frame_ms_ago", time.Since(time.UnixMilli(lastFrame.Load())).Milliseconds(),
				)
			}
		}
	}()
}

func safeDiv(a, b int64) int64 {
	if b == 0 {
		return 0
	}
	return a / b
}
