package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/quolpr/distributeds3/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	app, err := app.NewApp(ctx)

	if err != nil {
		panic(err)
	}

	logger := app.Logger()

	err = app.ServeHTTP(ctx)
	if err != nil {
		logger.Error("Server closed", "error", err)

		defer os.Exit(1)

		return
	}

	logger.Info("Server closed")
}
