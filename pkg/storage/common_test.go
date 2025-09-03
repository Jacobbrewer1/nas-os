package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple path", "folder/file.txt", "folder/file.txt"},
		{"path with spaces", "folder with spaces/file name.txt", "folder_with_spaces/file_name.txt"},
		{"complex path", "  /folder/../folder with spaces//file name.txt  ", "/folder_with_spaces/file_name.txt"},
		{"empty path", "", "."},
		{"root path", "/", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := cleanPath(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
