package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

type client struct {
	upgrader websocket.Upgrader
}

func NewClient() *client {
	bufSize := 1024

	return &client{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  bufSize,
			WriteBufferSize: bufSize,
		}}
}

func (c *client) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	for {
		// Read message from the WebSocket
		messageType, message, err := conn.ReadMessage()
		slog.Info(fmt.Sprintf("message type: %d", messageType))

		if err != nil {
			slog.Error("Error reading message from WebSocket:", err)
			break
		}

		// Print the received message
		slog.Info(fmt.Sprintf("Received message: %s\n", message))

		// Echo the message back to the client
		if err := conn.WriteMessage(messageType, message); err != nil {
			slog.Error("Error echoing message:", err)
			break
		}
	}
}
