package main

import (
	"context"
	"github.com/foohq/foojank/internal/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err := server.New().Run(ctx, os.Args)
	if err != nil {
		// Error logging is done inside each command no need to have a logger in this place.
		os.Exit(1)
	}
}
