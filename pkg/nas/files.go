package nas

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/jacobbrewer1/nas-os/pkg/storage"
	"github.com/jacobbrewer1/web/logging"
)

// downloadFile handles file download requests
func downloadFile(backend storage.Backend) SMBHandlerFunc {
	return func(ctx context.Context, l *slog.Logger, w http.ResponseWriter, r *http.Request) {
		path := getPath(r)

		reader, err := backend.ReadFile(path)
		if err != nil {
			l.Warn("File not found",
				slog.String("path", path),
				slog.Any(logging.KeyError, err),
			)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		defer func() {
			if err := reader.Close(); err != nil {
				l.Warn("Error closing file reader", slog.Any(logging.KeyError, err))
			}
		}()

		// Set appropriate headers
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(path)))

		if _, err := io.Copy(w, reader); err != nil { // nolint:revive // Bad lint
			l.Error("Error sending file",
				slog.String("path", path),
				slog.Any(logging.KeyError, err),
			)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// uploadFile handles file upload requests
func uploadFile(backend storage.Backend) SMBHandlerFunc {
	return func(ctx context.Context, l *slog.Logger, w http.ResponseWriter, r *http.Request) {
		path := getPath(r)

		if err := backend.WriteFile(path, r.Body); err != nil {
			l.Error("Error uploading file",
				slog.String("path", path),
				slog.Any(logging.KeyError, err),
			)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// deleteFile handles file deletion requests
func deleteFile(backend storage.Backend) SMBHandlerFunc {
	return func(ctx context.Context, l *slog.Logger, w http.ResponseWriter, r *http.Request) {
		path := getPath(r)

		if err := backend.DeleteFile(path); err != nil {
			l.Error("Error deleting file",
				slog.String("path", path),
				slog.Any(logging.KeyError, err),
			)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
