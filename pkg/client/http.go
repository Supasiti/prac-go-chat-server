package client

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
)

func NewWsConnection(url, username string) (*websocket.Conn, error) {

	header := http.Header{
		"Username": []string{username},
	}

	// Establish WebSocket connection
	conn, resp, err := websocket.DefaultDialer.Dial(url, header)
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return nil, err
		}

		fmt.Println(string(body))
		return nil, err
	}
	return conn, err
}
