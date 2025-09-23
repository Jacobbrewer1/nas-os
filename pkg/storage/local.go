package storage

import (
	"fmt"
	"os"
	"strings"
)

// IsLocalStorage checks if the storage mechanism is local.
func IsLocalStorage(mechanism string) bool {
	return strings.EqualFold(mechanism, "local")
}

// EnsureLocalPath ensures that the given path is suitable for local storage or creates it if it doesn't exist.
func EnsureLocalPath(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	} else if err != nil {
		return fmt.Errorf("failed to stat path %s: %w", path, err)
	}
	return nil
}
