package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	http.HandleFunc("/chat", handleWebSocket)

	server := &http.Server{Addr: ":8080"}
	handleSigTerms(server)

	slog.Info("Starting a chat server at port: 8080")
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		slog.Error("ListenAndServe() error", err)
		os.Exit(1)
	}

}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	for {
		// Read message from the WebSocket
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			slog.Error("Error reading message from WebSocket:", err)
			break
		}

		// Print the received message
		slog.Info(fmt.Sprintf("Received message: %s\n", message))

		// Echo the message back to the client
		if err := conn.WriteMessage(messageType, message); err != nil {
			slog.Error("Error echoing message:", err)
			break
		}
	}
}

// For gracefully shutdown server
func handleSigTerms(server *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		slog.Info("received SIGTERM, exiting...")

		err := server.Shutdown(context.Background())
		if err != nil {
			slog.Error("server.Shutdown() error", err)
		}

		os.Exit(1)
	}()
}
