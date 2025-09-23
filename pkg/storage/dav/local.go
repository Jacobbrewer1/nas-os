package dav

import (
	"golang.org/x/net/webdav"
)

var _ Handler = (*Local)(nil)

// Local represents a local storage handler.
type Local struct {
	Path string
}

// NewLocal creates a new Local storage handler with the given path.
func NewLocal(path string) *Local {
	return &Local{
		Path: path,
	}
}

func (l Local) FileSystem() webdav.FileSystem {
	return webdav.Dir(l.Path)
}

func (l Local) LockSystem() webdav.LockSystem {
	return webdav.NewMemLS()
}
