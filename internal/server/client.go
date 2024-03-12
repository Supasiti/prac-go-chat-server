package server

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"github.com/supasiti/prac-go-chat-server/pkg/errgroup"
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

func NewClient(hub *hub, conn *websocket.Conn) *client {
	return &client{
		id:   generateId(),
		hub:  hub,
		conn: conn,
		send: make(chan *chatMessage),
	}
}

func (c *client) Start() {
	c.hub.register <- c

	g := errgroup.WithContext(context.Background())

	g.Go(c.readMessage)
	g.Go(c.writeMessage)

	// will close connection in all cases
	g.Wait()
	slog.Info("closing websocket connection...")
	c.hub.unregister <- c
	c.conn.Close()
}

func (c *client) readMessage(ctx context.Context) error {

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
				return err
			}
			slog.Error("error reading message...", slog.Any("err", err))
			return err
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

func (c *client) writeMessage(ctx context.Context) error {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case toSend, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			// sending messages
			if !ok {
				// close channel
				c.conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(
						websocket.CloseNormalClosure,
						"closing chat room",
					),
				)
				return errors.New("closing sending channel")
			}

			// Don't write if it is from itself
			if toSend.From == c.id {
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, toSend.Msg); err != nil {
				slog.Error("error sending message:", slog.Any("err", err))
				return err
			}
		case <-ticker.C:
			// regularly check connection
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("error pinging:", slog.Any("err", err))
				return err
			}
		}
	}
}

func generateId() int {
	return int(time.Now().UnixMilli())
}
