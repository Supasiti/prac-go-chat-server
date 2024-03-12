package server

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type hub struct {
	clients    map[*client]bool
	register   chan *client
	unregister chan *client
	notify     chan *chatMessage
}

func NewHub() *hub {
	return &hub{
		clients:    make(map[*client]bool),
		register:   make(chan *client),
		unregister: make(chan *client),
		notify:     make(chan *chatMessage),
	}
}

func (h *hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			delete(h.clients, client)
		case chat := <-h.notify:
			for client := range h.clients {
				client.send <- chat
			}
		}
	}
}

func (h *hub) Serve(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("error upgrading to WebSocket:", slog.Any("err", err))
		// probably need to send something back
		return
	}

	client := NewClient(h, conn)
	client.Start()
}
