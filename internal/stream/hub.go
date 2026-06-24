package stream

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// Cmd represents a command from the browser client.
type Cmd struct {
	Type string `json:"type"`
	X    int    `json:"x,omitempty"`
	Y    int    `json:"y,omitempty"`
	X1   int    `json:"x1,omitempty"`
	Y1   int    `json:"y1,omitempty"`
	X2   int    `json:"x2,omitempty"`
	Y2   int    `json:"y2,omitempty"`
	Key  string `json:"key,omitempty"`
}

type CommandHandler func(Cmd)

type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
	frame   []byte
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]bool)}
}

func (h *Hub) SetFrame(jpeg []byte) {
	h.mu.Lock()
	h.frame = jpeg
	h.mu.Unlock()

	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.clients {
		conn.WriteMessage(websocket.BinaryMessage, jpeg)
	}
}

func (h *Hub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
	conn.Close()
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request, onCmd CommandHandler) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	go func() {
		defer h.remove(conn)
		for {
			var cmd Cmd
			if err := conn.ReadJSON(&cmd); err != nil {
				return
			}
			if onCmd != nil && cmd.Type != "" && cmd.Type != "ping" {
				onCmd(cmd)
			}
		}
	}()
}
