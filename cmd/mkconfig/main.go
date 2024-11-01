package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app := NewApplication()
	err := app.Run(ctx, os.Args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
}
