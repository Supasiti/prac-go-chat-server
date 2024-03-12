package client

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

type client struct {
	conn *websocket.Conn

    wg sync.WaitGroup
}

func NewClient(conn *websocket.Conn) *client {
	return &client{conn: conn}
}

func (c *client) Start() {
    c.wg.Add(2)

	go c.read()
	go c.send()
    
    c.wg.Wait()
}

func (c *client) read() {
	defer c.conn.Close()
    defer c.wg.Done()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			fmt.Println("exiting chat room")
			return
		}
		fmt.Printf("Received: %s\n", message)
	}
}

// See here https://github.com/gorilla/websocket/issues/720
// for discussion around closing connection properly
// In summary, ask the following questions
//  1. Do you want to specify the close code and reason reported to the peer applications?
//  2. Do you want to be able to read all messages written by the
//     peer - 'clean' close. If it's OK to ignore a message
//     written by the peer, then you don't need a clean close.
//
// Algorithm for closing is
//   - if the answer to (1) and (2) are yes, initiate a close message
//   - if the answer to (2) is yes, then wait with timeout for the peer to response with
//     a close message
//   - else close the connection
//
// In most cases, the last option is fine
func (c *client) send() {
	defer c.conn.Close()
    defer c.wg.Done()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter message to send (type 'exit' to quit): ")
		text, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		if text == "exit\n" {
			return
		}

		// Send the message to the WebSocket server
		err = c.conn.WriteMessage(websocket.TextMessage, []byte(text))
		if err != nil {
			fmt.Println("send error:", err)
			return
		}
	}
}
