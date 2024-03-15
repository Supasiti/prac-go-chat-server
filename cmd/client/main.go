package main

import (
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/supasiti/prac-go-chat-server/pkg/client"
)

const (
	wsUrl = "ws://127.0.0.1:8080/chat"
)

var (
	debug = os.Getenv("DEBUG") == "1"
)

func main() {
	log.SetOutput(io.Discard)
	if debug {
		fmt.Println("Debug mode")
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			log.Fatal("fatal:", err)
		}
		defer f.Close()
	}

	if len(os.Args) < 2 {
		fmt.Println("Missing argument! Please provide a name.")
		return
	}

	username := os.Args[1]
	conn, err := client.NewWsConnection(wsUrl, username)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	c := client.NewClient(conn)
	m := client.NewModel(c)
	p := tea.NewProgram(m)

	// start listening on ws
	go c.StartListening(client.HandleOnRead(p))

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
