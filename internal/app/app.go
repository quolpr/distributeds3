package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type App struct {
	ServiceProvider *serviceProvider
}

func NewApp(ctx context.Context) *App {
	return &App{
		ServiceProvider: newServiceProvider(ctx),
	}
}

func (app *App) Logger() *slog.Logger {
	return app.ServiceProvider.Logger
}

func (app *App) ServeHTTP(ctx context.Context) error {
	logger := app.ServiceProvider.Logger

	httpServer := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler:      newRoutes(app.ServiceProvider),
	}

	serverErrorCh := make(chan error)
	go func() {
		defer close(serverErrorCh)

		logger.Info("Server started", "port", 8080)
		err := httpServer.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server closing", "error", err)
		}

		select {
		case serverErrorCh <- err:
			return
		default:
			return
		}
	}()

	select {
	case err := <-serverErrorCh:
		return err
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		//nolint: contextcheck
		return httpServer.Shutdown(ctx)
	}
}
