package storage_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/pointlet/plinth/internal/core/storage"
)

/* ########################################################
   Test cases for storage.WriteRawFile function
######################################################## */

func TestWriteRawFile(t *testing.T) {
	tests := []struct {
		name string
		path string
		data []byte
	}{
		{
			name: "simple file",
			path: "testfile.txt",
			data: []byte("hello world"),
		},
		{
			name: "nested directories",
			path: "a/b/c/d/file.txt",
			data: []byte("nested content"),
		},
		{
			name: "special chars in directory",
			path: "dir with spaces/another-dir_123/file.txt",
			data: []byte("special dir content"),
		},
		{
			name: "special chars in filename",
			path: "docs/file with spaces & symbols (1)_v2.0.txt",
			data: []byte("special filename content"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()

			storage.WriteRawFile(tmpDir, tt.path, tt.data)

			read, err := os.ReadFile(filepath.Join(tmpDir, tt.path))
			if err != nil {
				t.Fatalf("failed to read written file: %v", err)
			}
			if string(read) != string(tt.data) {
				t.Errorf("expected %q, got %q", tt.data, read)
			}
		})
	}

	t.Run("concurrent writes to different files", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		numWriters := 10

		var wg sync.WaitGroup
		wg.Add(numWriters)

		for i := range numWriters {
			go func(n int) {
				defer wg.Done()
				path := fmt.Sprintf("concurrent/file_%d.txt", n)
				data := fmt.Appendf(nil, "content from writer %d", n)
				storage.WriteRawFile(tmpDir, path, data)
			}(i)
		}

		wg.Wait()

		for i := range numWriters {
			path := fmt.Sprintf("concurrent/file_%d.txt", i)
			expected := fmt.Appendf(nil, "content from writer %d", i)

			read, err := os.ReadFile(filepath.Join(tmpDir, path))
			if err != nil {
				t.Errorf("failed to read file %d: %v", i, err)
				continue
			}
			if string(read) != string(expected) {
				t.Errorf("file %d: expected %q, got %q", i, expected, read)
			}
		}
	})
}

/* ########################################################
	Test cases for storage.BatchWrite function
######################################################## */

/* ########################################################
   Test cases for storage.ListDirectoryContents function
######################################################## */

func TestListDirectoryContents(t *testing.T) {
	t.Run("file path returns error", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		storage.WriteRawFile(tmpDir, "file.txt", []byte("content"))

		_, err := storage.ListDirectoryContents(tmpDir, "file.txt")
		if err == nil {
			t.Error("expected error when listing a file, got nil")
		}
	})

	t.Run("empty directory returns empty slice", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		os.MkdirAll(filepath.Join(tmpDir, "empty"), 0o755)

		entries, err := storage.ListDirectoryContents(tmpDir, "empty")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(entries))
		}
	})

	t.Run("directory with only subdirectories", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		dirs := []string{"alpha", "beta", "gamma"}
		for _, dir := range dirs {
			os.MkdirAll(filepath.Join(tmpDir, "parent", dir), 0o755)
		}

		entries, err := storage.ListDirectoryContents(tmpDir, "parent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != len(dirs) {
			t.Fatalf("expected %d entries, got %d", len(dirs), len(entries))
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				t.Errorf("expected %q to be a directory", entry.Name())
			}
		}
	})

	t.Run("directory with only files", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		files := []struct {
			name string
			data []byte
		}{
			{"file1.txt", []byte("content one")},
			{"file2.txt", []byte("content two")},
			{"file3.txt", []byte("content three")},
		}

		for _, f := range files {
			storage.WriteRawFile(tmpDir, filepath.Join("docs", f.name), f.data)
		}

		entries, err := storage.ListDirectoryContents(tmpDir, "docs")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != len(files) {
			t.Fatalf("expected %d entries, got %d", len(files), len(entries))
		}

		for _, entry := range entries {
			if entry.IsDir() {
				t.Errorf("expected %q to be a file", entry.Name())
			}
			info, err := entry.Info()
			if err != nil {
				t.Errorf("failed to get info for %q: %v", entry.Name(), err)
				continue
			}
			if info.Size() == 0 {
				t.Errorf("expected %q to have non-zero size", entry.Name())
			}
		}
	})

	t.Run("non-existent directory returns error", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		_, err := storage.ListDirectoryContents(tmpDir, "does-not-exist")
		if err == nil {
			t.Error("expected error for non-existent directory, got nil")
		}
	})

	t.Run("mixed content distinguishes files and directories", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		os.MkdirAll(filepath.Join(tmpDir, "mixed", "subdir1"), 0o755)
		os.MkdirAll(filepath.Join(tmpDir, "mixed", "subdir2"), 0o755)
		storage.WriteRawFile(tmpDir, "mixed/file1.txt", []byte("one"))
		storage.WriteRawFile(tmpDir, "mixed/file2.txt", []byte("two"))

		entries, err := storage.ListDirectoryContents(tmpDir, "mixed")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 4 {
			t.Fatalf("expected 4 entries, got %d", len(entries))
		}

		dirCount := 0
		fileCount := 0
		for _, entry := range entries {
			if entry.IsDir() {
				dirCount++
			} else {
				fileCount++
			}
		}

		if dirCount != 2 {
			t.Errorf("expected 2 directories, got %d", dirCount)
		}
		if fileCount != 2 {
			t.Errorf("expected 2 files, got %d", fileCount)
		}
	})
}
