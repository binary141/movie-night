package main

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/coder/websocket"
)

// Hub fans out a single "board changed" notification to every connected
// client. Clients react by re-fetching /board themselves, so each viewer
// gets a render personalized to their own vote — the hub never needs to
// know who's on the other end.
type Hub struct {
	mu    sync.Mutex
	conns map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{conns: make(map[*websocket.Conn]struct{})}
}

func (h *Hub) add(c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[c] = struct{}{}
}

func (h *Hub) remove(c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, c)
}

// Broadcast tells every connected client the board changed.
func (h *Hub) Broadcast() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.conns {
		if err := c.Write(context.Background(), websocket.MessageText, []byte("board-changed")); err != nil {
			log.Printf("ws broadcast: %v", err)
		}
	}
}

func (a *App) serveWS(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("ws accept: %v", err)
		return
	}
	defer c.CloseNow()

	a.hub.add(c)
	defer a.hub.remove(c)

	ctx := r.Context()
	for {
		if _, _, err := c.Read(ctx); err != nil {
			return
		}
	}
}
