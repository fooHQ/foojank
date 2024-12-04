package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/foohq/foojank/internal/client"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err := client.New().Run(ctx, os.Args)
	if err != nil {
		// Error logging is done inside each command no need to have a logger in this place.
		os.Exit(1)
	}
}
