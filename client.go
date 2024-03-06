package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// maximum message size
	maxMsgSize = 512

	// buffer size for web socket. Must be bigger than maxMsgSize
	bufSize = 2 * maxMsgSize

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  bufSize,
		WriteBufferSize: bufSize,
	}
)

type client struct {
	id   int
	hub  *hub
	conn *websocket.Conn
	send chan *chatMessage
}

type chatMessage struct {
	From int
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
		id:   generateId(),
		hub:  hub,
		conn: conn,
		send: make(chan *chatMessage),
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

	c.conn.SetReadLimit(maxMsgSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				slog.Error("client closes connection...", slog.Any("err", err))
				return
			}

			slog.Error("error reading message...", slog.Any("err", err))
			return
		}

		if messageType != websocket.TextMessage {
			slog.Error("incorrect message type")
			continue
		}

		message = bytes.TrimSpace(message)
		slog.Info("received message:",
			slog.String("message", string(message)),
			slog.Int("from", c.id))

		c.hub.notify <- &chatMessage{From: c.id, Msg: message}
	}
}

func (c *client) writeMessage() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		slog.Info("closing websocket connection...")
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case toSend, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			// sending messages
			if !ok {
				// close channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Don't write if it is from itself
			if toSend.From == c.id {
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, toSend.Msg); err != nil {
				slog.Error("error sending message:", slog.Any("err", err))
				return
			}
		case <-ticker.C:
			// regularly check connection
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("error pinging:", slog.Any("err", err))
				return
			}
		}
	}
}

func generateId() int {
	return int(time.Now().UnixMilli())
}
