package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/foohq/foojank/cmd/foojankd/commands"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	err := commands.NewCommand().Run(ctx, os.Args)
	if err != nil {
		cancel()
		os.Exit(1)
	}
	cancel()
}
