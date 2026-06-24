package main

import (
	"context"
	"embed"
	"flag"
	"log"
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
	startURL  = flag.String("start-url", "", "Navigate to URL on startup (optional, for standalone use)")
)

func main() {
	flag.Parse()

	log.Printf("[main] creating device (chrome)...")
	dev, err := device.NewChrome(*chromeURL, *authFile)
	if err != nil {
		log.Fatalf("[main] device: %v", err)
	}

	// Graceful shutdown on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := server.New(dev, server.Config{
		Port:     *port,
		FPS:      *fps,
		JPEGQ:    *jpegQ,
		Frontend: frontend,
	})

	log.Printf("[main] ready → http://localhost:%d", *port)
	if err := srv.Start(ctx); err != nil {
		log.Printf("[main] server: %v", err)
	}

	// Clean shutdown
	log.Printf("[main] shutting down...")
	dev.Stop(ctx)
	log.Printf("[main] done")
}
