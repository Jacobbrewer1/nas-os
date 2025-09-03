package storage

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProviderFromENV(t *testing.T) {
	tests := []struct {
		name     string
		envVal   string
		expected string
	}{
		{
			name:     "Memory Backend",
			envVal:   "memory",
			expected: "memory",
		},
		{
			name:     "Mock Backend",
			envVal:   "mock",
			expected: "memory",
		},
		{
			name:     "Local Backend",
			envVal:   "local",
			expected: "local",
		},
		{
			name:     "Unsupported Backend",
			envVal:   "unsupported",
			expected: "panic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the environment variable
			t.Setenv("PROVIDER", tt.envVal)

			if tt.expected == "panic" {
				require.Panics(t, func() {
					ProviderFromENV(slog.New(slog.DiscardHandler))
				})
			} else {
				require.NotPanics(t, func() {
					backend := ProviderFromENV(slog.New(slog.DiscardHandler))
					require.Equal(t, tt.expected, backend.Name())
				})
			}
		})
	}
}
