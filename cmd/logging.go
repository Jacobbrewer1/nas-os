package main

import (
	"log/slog"
	"net/http"

	"github.com/jacobbrewer1/web/logging"
)

func davRequestLogger(l *slog.Logger) func(*http.Request, error) {
	return func(r *http.Request, err error) {
		if err != nil {
			l.Error("webdav error",
				slog.String("method", r.Method),
				slog.String("url", r.URL.String()),
				slog.Any(logging.KeyError, err),
			)
			return
		}

		l.Debug("webdav request",
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
		)
	}
}
