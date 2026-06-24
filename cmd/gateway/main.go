package main

import (
	"embed"
	"flag"
	"log"

	"src-web-gateway/internal/device"
	"src-web-gateway/internal/server"
)

//go:embed frontend/*
var frontend embed.FS

var (
	port      = flag.Int("port", 8090, "HTTP server port")
	chromeURL = flag.String("remote", "", "Connect to existing Chrome (ws://host:9222)")
	authFile  = flag.String("auth", "cloud_auth.json", "Cookie JSON file")
	fps       = flag.Int("fps", 30, "Target capture frame rate")
	jpegQ     = flag.Int("jpeg-quality", 75, "JPEG quality 1-100")
)

func main() {
	flag.Parse()

	log.Printf("[main] creating device (chrome)...")
	dev, err := device.NewChrome(*chromeURL, *authFile)
	if err != nil {
		log.Fatalf("[main] device: %v", err)
	}
	defer dev.Stop(nil)

	log.Printf("[main] starting cloud game...")
	if err := dev.Start(nil); err != nil {
		log.Fatalf("[main] navigate: %v", err)
	}

	srv := server.New(dev, server.Config{
		Port:     *port,
		FPS:      *fps,
		JPEGQ:    *jpegQ,
		Frontend: frontend,
	})

	log.Printf("[main] ready → http://localhost:%d", *port)
	log.Fatal(srv.Start())
}
