package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/foohq/foojank/internal/commands"
)

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	return commands.NewCommand().Run(ctx, os.Args)
}

func main() {
	err := run()
	if err != nil {
		// Error logging is done inside each command no need to have a logger in this place.
		os.Exit(1)
	}
}
