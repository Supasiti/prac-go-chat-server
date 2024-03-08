package main

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/supasiti/prac-go-chat-server/pkg/client"
)

const (
	wsUrl = "ws://127.0.0.1:8080/chat"
)

func main() {
	// Establish WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()
    
    c:= client.NewClient(conn)
    c.Start()
}
