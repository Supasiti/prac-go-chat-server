package main

import (
	"bytes"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	bufSize = 1024
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  bufSize,
		WriteBufferSize: bufSize,
	}
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type client struct {
	readCh chan []byte
	conn   *websocket.Conn
}

func ServeWebSocket(w http.ResponseWriter, r *http.Request) {

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("error upgrading to WebSocket:", slog.Any("err", err))
		return
	}

	c := &client{
		readCh: make(chan []byte, bufSize),
		conn:   conn,
	}

	go c.readMessage()
	go c.writeMessage()
}

func (c *client) readMessage() {
	defer func() {
		slog.Info("closing websocket connection...")
		if c.conn != nil {
			c.conn.Close()
		}
	}()

	for {
		// Read message from the WebSocket
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				slog.Error("client closes connection...", slog.Any("err", err))
				return
			}

			slog.Error("error reading message...", slog.Any("err", err))
			return
		}

		message = bytes.TrimSpace(message)
		slog.Info("received message:", slog.String("message", string(message)))
		c.readCh <- message
	}
}

func (c *client) writeMessage() {
	defer func() {
		slog.Info("closing websocket connection...")
		if c.conn != nil {
			c.conn.Close()
		}
	}()

	for {
		message, ok := <-c.readCh
		if !ok {
			// close channel
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			slog.Error("error echoing message:", slog.Any("err", err))
			return
		}
	}
}
