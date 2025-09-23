package dav

import (
	"fmt"
	"strings"

	"golang.org/x/net/webdav"
)

type Handler interface {
	FileSystem() webdav.FileSystem
	LockSystem() webdav.LockSystem
}

// HandlerFromString creates a storage handler based on the given mechanism and path.
func HandlerFromString(mechanism, path string) (Handler, error) {
	switch strings.ToLower(mechanism) {
	case "local":
		return NewLocal(path), nil
	default:
		return nil, fmt.Errorf("invalid mechanism: %s", mechanism)
	}
}
