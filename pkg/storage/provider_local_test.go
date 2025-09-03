package storage

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocal_ReadFile(t *testing.T) {
	t.Parallel()

	t.Run("golden", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Write a test file
		testFilePath := "testfile.txt"
		testFileContent := []byte("Hello, World!")
		err := os.WriteFile(filepath.Join(testDir, testFilePath), testFileContent, 0o600)
		require.NoError(t, err)

		got, err := localBackend.ReadFile(testFilePath)
		require.NoError(t, err)

		t.Cleanup(func() {
			err := got.Close()
			require.NoError(t, err)
		})

		data, err := io.ReadAll(got)
		require.NoError(t, err)
		require.Equal(t, testFileContent, data)
	})

	t.Run("replaces spaces", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)
		// Write a test file with spaces in the name
		testFilePath := "test_file_with_spaces.txt"
		testFileContent := []byte("Hello, World with spaces!")
		err := os.WriteFile(filepath.Join(testDir, testFilePath), testFileContent, 0o600)
		require.NoError(t, err)

		got, err := localBackend.ReadFile("test file with spaces.txt")
		require.NoError(t, err)
		t.Cleanup(func() {
			err := got.Close()
			require.NoError(t, err)
		})

		data, err := io.ReadAll(got)
		require.NoError(t, err)
		require.Equal(t, testFileContent, data)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		_, err := localBackend.ReadFile("nonexistent.txt")
		require.Error(t, err)
		require.Contains(t, err.Error(), "file not found")
	})
}

func TestLocal_WriteFile(t *testing.T) {
	t.Parallel()

	t.Run("golden", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		testFilePath := "write_testfile.txt"
		testFileContent := []byte("Writing to local storage!")

		err := localBackend.WriteFile(testFilePath, io.NopCloser(bytes.NewReader(testFileContent)))
		require.NoError(t, err)

		// Verify the file was written
		data, err := os.ReadFile(filepath.Join(testDir, testFilePath))
		require.NoError(t, err)
		require.Equal(t, testFileContent, data)
	})

	t.Run("creates directories", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		testFilePath := "nested/dir/write_testfile.txt"
		testFileContent := []byte("Writing to nested directory!")

		err := localBackend.WriteFile(testFilePath, io.NopCloser(bytes.NewReader(testFileContent)))
		require.NoError(t, err)

		// Verify the file was written
		data, err := os.ReadFile(filepath.Join(testDir, testFilePath))
		require.NoError(t, err)
		require.Equal(t, testFileContent, data)
	})

	t.Run("creates directories with leading slash", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		testFilePath := "/nested/dir/write_testfile.txt"
		testFileContent := []byte("Writing to nested directory!")

		err := localBackend.WriteFile(testFilePath, io.NopCloser(bytes.NewReader(testFileContent)))
		require.NoError(t, err)

		// Verify the file was written
		data, err := os.ReadFile(filepath.Join(testDir, testFilePath))
		require.NoError(t, err)
		require.Equal(t, testFileContent, data)
	})

	t.Run("replaces spaces", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		testFilePath := "file with spaces.txt"
		testFileContent := []byte("Writing to file with spaces!")

		err := localBackend.WriteFile(testFilePath, io.NopCloser(bytes.NewReader(testFileContent)))
		require.NoError(t, err)

		// Verify the file was written
		data, err := os.ReadFile(filepath.Join(testDir, "file_with_spaces.txt"))
		require.NoError(t, err)
		require.Equal(t, testFileContent, data)
	})
}

func TestLocal_DeleteFile(t *testing.T) {
	t.Parallel()

	t.Run("golden", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Write a test file to delete
		testFilePath := "delete_testfile.txt"
		testFileContent := []byte("This file will be deleted.")
		err := os.WriteFile(filepath.Join(testDir, testFilePath), testFileContent, 0o600)
		require.NoError(t, err)

		err = localBackend.DeleteFile(testFilePath)
		require.NoError(t, err)

		// Verify the file was deleted
		_, err = os.ReadFile(filepath.Join(testDir, testFilePath))
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("replaces spaces", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Write a test file with spaces in the name to delete
		testFilePath := "file with spaces.txt"
		testFileContent := []byte("This file with spaces will be deleted.")
		err := os.WriteFile(filepath.Join(testDir, "file_with_spaces.txt"), testFileContent, 0o600)
		require.NoError(t, err)

		err = localBackend.DeleteFile(testFilePath)
		require.NoError(t, err)

		// Verify the file was deleted
		_, err = os.ReadFile(filepath.Join(testDir, "file_with_spaces.txt"))
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		err := localBackend.DeleteFile("nonexistent.txt")
		require.Error(t, err)
		require.Contains(t, err.Error(), "file not found")
	})
}

func TestLocal_ListDirectory(t *testing.T) {
	t.Parallel()

	t.Run("golden", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Create test files and directories
		err := os.MkdirAll(filepath.Join(testDir, "subdir"), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("File 1"), 0o600)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(testDir, "subdir", "file2.txt"), []byte("File 2"), 0o600)
		require.NoError(t, err)

		files, err := localBackend.ListDirectory(".")
		require.NoError(t, err)
		require.Len(t, files, 2)

		var names []string
		for _, fi := range files {
			names = append(names, fi.Name)
		}
		require.Contains(t, names, "file1.txt")
		require.Contains(t, names, "subdir")
	})

	t.Run("nested directory", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Create test files and directories
		err := os.MkdirAll(filepath.Join(testDir, "nested", "subdir"), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(testDir, "nested", "file1.txt"), []byte("File 1"), 0o600)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(testDir, "nested", "subdir", "file2.txt"), []byte("File 2"), 0o600)
		require.NoError(t, err)

		files, err := localBackend.ListDirectory("nested")
		require.NoError(t, err)
		require.Len(t, files, 2)

		var names []string
		for _, fi := range files {
			names = append(names, fi.Name)
		}
		require.Contains(t, names, "file1.txt")
		require.Contains(t, names, "subdir")
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		_, err := localBackend.ListDirectory("nonexistent")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to list directory")
	})

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Create an empty directory
		err := os.MkdirAll(filepath.Join(testDir, "emptydir"), 0o755)
		require.NoError(t, err)

		files, err := localBackend.ListDirectory("emptydir")
		require.NoError(t, err)
		require.Len(t, files, 0)
	})

	t.Run("replaces spaces", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Create test files and directories with spaces
		err := os.MkdirAll(filepath.Join(testDir, "/test_dir"), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(testDir, "/test_dir/file_with_spaces.txt"), []byte("File with spaces"), 0o600)
		require.NoError(t, err)

		files, err := localBackend.ListDirectory("test dir") // Intentional space in path
		require.NoError(t, err)
		require.Len(t, files, 1)

		var names []string
		for _, fi := range files {
			names = append(names, fi.Name)
		}
		require.Contains(t, names, "file_with_spaces.txt")
	})

	t.Run("leading slash", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Create test files and directories
		err := os.MkdirAll(filepath.Join(testDir, "subdir"), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("File 1"), 0o600)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(testDir, "subdir", "file2.txt"), []byte("File 2"), 0o600)
		require.NoError(t, err)

		files, err := localBackend.ListDirectory("/.") // Leading slash
		require.NoError(t, err)
		require.Len(t, files, 2)

		var names []string
		for _, fi := range files {
			names = append(names, fi.Name)
		}
		require.Contains(t, names, "file1.txt")
		require.Contains(t, names, "subdir")
	})
}

func TestLocal_CreateDirectory(t *testing.T) {
	t.Parallel()

	t.Run("golden", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		dirPath := "newdir/subdir"
		err := localBackend.CreateDirectory(dirPath)
		require.NoError(t, err)

		// Verify the directory was created
		info, err := os.Stat(filepath.Join(testDir, dirPath))
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})

	t.Run("leading slash", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		dirPath := "/newdir/subdir"
		err := localBackend.CreateDirectory(dirPath)
		require.NoError(t, err)

		// Verify the directory was created
		info, err := os.Stat(filepath.Join(testDir, "newdir", "subdir"))
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})

	t.Run("replaces spaces", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		dirPath := "dir with spaces/subdir"
		err := localBackend.CreateDirectory(dirPath)
		require.NoError(t, err)

		// Verify the directory was created
		info, err := os.Stat(filepath.Join(testDir, "dir_with_spaces", "subdir"))
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})

	t.Run("already exists", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		dirPath := "existingdir"
		err := localBackend.CreateDirectory(dirPath)
		require.NoError(t, err)

		// Create again should not error
		err = localBackend.CreateDirectory(dirPath)
		require.NoError(t, err)

		// Verify the directory still exists
		info, err := os.Stat(filepath.Join(testDir, dirPath))
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})
}

func TestLocal_DeleteDirectory(t *testing.T) {
	t.Parallel()

	t.Run("golden", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Create a directory to delete
		dirPath := "dirtodelete/subdir"
		err := os.MkdirAll(filepath.Join(testDir, dirPath), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(testDir, dirPath, "file.txt"), []byte("To be deleted"), 0o600)
		require.NoError(t, err)

		err = localBackend.DeleteDirectory("dirtodelete")
		require.NoError(t, err)

		// Verify the directory was deleted
		_, err = os.Stat(filepath.Join(testDir, "dirtodelete"))
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("leading slash", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Create a directory to delete
		dirPath := "dirtodelete/subdir"
		err := os.MkdirAll(filepath.Join(testDir, dirPath), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(testDir, dirPath, "file.txt"), []byte("To be deleted"), 0o600)
		require.NoError(t, err)

		err = localBackend.DeleteDirectory("/dirtodelete") // Leading slash
		require.NoError(t, err)

		// Verify the directory was deleted
		_, err = os.Stat(filepath.Join(testDir, "dirtodelete"))
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("replaces spaces", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Create a directory with spaces to delete
		err := os.MkdirAll(filepath.Join(testDir, "dir_with_spaces", "subdir"), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(testDir, "dir_with_spaces", "subdir", "file.txt"), []byte("To be deleted"), 0o600)
		require.NoError(t, err)

		err = localBackend.DeleteDirectory("dir with spaces")
		require.NoError(t, err)
		// Verify the directory was deleted
		_, err = os.Stat(filepath.Join(testDir, "dir_with_spaces"))
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		err := localBackend.DeleteDirectory("nonexistentdir")
		require.NoError(t, err) // os.RemoveAll does not error if the path does not exist
	})
}

func TestLocal_GetFileInfo(t *testing.T) {
	t.Parallel()

	t.Run("file golden", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Create a test file
		testFilePath := "info_testfile.txt"
		testFileContent := []byte("File info content")
		err := os.WriteFile(filepath.Join(testDir, testFilePath), testFileContent, 0o600)
		require.NoError(t, err)

		info, err := localBackend.GetFileInfo(testFilePath)
		require.NoError(t, err)
		require.Equal(t, testFilePath, info.Name)
		require.Equal(t, testFilePath, info.Path)
		require.Equal(t, int64(len(testFileContent)), info.Size)
		require.False(t, info.IsDirectory)
		require.NotZero(t, info.ModTime)
		require.Equal(t, uint32(0o600), info.Mode)
	})

	t.Run("directory golden", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Create a test directory
		dirPath := "infodir"
		err := os.MkdirAll(filepath.Join(testDir, dirPath), 0o755)
		require.NoError(t, err)

		info, err := localBackend.GetFileInfo(dirPath)
		require.NoError(t, err)
		require.Equal(t, "infodir", info.Name)
		require.Equal(t, dirPath, info.Path)
		require.Equal(t, int64(4096), info.Size)
		require.True(t, info.IsDirectory)
		require.NotZero(t, info.ModTime)
		require.Equal(t, uint32(0o755), info.Mode)
	})

	t.Run("replaces spaces", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Create a test file with spaces
		testFilePath := "file with spaces.txt"
		testFileContent := []byte("File with spaces content")
		err := os.WriteFile(filepath.Join(testDir, "file_with_spaces.txt"), testFileContent, 0o600)
		require.NoError(t, err)
		info, err := localBackend.GetFileInfo(testFilePath)
		require.NoError(t, err)
		require.Equal(t, "file_with_spaces.txt", info.Name)
		require.Equal(t, "file_with_spaces.txt", info.Path)
		require.Equal(t, int64(len(testFileContent)), info.Size)
		require.False(t, info.IsDirectory)
		require.NotZero(t, info.ModTime)
		require.Equal(t, uint32(0o600), info.Mode)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		_, err := localBackend.GetFileInfo("nonexistent.txt")
		require.Error(t, err)
		require.Contains(t, err.Error(), "file not found")
	})
}

func TestLocal_FileExists(t *testing.T) {
	t.Parallel()

	t.Run("file exists", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Create a test file
		testFilePath := "exist_testfile.txt"
		testFileContent := []byte("File exists content")
		err := os.WriteFile(filepath.Join(testDir, testFilePath), testFileContent, 0o600)
		require.NoError(t, err)

		exists, err := localBackend.FileExists(testFilePath)
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("file does not exist", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		exists, err := localBackend.FileExists("nonexistent.txt")
		require.NoError(t, err)
		require.False(t, exists)
	})

	t.Run("replaces spaces", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Create a test file with spaces
		testFilePath := "file with spaces.txt"
		testFileContent := []byte("File with spaces content")
		err := os.WriteFile(filepath.Join(testDir, "file_with_spaces.txt"), testFileContent, 0o600)
		require.NoError(t, err)

		exists, err := localBackend.FileExists(testFilePath)
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("leading slash", func(t *testing.T) {
		t.Parallel()

		testDir := t.TempDir()
		localBackend := NewLocal(slog.New(slog.DiscardHandler), testDir)

		// Create a test file
		testFilePath := "exist_testfile.txt"
		testFileContent := []byte("File exists content")
		err := os.WriteFile(filepath.Join(testDir, testFilePath), testFileContent, 0o600)
		require.NoError(t, err)

		exists, err := localBackend.FileExists("/exist_testfile.txt") // Leading slash
		require.NoError(t, err)
		require.True(t, exists)
	})
}
