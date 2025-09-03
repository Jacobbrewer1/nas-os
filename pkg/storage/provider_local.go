package storage

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

var _ Backend = (*Local)(nil)

// Local implements Backend interface for local filesystem storage
type Local struct {
	// logger is used for logging operations
	logger *slog.Logger

	// basePath is the root directory for local storage operations
	basePath string
}

// NewLocal creates a new local storage backend
func NewLocal(l *slog.Logger, basePath string) *Local {
	// Does the base path exist? If not, create it
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		if err := os.MkdirAll(basePath, 0o755); err != nil {
			l.Error("Failed to create base path for local storage",
				slog.String("base_path", basePath),
				slog.Any("error", err),
			)
		}
	} else if err != nil {
		l.Error("Failed to stat base path for local storage",
			slog.String("base_path", basePath),
			slog.Any("error", err),
		)
	}

	return &Local{
		logger:   l,
		basePath: basePath,
	}
}

func (l Local) Name() string {
	return "local"
}

func (l Local) ReadFile(path string) (io.ReadCloser, error) {
	l.logger.Debug("Reading file",
		slog.String("path", path),
	)

	// Read the file from the local filesystem
	fullPath := filepath.Join(l.basePath, path)
	fullPath = cleanPath(fullPath)

	file, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return io.NopCloser(bytes.NewReader(file)), nil
}

func (l Local) WriteFile(path string, data io.Reader) error {
	l.logger.Debug("Writing file",
		slog.String("path", path),
	)

	// Write the file to the local filesystem
	content, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	fullPath := filepath.Join(l.basePath, path)
	fullPath = cleanPath(fullPath)

	// Ensure the directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	if err := os.WriteFile(fullPath, content, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func (l Local) DeleteFile(path string) error {
	l.logger.Debug("Deleting file",
		slog.String("path", path),
	)

	// Delete the file from the local filesystem
	fullPath := filepath.Join(l.basePath, path)
	fullPath = cleanPath(fullPath)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (l Local) ListDirectory(path string) ([]FileInfo, error) {
	l.logger.Debug("Listing directory",
		slog.String("path", path),
	)

	fullPath := filepath.Join(l.basePath, path)
	fullPath = cleanPath(fullPath)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to get file info: %w", err)
		}

		files = append(files, FileInfo{
			Name:        entry.Name(),
			Path:        filepath.Join(path, entry.Name()),
			Size:        info.Size(),
			IsDirectory: entry.IsDir(),
			ModTime:     info.ModTime(),
			Mode:        uint32(info.Mode().Perm()),
		})
	}

	return files, nil
}

func (l Local) CreateDirectory(path string) error {
	l.logger.Debug("Creating directory",
		slog.String("path", path),
	)

	fullPath := filepath.Join(l.basePath, path)
	fullPath = cleanPath(fullPath)
	if err := os.MkdirAll(fullPath, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

func (l Local) DeleteDirectory(path string) error {
	l.logger.Debug("Deleting directory",
		slog.String("path", path),
	)

	fullPath := filepath.Join(l.basePath, path)
	fullPath = cleanPath(fullPath)

	if err := os.RemoveAll(fullPath); err != nil {
		return fmt.Errorf("failed to delete directory: %w", err)
	}
	return nil
}

func (l Local) GetFileInfo(path string) (*FileInfo, error) {
	l.logger.Debug("Getting file info",
		slog.String("path", path),
	)

	fullPath := filepath.Join(l.basePath, path)
	fullPath = cleanPath(fullPath)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileInfo{
		Name:        info.Name(),
		Path:        cleanPath(path),
		Size:        info.Size(),
		IsDirectory: info.IsDir(),
		ModTime:     info.ModTime(),
		Mode:        uint32(info.Mode().Perm()),
	}, nil
}

func (l Local) FileExists(path string) (bool, error) {
	l.logger.Debug("Checking if file exists",
		slog.String("path", path),
	)

	fullPath := filepath.Join(l.basePath, path)
	fullPath = cleanPath(fullPath)

	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	return true, nil
}
