package client

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/gorilla/websocket"
	"github.com/supasiti/prac-go-chat-server/pkg/errgroup"
)

type client struct {
	conn *websocket.Conn
}

func NewClient(conn *websocket.Conn) *client {
	return &client{conn: conn}
}

func (c *client) Start() {
	g := errgroup.WithContext(context.Background())

	g.Go(c.read)
	g.Go(c.send)

	// will close connection in all cases
	g.Wait()
    if err := c.conn.Close(); err != nil {
		return
	}
}

func (c *client) read(ctx context.Context) error {
	for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
                fmt.Println("exiting chat room")
                return err
            }
            return err
        } 
        fmt.Printf("Received: %s\n", message)
	}
}

func (c *client) send(ctx context.Context) error {
	defer c.conn.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter message to send (type 'exit' to quit): ")
			text, err := reader.ReadString('\n')
			if err != nil {
				return err
			}

			if text == "exit\n" {
				err = c.conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(
						websocket.CloseNormalClosure,
						"Leaving chat room",
					),
				)
				if err != nil {
					return err
				}
				return nil
			}

			// Send the message to the WebSocket server
			err = c.conn.WriteMessage(websocket.TextMessage, []byte(text))
			if err != nil {
				fmt.Println("send error:", err)
				return err
			}
		}
	}
}
