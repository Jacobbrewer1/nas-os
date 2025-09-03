package storage

import (
	"log/slog"
	"strings"

	"github.com/caarlos0/env/v10"
)

type (
	// config holds storage configuration from environment
	config struct {
		Provider string `env:"PROVIDER" envDefault:"memory"` // "gcs", "azure", "aws", etc.
		BasePath string `env:"STORAGE_BASE_PATH" envDefault:"./data"`
	}
)

// ProviderFromENV creates a storage backend based on environment configuration
func ProviderFromENV(
	l *slog.Logger,
) Backend {
	cfg := new(config)
	if err := env.Parse(cfg); err != nil {
		l.Error("Failed to parse storage configuration from environment",
			slog.Any("error", err),
		)
	}

	switch strings.ToLower(cfg.Provider) {
	case "memory", "mock":
		return NewMemoryBackend(l)
	case "local":
		return NewLocal(l, cfg.BasePath)
	default:
		panic("unsupported storage provider: " + cfg.Provider)
	}
}
