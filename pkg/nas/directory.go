package nas

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"

	"github.com/jacobbrewer1/nas-os/pkg/storage"
	"github.com/jacobbrewer1/web/logging"
)

// handleUpload handles file upload requests via multipart form data
func handleUpload(backend storage.Backend, maxUploadSize int64) SMBHandlerFunc {
	return func(ctx context.Context, l *slog.Logger, w http.ResponseWriter, r *http.Request) {
		path := getPath(r)

		// Parse multipart form
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			l.Warn("Error parsing multipart form", slog.Any(logging.KeyError, err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			l.Warn("Error retrieving file from form data", slog.Any(logging.KeyError, err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer func() {
			if err := file.Close(); err != nil {
				l.Warn("Error closing uploaded file", slog.Any(logging.KeyError, err))
			}
		}()

		// Construct full file path
		fullPath := filepath.Join(path, header.Filename)
		fullPath = filepath.ToSlash(fullPath) // Normalize to forward slashes

		if err := backend.WriteFile(fullPath, file); err != nil {
			l.Error("Error uploading file",
				slog.String("path", fullPath),
				slog.Any(logging.KeyError, err),
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fileInfo, err := backend.GetFileInfo(fullPath)
		if err != nil {
			l.Error("Error retrieving file info after upload",
				slog.String("path", fullPath),
				slog.Any(logging.KeyError, err),
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(fileInfo); err != nil {
			l.Warn("Error encoding upload response", slog.Any(logging.KeyError, err))
			return
		}
	}
}

// listDirectory handles directory listing requests
func listDirectory(backend storage.Backend) SMBHandlerFunc {
	return func(ctx context.Context, l *slog.Logger, w http.ResponseWriter, r *http.Request) {
		path := getPath(r)

		entries, err := backend.ListDirectory(path)
		if err != nil {
			l.Warn("Error listing directory",
				slog.String("path", path),
				slog.Any(logging.KeyError, err),
			)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if len(entries) == 0 {
			entries = make([]storage.FileInfo, 0)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(entries); err != nil {
			l.Warn("Error encoding directory listing response", slog.Any(logging.KeyError, err))
			return
		}
	}
}

// createDirectory handles directory creation requests
func createDirectory(backend storage.Backend) SMBHandlerFunc {
	return func(ctx context.Context, l *slog.Logger, w http.ResponseWriter, r *http.Request) {
		path := getPath(r)

		if err := backend.CreateDirectory(path); err != nil {
			l.Error("Error creating directory",
				slog.String("path", path),
				slog.Any(logging.KeyError, err),
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// handleDelete handles file or directory deletion requests
func handleDelete(backend storage.Backend) SMBHandlerFunc {
	return func(ctx context.Context, l *slog.Logger, w http.ResponseWriter, r *http.Request) {
		path := getPath(r)

		if err := backend.DeleteFile(path); err != nil {
			// Might not be a file, try deleting as directory
			if derr := backend.DeleteDirectory(path); derr != nil {
				l.Error("Error deleting file or directory",
					slog.String("path", path),
					slog.Any(logging.KeyError, err),
					slog.Any("delete_error", derr),
				)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// handleGetInfo handles requests to get file or directory info
func handleGetInfo(backend storage.Backend) SMBHandlerFunc {
	return func(ctx context.Context, l *slog.Logger, w http.ResponseWriter, r *http.Request) {
		path := getPath(r)

		info, err := backend.GetFileInfo(path)
		if err != nil {
			l.Warn("Error retrieving file info",
				slog.String("path", path),
				slog.Any(logging.KeyError, err),
			)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(info); err != nil {
			l.Warn("Error encoding file info response", slog.Any(logging.KeyError, err))
			return
		}
	}
}

func getPath(r *http.Request) string {
	path := mux.Vars(r)["path"]
	if path == "" {
		path = "/"
	}
	// Remove all spaces from path to avoid issues
	path = strings.ReplaceAll(path, " ", "_")
	return path
}
