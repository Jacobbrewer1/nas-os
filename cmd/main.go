package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/gorilla/mux"

	"github.com/jacobbrewer1/nas-os/pkg/nas"
	"github.com/jacobbrewer1/nas-os/pkg/storage"
	smbweb "github.com/jacobbrewer1/nas-os/pkg/web"
	"github.com/jacobbrewer1/web"
	"github.com/jacobbrewer1/web/health"
	"github.com/jacobbrewer1/web/logging"
)

const (
	// appName is the name of the application.
	appName = "smb-gateway"
)

type (
	// AppConfig holds application configuration
	AppConfig struct {
		SmbPort       string `env:"SMB_PORT" envDefault:"445"`
		SmbBindAddr   string `env:"SMB_BIND_ADDR" envDefault:"127.0.0.1"`
		WebServerPort string `env:"WEB_SERVER_PORT" envDefault:"8080"`
		WebServerAddr string `env:"WEB_SERVER_ADDR" envDefault:"127.0.0.1"`
	}

	// App represents the main application
	App struct {
		cfg  *AppConfig
		base *web.App
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
	smbRouter := mux.NewRouter()
	webSvr := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := a.base.Start(
		web.WithDependencyBootstrap(func(ctx context.Context) error {
			storageBackend := storage.ProviderFromENV(
				logging.LoggerWithComponent(a.base.Logger(), "storage_backend"),
			)

			smbServer := nas.NewSMBServer(
				logging.LoggerWithComponent(a.base.Logger(), "nas_server"),
				storageBackend,
			)
			smbServer.SetupRoutes(smbRouter)

			webSvr = smbweb.NewServer(
				logging.LoggerWithComponent(a.base.Logger(), "smb_web_server"),
				a.cfg.WebServerAddr,
				a.cfg.WebServerPort,
				net.JoinHostPort(a.cfg.SmbBindAddr, a.cfg.SmbPort),
				storageBackend.Name(),
			)
			return nil
		}),
		web.WithHealthCheck(health.NewCheck("ok", func(ctx context.Context) error {
			return nil
		})),
	); err != nil {
		return fmt.Errorf("failed to start web app: %w", err)
	}

	if err := a.base.StartServer("smb-server", &http.Server{
		Addr:              net.JoinHostPort(a.cfg.SmbBindAddr, a.cfg.SmbPort),
		Handler:           smbRouter,
		ReadHeaderTimeout: 10 * time.Second,
	}); err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}

	if err := a.base.StartServer("smb-web-server", webSvr); err != nil {
		return fmt.Errorf("failed to start web server: %w", err)
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
