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

	provider, err := app.NewServiceProvider(ctx)

	if err != nil {
		panic(err.Error())
	}

	err = provider.UploadSvc.CleanDangleUploads(ctx)
	if err != nil {
		panic(err.Error())
	}

	provider.Logger.Info("Dangle uploads cleaned")
}
