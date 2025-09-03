package storage

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/jacobbrewer1/web/logging"
)

var _ Backend = (*MemoryBackend)(nil)

// MemoryBackend implements Backend interface for in-memory storage
type MemoryBackend struct {
	// logger is used for logging operations
	logger *slog.Logger

	// files is an in-memory storage for files
	files map[string][]byte
}

// NewMemoryBackend creates a new in-memory storage backend
func NewMemoryBackend(l *slog.Logger) *MemoryBackend {
	l = l.With(
		slog.String(logging.KeyComponent, "memory_backend"),
	)

	return &MemoryBackend{
		logger: l,
		files:  make(map[string][]byte),
	}
}

func (m MemoryBackend) Name() string {
	return "memory"
}

func (m MemoryBackend) ReadFile(path string) (io.ReadCloser, error) {
	m.logger.Debug("Reading file",
		slog.String("path", path),
	)

	path = cleanPath(path)

	data, exists := m.files[path]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m MemoryBackend) WriteFile(path string, data io.Reader) error {
	m.logger.Debug("Writing file",
		slog.String("path", path),
	)

	path = cleanPath(path)

	content, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	m.files[path] = content
	return nil
}

func (m MemoryBackend) DeleteFile(path string) error {
	m.logger.Debug("Deleting file",
		slog.String("path", path),
	)

	path = cleanPath(path)

	if _, exists := m.files[path]; !exists {
		return fmt.Errorf("file not found: %s", path)
	}

	delete(m.files, path)
	return nil
}

func (m MemoryBackend) ListDirectory(path string) ([]FileInfo, error) {
	m.logger.Debug("Listing directory",
		slog.String("path", path),
	)

	path = cleanPath(path)

	var files []FileInfo

	// Normalize path
	if path != "/" && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	// Find all files that start with this path
	for filePath := range m.files {
		if path == "/" {
			// Root directory - show top-level files and directories
			parts := strings.Split(strings.TrimPrefix(filePath, "/"), "/")
			if len(parts) == 1 && parts[0] != "" {
				files = append(files, FileInfo{
					Name:        parts[0],
					Path:        filePath,
					Size:        int64(len(m.files[filePath])),
					IsDirectory: false,
					ModTime:     time.Now(),
					Mode:        0o644,
				})
			}
		} else if strings.HasPrefix(filePath, path+"/") {
			// Files in subdirectory
			relativePath := strings.TrimPrefix(filePath, path+"/")
			parts := strings.Split(relativePath, "/")
			if len(parts) == 1 {
				files = append(files, FileInfo{
					Name:        parts[0],
					Path:        filePath,
					Size:        int64(len(m.files[filePath])),
					IsDirectory: false,
					ModTime:     time.Now(),
					Mode:        0o644,
				})
			}
		}
	}

	return files, nil
}

func (m MemoryBackend) CreateDirectory(path string) error {
	m.logger.Debug("Creating directory",
		slog.String("path", path),
	)

	// In-memory backend does not need to explicitly create directories
	return nil
}

func (m MemoryBackend) DeleteDirectory(path string) error {
	m.logger.Debug("Deleting directory",
		slog.String("path", path),
	)

	path = cleanPath(path)

	// Delete all files in the directory
	for filePath := range m.files {
		if strings.HasPrefix(filePath, path+"/") {
			delete(m.files, filePath)
		}
	}

	return nil
}

func (m MemoryBackend) GetFileInfo(path string) (*FileInfo, error) {
	m.logger.Debug("Getting file info",
		slog.String("path", path),
	)

	path = cleanPath(path)

	data, exists := m.files[path]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	// Extract filename from path
	name := path
	if idx := strings.LastIndex(path, "/"); idx != -1 {
		name = path[idx+1:]
	}

	return &FileInfo{
		Name:        name,
		Path:        path,
		Size:        int64(len(data)),
		IsDirectory: false,
		ModTime:     time.Now(),
		Mode:        0o644,
	}, nil
}

func (m MemoryBackend) FileExists(path string) (bool, error) {
	m.logger.Debug("Checking if file exists",
		slog.String("path", path),
	)

	path = cleanPath(path)

	_, exists := m.files[path]
	return exists, nil
}
