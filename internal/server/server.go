package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"time"

	"webdeck/internal/device"
	"webdeck/internal/stream"
)

// Config holds server configuration.
type Config struct {
	Port    int
	FPS     int
	JPEGQ   int
	Frontend embed.FS
}

// Server wraps HTTP routes and WebSocket hub.
type Server struct {
	dev device.Device
	hub *stream.Hub
	cfg Config
	mux *http.ServeMux
}

// New creates a Server. All routes go through the Device interface.
func New(dev device.Device, cfg Config) *Server {
	s := &Server{
		dev: dev,
		hub: stream.NewHub(),
		cfg: cfg,
		mux: http.NewServeMux(),
	}

	// Frontend
	frontendFS, _ := fs.Sub(cfg.Frontend, "frontend/dist")
	s.mux.Handle("/", http.FileServer(http.FS(frontendFS)))

	// WebSocket
	s.mux.HandleFunc("/ws", s.handleWS)

	// MJPEG stream
	s.mux.HandleFunc("/stream", s.handleStream)

	// v1 stable API
	s.mux.HandleFunc("/api/v1/health", s.handleV1Health)
	s.mux.HandleFunc("/api/v1/device/info", s.handleV1DeviceInfo)
	s.mux.HandleFunc("/api/v1/device/screenshot", s.handleV1Screenshot)
	s.mux.HandleFunc("/api/v1/input/tap", s.handleV1Tap)
	s.mux.HandleFunc("/api/v1/input/swipe", s.handleV1Swipe)
	s.mux.HandleFunc("/api/v1/input/key", s.handleV1Key)
	s.mux.HandleFunc("/api/v1/app/start", s.handleV1AppStart)
	s.mux.HandleFunc("/api/v1/app/stop", s.handleV1AppStop)
	s.mux.HandleFunc("/api/v1/app/restart", s.handleV1AppRestart)
	s.mux.HandleFunc("/api/v1/session/reset", s.handleV1SessionReset)

	// Compat aliases (deprecated, mapped to v1)
	s.mux.HandleFunc("/api/click", s.handleCompatClick)
	s.mux.HandleFunc("/api/swipe", s.handleCompatSwipe)
	s.mux.HandleFunc("/api/screenshot", s.handleCompatScreenshot)
	s.mux.HandleFunc("/api/navigate", s.handleCompatNavigate)
	s.mux.HandleFunc("/api/dismiss", s.handleCompatDismiss)
	s.mux.HandleFunc("/api/health", s.handleCompatHealth)

	return s
}

// Start begins the HTTP server and capture loop. Blocks until ctx is cancelled,
// then shuts down gracefully.
func (s *Server) Start(ctx context.Context) error {
	// Capture loop
	interval := time.Second / time.Duration(s.cfg.FPS)
	stopCapture := make(chan struct{})
	go func() {
		defer close(stopCapture)
		log.Printf("[server] capture %d FPS, JPEG Q%d", s.cfg.FPS, s.cfg.JPEGQ)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			t0 := time.Now()
			jpeg, err := s.dev.Screenshot(context.Background(), device.ScreenshotOptions{
				Format: "jpeg", Quality: s.cfg.JPEGQ,
			})
			if err != nil {
				log.Printf("[server] screenshot: %v", err)
				time.Sleep(interval)
				continue
			}
			s.hub.SetFrame(jpeg)
			if elapsed := time.Since(t0); elapsed < interval {
				time.Sleep(interval - elapsed)
			}
		}
	}()

	addr := fmt.Sprintf(":%d", s.cfg.Port)
	srv := &http.Server{Addr: addr, Handler: s.mux}

	// Run HTTP server in background
	go func() {
		<-ctx.Done()
		log.Printf("[server] shutting down...")
		srv.Shutdown(context.Background())
	}()

	log.Printf("[server] listening on %s", addr)
	err := srv.ListenAndServe()
	<-stopCapture
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// ── v1 API handlers ──

func (s *Server) handleV1Health(w http.ResponseWriter, r *http.Request) {
	status := s.dev.Health(r.Context())
	writeJSON(w, status)
}

func (s *Server) handleV1DeviceInfo(w http.ResponseWriter, r *http.Request) {
	info, err := s.dev.Info(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, info)
}

func (s *Server) handleV1Screenshot(w http.ResponseWriter, r *http.Request) {
	jpeg, err := s.dev.Screenshot(r.Context(), device.ScreenshotOptions{
		Format: "jpeg", Quality: s.cfg.JPEGQ,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(jpeg)
}

func (s *Server) handleV1Tap(w http.ResponseWriter, r *http.Request) {
	var req struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	if r.Method == "POST" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", 400)
			return
		}
	} else {
		req.X, _ = strconv.Atoi(r.URL.Query().Get("x"))
		req.Y, _ = strconv.Atoi(r.URL.Query().Get("y"))
	}
	if err := s.dev.Tap(r.Context(), req.X, req.Y); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprint(w, "ok")
}

func (s *Server) handleV1Swipe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		X1          int `json:"x1"`
		Y1          int `json:"y1"`
		X2          int `json:"x2"`
		Y2          int `json:"y2"`
		DurationMs int `json:"duration_ms"`
	}
	if r.Method == "POST" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", 400)
			return
		}
	} else {
		req.X1, _ = strconv.Atoi(r.URL.Query().Get("x1"))
		req.Y1, _ = strconv.Atoi(r.URL.Query().Get("y1"))
		req.X2, _ = strconv.Atoi(r.URL.Query().Get("x2"))
		req.Y2, _ = strconv.Atoi(r.URL.Query().Get("y2"))
	}
	if req.DurationMs <= 0 {
		req.DurationMs = 300
	}
	if err := s.dev.Swipe(r.Context(), req.X1, req.Y1, req.X2, req.Y2, req.DurationMs); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprint(w, "ok")
}

func (s *Server) handleV1Key(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key string `json:"key"`
	}
	if r.Method == "POST" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", 400)
			return
		}
	} else {
		req.Key = r.URL.Query().Get("k")
	}
	if err := s.dev.Key(r.Context(), req.Key); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprint(w, "ok")
}

func (s *Server) handleV1AppStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	err := s.dev.Start(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprint(w, "ok")
}

func (s *Server) handleV1AppStop(w http.ResponseWriter, r *http.Request) {
	err := s.dev.Stop(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprint(w, "ok")
}

func (s *Server) handleV1AppRestart(w http.ResponseWriter, r *http.Request) {
	err := s.dev.Restart(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprint(w, "ok")
}

func (s *Server) handleV1SessionReset(w http.ResponseWriter, r *http.Request) {
	err := s.dev.Reset(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprint(w, "ok")
}

// ── Compat aliases ──

func (s *Server) handleCompatClick(w http.ResponseWriter, r *http.Request) {
	x, _ := strconv.Atoi(r.URL.Query().Get("x"))
	y, _ := strconv.Atoi(r.URL.Query().Get("y"))
	if err := s.dev.Tap(r.Context(), x, y); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprint(w, "ok")
}

func (s *Server) handleCompatSwipe(w http.ResponseWriter, r *http.Request) {
	x1, _ := strconv.Atoi(r.URL.Query().Get("x1"))
	y1, _ := strconv.Atoi(r.URL.Query().Get("y1"))
	x2, _ := strconv.Atoi(r.URL.Query().Get("x2"))
	y2, _ := strconv.Atoi(r.URL.Query().Get("y2"))
	if err := s.dev.Swipe(r.Context(), x1, y1, x2, y2, 300); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprint(w, "ok")
}

func (s *Server) handleCompatScreenshot(w http.ResponseWriter, r *http.Request) {
	s.handleV1Screenshot(w, r)
}

func (s *Server) handleCompatNavigate(w http.ResponseWriter, r *http.Request) {
	s.handleV1AppStart(w, r)
}

func (s *Server) handleCompatDismiss(w http.ResponseWriter, r *http.Request) {
	// Special: HTML-only dismiss, not in Device interface
	if cd, ok := s.dev.(*device.ChromeDevice); ok {
		cd.DismissHTML()
	}
	fmt.Fprint(w, "ok")
}

func (s *Server) handleCompatHealth(w http.ResponseWriter, r *http.Request) {
	status := s.dev.Health(r.Context())
	if status.OK {
		fmt.Fprint(w, "ok")
	} else {
		fmt.Fprint(w, status.State)
	}
}

// ── WebSocket ──

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	s.hub.HandleWS(w, r, func(cmd stream.Cmd) {
		switch cmd.Type {
		case "click":
			s.dev.Tap(r.Context(), cmd.X, cmd.Y)
		case "swipe":
			s.dev.Swipe(r.Context(), cmd.X1, cmd.Y1, cmd.X2, cmd.Y2, 300)
		case "key":
			s.dev.Key(r.Context(), cmd.Key)
		case "dismiss":
			if cd, ok := s.dev.(*device.ChromeDevice); ok {
				cd.DismissHTML()
			}
		}
	})
}

// ── MJPEG ──

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=FRAME")
	w.Header().Set("Cache-Control", "no-cache")
	interval := time.Second / time.Duration(s.cfg.FPS)
	for {
		jpeg, err := s.dev.Screenshot(r.Context(), device.ScreenshotOptions{
			Format: "jpeg", Quality: s.cfg.JPEGQ,
		})
		if err != nil {
			break
		}
		fmt.Fprintf(w, "--FRAME\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(jpeg))
		w.Write(jpeg)
		w.Write([]byte("\r\n"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		time.Sleep(interval)
	}
}

// ── helpers ──

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
