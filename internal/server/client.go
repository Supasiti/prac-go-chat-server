package server

import (
	"bytes"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// maximum message size
	maxMsgSize = 512

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type token struct{}

type client struct {
	id       int
	username string
	hub      *hub
	conn     *websocket.Conn
	send     chan *chatMessage
	done     chan *token
	wg       sync.WaitGroup
}

type chatMessage struct {
	FromId int
	From   string
	Msg    []byte
}

func NewClient(hub *hub, conn *websocket.Conn, username string) *client {
	return &client{
		id:       generateId(),
		username: username,
		hub:      hub,
		conn:     conn,
		send:     make(chan *chatMessage),
		done:     make(chan *token),
	}
}

func (c *client) Start() {
	c.hub.register <- c

	c.wg.Add(2)
	go c.read()
	go c.write()

	c.wg.Wait()
}

func (c *client) read() {
	defer c.cleanUp()
	defer func() { close(c.done) }()

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
				slog.Info("client closes connection...", slog.Any("err", err))
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
			slog.String("from", c.username))

		c.hub.notify <- &chatMessage{FromId: c.id, From: c.username, Msg: message}
	}
}

func (c *client) write() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	defer c.cleanUp()

	for {
		select {
		case <-c.done:
			return
		case toSend, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			// sending messages
			if !ok {
				slog.Error("closing sending channel")
				return
			}

			// Don't write if it is from itself
			if toSend.FromId == c.id {
				continue
			}

			// message sent will be of the format:
			// <username>:<message>
			message := append([]byte(toSend.From), ':')
			message = append(message, toSend.Msg...)

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
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

func (c *client) cleanUp() {
	c.hub.unregister <- c
	c.conn.Close()
	c.wg.Done()
}

func generateId() int {
	return int(time.Now().UnixMilli())
}
