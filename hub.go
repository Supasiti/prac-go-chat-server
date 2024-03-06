package main

import (
	"context"
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
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
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

	ctx := context.Background()
	client := NewClient(ctx, h, conn)
	client.Start()
}
