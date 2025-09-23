package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/caarlos0/env/v10"
	"golang.org/x/net/webdav"

	"github.com/jacobbrewer1/nas-os/pkg/storage"
	"github.com/jacobbrewer1/nas-os/pkg/storage/dav"
	"github.com/jacobbrewer1/web"
	"github.com/jacobbrewer1/web/logging"
)

const (
	appName = "nas-os"
)

type (
	// AppConfig holds the application configuration settings.
	AppConfig struct {
		StorageMechanism string `env:"STORAGE_MECHANISM" envDefault:"local"`        // e.g., "local", "s3"
		StoragePath      string `env:"STORAGE_PATH" envDefault:"./brewer/data/nas"` // Path for local storage or S3 bucket name
		Port             string `env:"PORT" envDefault:"8080"`                      // Port to run the server on
	}

	// App represents the main application structure.
	App struct {
		base *web.App
		cfg  *AppConfig
	}
)

// NewApp creates a new instance of App.
func NewApp(l *slog.Logger) (*App, error) {
	base, err := web.NewApp(l)
	if err != nil {
		return nil, fmt.Errorf("failed to create web app: %w", err)
	}

	cfg := new(AppConfig)
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return &App{
		base: base,
		cfg:  cfg,
	}, nil
}

// Start starts the application.
func (a *App) Start() error {
	var davHandler *webdav.Handler

	if err := a.base.Start(
		web.WithDependencyBootstrap(func(ctx context.Context) error {
			// If local dav is used, ensure the dav path exists
			if !storage.IsLocalStorage(a.cfg.StorageMechanism) {
				return nil
			}

			if err := storage.EnsureLocalPath(a.cfg.StoragePath); err != nil {
				return fmt.Errorf("failed to ensure local dav path: %w", err)
			}

			a.base.Logger().Info("local dav path ensured", slog.String("path", a.cfg.StoragePath))
			return nil
		}),
		web.WithDependencyBootstrap(func(ctx context.Context) error {
			// Initialize dav handler
			storageHandler, err := dav.HandlerFromString(a.cfg.StorageMechanism, a.cfg.StoragePath)
			if err != nil {
				return fmt.Errorf("failed to create dav handler: %w", err)
			}
			a.base.Logger().Info("dav handler initialized",
				slog.String("mechanism", a.cfg.StorageMechanism),
				slog.String("path", a.cfg.StoragePath),
			)

			l := a.base.Logger().With(slog.String("component", "webdav"))

			davHandler = &webdav.Handler{
				Prefix:     "/",
				FileSystem: storageHandler.FileSystem(),
				LockSystem: storageHandler.LockSystem(),
				Logger: func(request *http.Request, err error) {
					if err != nil {
						l.Error("webdav error",
							slog.String("method", request.Method),
							slog.String("url", request.URL.String()),
							slog.Any(logging.KeyError, err),
						)
					} else {
						l.Info("webdav request",
							slog.String("method", request.Method),
							slog.String("url", request.URL.String()),
						)
					}
				},
			}
			return nil
		}),
	); err != nil {
		return fmt.Errorf("failed to start web app: %w", err)
	}

	if err := a.base.StartServer("dav-server", &http.Server{
		Addr:              ":" + a.cfg.Port,
		Handler:           davHandler,
		ReadHeaderTimeout: 10 * time.Second,
	}); err != nil {
		return fmt.Errorf("failed to start NAS server: %w", err)
	}

	return nil
}

// WaitForEnd waits for the application to finish.
func (a *App) WaitForEnd() {
	a.base.WaitForEnd(a.Shutdown)
}

// Shutdown gracefully shuts down the application.
func (a *App) Shutdown() {
	a.base.Shutdown()
}

func main() {
	l := logging.NewLogger(
		logging.WithAppName(appName),
	)

	app, err := NewApp(l)
	if err != nil {
		l.Error("failed to create application",
			slog.Any(logging.KeyError, err),
		)
		panic(err)
	}

	if err := app.Start(); err != nil {
		l.Error("failed to start application",
			slog.Any(logging.KeyError, err),
		)
		panic(err)
	}

	app.WaitForEnd()
}
