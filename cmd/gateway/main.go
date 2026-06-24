package main

import (
	"context"
	"embed"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"webdeck/internal/device"
	"webdeck/internal/server"
)

//go:embed frontend/dist/*
var frontend embed.FS

var (
	port      = flag.Int("port", 8090, "HTTP server port")
	chromeURL = flag.String("remote", "", "Connect to existing Chrome (ws://host:9222)")
	authFile  = flag.String("auth", "cloud_auth.json", "Cookie JSON file")
	fps       = flag.Int("fps", 30, "Target capture frame rate")
	jpegQ     = flag.Int("jpeg-quality", 75, "JPEG quality 1-100")
	startURL  = flag.String("start-url", "", "Navigate to URL on startup (optional)")
)

func main() {
	flag.Parse()
	slog.Info("starting", "port", *port, "fps", *fps)

	dev, err := device.NewChrome(*chromeURL, *authFile)
	if err != nil {
		slog.Error("device create failed", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := server.New(dev, server.Config{
		Port:     *port,
		FPS:      *fps,
		JPEGQ:    *jpegQ,
		Frontend: frontend,
	})

	if err := srv.Start(ctx); err != nil {
		slog.Error("server error", "err", err)
	}

	slog.Info("shutting down")
	dev.Stop(ctx)
	slog.Info("bye")
}
