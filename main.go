package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"time"

	"src-web-gateway/browser"
	"src-web-gateway/stream"
)

//go:embed frontend/*
var frontend embed.FS

var (
	port     = flag.Int("port", 8090, "HTTP server port")
	authFile = flag.String("auth", "config/cloud_auth.json", "Playwright cookie JSON")
	fps      = flag.Int("fps", 30, "Target frame rate")
	jpegQ    = flag.Int("jpeg-quality", 75, "JPEG quality 1-100")
)

func main() {
	flag.Parse()

	log.Printf("[main] starting browser (headless chromium)...")
	br, err := browser.New(*authFile)
	if err != nil {
		log.Fatalf("[main] browser init: %v", err)
	}
	defer br.Close()

	log.Printf("[main] navigating to cloud game...")
	if err := br.Navigate(); err != nil {
		log.Fatalf("[main] navigate: %v", err)
	}

	hub := stream.NewHub()

	// Capture loop: screenshot → compress → broadcast
	frameInterval := time.Second / time.Duration(*fps)
	go func() {
		log.Printf("[capture] %d FPS, JPEG Q%d", *fps, *jpegQ)
		for {
			t0 := time.Now()
			jpeg, err := br.ScreenshotJPEG(*jpegQ)
			if err != nil {
				log.Printf("[capture] screenshot: %v", err)
				time.Sleep(frameInterval)
				continue
			}
			hub.SetFrame(jpeg)
			elapsed := time.Since(t0)
			if elapsed < frameInterval {
				time.Sleep(frameInterval - elapsed)
			}
		}
	}()

	// Command handler: dispatch WS commands to browser
	onCmd := func(cmd stream.Cmd) {
		switch cmd.Type {
		case "click":
			br.Click(cmd.X, cmd.Y)
			log.Printf("[cmd] click (%d,%d)", cmd.X, cmd.Y)
		case "swipe":
			br.Swipe(cmd.X1, cmd.Y1, cmd.X2, cmd.Y2)
			log.Printf("[cmd] swipe (%d,%d)→(%d,%d)", cmd.X1, cmd.Y1, cmd.X2, cmd.Y2)
		case "key":
			br.Key(cmd.Key)
			log.Printf("[cmd] key %s", cmd.Key)
		case "dismiss":
			br.DismissHTML()
			log.Printf("[cmd] dismiss")
		}
	}

	mux := http.NewServeMux()

	// Frontend (embedded)
	frontendFS, _ := fs.Sub(frontend, "frontend")
	mux.Handle("/", http.FileServer(http.FS(frontendFS)))

	// WebSocket (streaming + commands)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.HandleWS(w, r, onCmd)
	})

	// HTTP API fallbacks
	mux.HandleFunc("/api/click", func(w http.ResponseWriter, r *http.Request) {
		x, _ := strconv.Atoi(r.URL.Query().Get("x"))
		y, _ := strconv.Atoi(r.URL.Query().Get("y"))
		br.Click(x, y)
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("/api/swipe", func(w http.ResponseWriter, r *http.Request) {
		x1, _ := strconv.Atoi(r.URL.Query().Get("x1"))
		y1, _ := strconv.Atoi(r.URL.Query().Get("y1"))
		x2, _ := strconv.Atoi(r.URL.Query().Get("x2"))
		y2, _ := strconv.Atoi(r.URL.Query().Get("y2"))
		br.Swipe(x1, y1, x2, y2)
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("/api/dismiss", func(w http.ResponseWriter, r *http.Request) {
		br.DismissHTML()
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("/api/screenshot", func(w http.ResponseWriter, r *http.Request) {
		jpeg, err := br.ScreenshotJPEG(*jpegQ)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(jpeg)
	})
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	// MJPEG stream (compatible with <img src="/stream">)
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=FRAME")
		w.Header().Set("Cache-Control", "no-cache")
		for {
			jpeg, err := br.ScreenshotJPEG(*jpegQ)
			if err != nil {
				break
			}
			_, err = fmt.Fprintf(w, "--FRAME\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(jpeg))
			if err != nil {
				break
			}
			w.Write(jpeg)
			w.Write([]byte("\r\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(time.Second / time.Duration(*fps))
		}
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("[main] ready → http://localhost%s  (%d FPS, JPEG Q%d)", addr, *fps, *jpegQ)
	log.Fatal(http.ListenAndServe(addr, mux))
}
