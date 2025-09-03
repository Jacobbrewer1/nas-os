package storage

import (
	"io"
	"path/filepath"
	"strings"
	"time"
)

// Backend defines the interface for cloud storage providers
type Backend interface {
	// Name returns the name of the storage backend
	Name() string

	// ReadFile reads a file from the storage backend
	ReadFile(path string) (io.ReadCloser, error)
	// WriteFile writes data to a file in the storage backend
	WriteFile(path string, data io.Reader) error
	// DeleteFile deletes a file from the storage backend
	DeleteFile(path string) error

	// ListDirectory lists files in a directory
	ListDirectory(path string) ([]FileInfo, error)
	// CreateDirectory creates a new directory
	CreateDirectory(path string) error
	// DeleteDirectory deletes a directory
	DeleteDirectory(path string) error

	// GetFileInfo retrieves metadata about a file or directory
	GetFileInfo(path string) (*FileInfo, error)
	// FileExists checks if a file or directory exists
	FileExists(path string) (bool, error)
}

// FileInfo represents file/directory metadata
type FileInfo struct {
	Name        string    `json:"name,omitempty"`
	Path        string    `json:"path,omitempty"`
	Size        int64     `json:"size,omitempty"`
	IsDirectory bool      `json:"is_directory,omitempty"`
	ModTime     time.Time `json:"mod_time"`
	Mode        uint32    `json:"mode,omitempty"` // File permissions
}

func cleanPath(path string) string {
	path = filepath.Clean(path)
	path = strings.TrimSpace(path)
	path = strings.ReplaceAll(path, " ", "_") // Replace spaces with underscores
	return path
}
