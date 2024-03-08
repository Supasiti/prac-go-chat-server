package client

import (
	"bufio"
	"context"
	"errors"
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

	g.Go(c.readMessage)
	g.Go(c.sendMessage)

	// will close connection in all cases
	err := g.Wait()

	fmt.Printf("closing websocket connection... %s\n", err)
	if err = c.conn.Close(); err != nil {
        return
    }
}

func (c *client) readMessage(ctx context.Context) error {
    defer func() {
        fmt.Println("close read")
    }()

	for {
		select {
		case <-ctx.Done():
            fmt.Println("closing")
			return nil
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				fmt.Println("error reading message...")
				return err
			}
			fmt.Printf("Received: %s\n", message)
		}
	}
}

func (c *client) sendMessage(ctx context.Context) error {
    defer func() {
        fmt.Println("close send")
    }()

	for {
		select {
		case <-ctx.Done():
            fmt.Println("closing")
			return nil
		default:
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter message to send (type 'exit' to quit): ")
			text, err := reader.ReadString('\n')
            if err != nil {
                return err
            }

			if text == "exit\n" {
				return errors.New("closing client")
			}

			// Send the message to the WebSocket server
			err = c.conn.WriteMessage(websocket.TextMessage, []byte(text))
			if err != nil {
				fmt.Println("write:", err)
				return err
			}

		}
	}
}
