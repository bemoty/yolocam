package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/bemoty/yolocam/cmd/yolocam/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := cli.Execute(ctx)
	stop()
	if err != nil {
		os.Exit(1)
	}
}
