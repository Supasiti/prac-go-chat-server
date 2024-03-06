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
	hub    *hub
	conn   *websocket.Conn
	send   chan *chatMessage
}

type chatMessage struct {
	From *client
	Msg  []byte
}

func ServeWebSocket(hub *hub, w http.ResponseWriter, r *http.Request) {

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("error upgrading to WebSocket:", slog.Any("err", err))
		return
	}

	c := &client{
		hub:    hub,
		conn:   conn,
		send:   make(chan *chatMessage),
	}
    c.hub.register <- c
    
	go c.readMessage()
	go c.writeMessage()
}

func (c *client) readMessage() {
	defer func() {
		slog.Info("closing websocket connection...")
        c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
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
		c.hub.notify <- &chatMessage{From: c, Msg: message}
	}
}

func (c *client) writeMessage() {
	defer func() {
		slog.Info("closing websocket connection...")
        c.hub.unregister <- c
        c.conn.Close()
	}()

	for {
		toSend, ok := <-c.send
		if !ok {
			// close channel
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		// Don't write if it is from itself
		if toSend.From == c {
			continue
		}

		if err := c.conn.WriteMessage(websocket.TextMessage, toSend.Msg); err != nil {
			slog.Error("error sending message:", slog.Any("err", err))
			return
		}
	}
}
