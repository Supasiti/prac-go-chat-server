package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/supasiti/prac-go-chat-server/internal/server"
)

func main() {
	hub := server.NewHub()
	go hub.Run()

	http.HandleFunc("/chat", hub.Serve)

	server := &http.Server{Addr: ":8080"}
	handleSigTerms(server)

	slog.Info("starting a chat server at port: 8080")
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("ListenAndServe() error", slog.Any("err", err))
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
			slog.Error("server.Shutdown() error: ", slog.Any("err", err))
		}

		os.Exit(1)
	}()
}
