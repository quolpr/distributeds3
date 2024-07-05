package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type App struct {
	ServiceProvider *serviceProvider
}

func NewApp(ctx context.Context) (*App, error) {
	serviceProvider, err := NewServiceProvider(ctx)

	if err != nil {
		return nil, fmt.Errorf("error while create service provider: %w", err)
	}

	return &App{
		ServiceProvider: serviceProvider,
	}, nil
}

func (app *App) Logger() *slog.Logger {
	return app.ServiceProvider.Logger
}

func (app *App) ServeHTTP(ctx context.Context) error {
	logger := app.ServiceProvider.Logger

	port := ":8080"

	httpServer := &http.Server{ //nolint:exhaustruct
		Addr:         port,
		ReadTimeout:  10 * time.Second, //nolint:gomnd
		WriteTimeout: 30 * time.Second, //nolint:gomnd
		Handler:      newRoutes(app.ServiceProvider),
	}

	serverErrorCh := make(chan error)
	go func() {
		defer close(serverErrorCh)

		logger.Info("Server started", "port", port)

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
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30) //nolint:gomnd
		defer cancel()
		//nolint: contextcheck
		return httpServer.Shutdown(ctx) //nolint:wrapcheck
	}
}
