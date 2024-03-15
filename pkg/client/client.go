package client

import (
	"bytes"
	"log"

	"github.com/gorilla/websocket"
)

type client struct {
	conn *websocket.Conn
}

type OnRead func(interface{})

func NewClient(conn *websocket.Conn) *client {
	return &client{conn: conn}
}

// Running read in a separate goroutine
func (c *client) StartListening(onRead OnRead) {
	defer c.conn.Close()

	log.Println("ws: start listening...")
    var from []byte
    var msg []byte
    var err error

	for {
		_, msg, err = c.conn.ReadMessage()
		log.Println("ws: message ", string(msg))
		if err != nil {
			log.Println("ws: err ", err)
			onRead(errMsg(err))
			return
		}
        
        // message is of the format <username>:<message>
		colonIndex := bytes.IndexByte(msg, ':')

		if colonIndex == -1 {
            log.Fatal("message is incorrect format")
        }

        from = msg[:colonIndex]
        msg = msg[colonIndex+1:]

		onRead(chatMsg{msg: string(msg), from: string(from)})
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
func (c *client) Send(text string) error {
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(text))
	if err != nil {
		c.Close()
		return err
	}
	return nil
}

// Close the underlying websocket connection
func (c *client) Close() error {
	return c.conn.Close()
}
