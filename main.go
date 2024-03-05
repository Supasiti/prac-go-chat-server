package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
    client := NewClient()
	http.HandleFunc("/chat", client.handleWebSocket)

	server := &http.Server{Addr: ":8080"}
	handleSigTerms(server)

	slog.Info("Starting a chat server at port: 8080")
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		slog.Error("ListenAndServe() error", err)
		os.Exit(1)
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
