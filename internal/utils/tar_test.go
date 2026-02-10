package utils

import (
	"archive/tar"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTar(t *testing.T) {
	t.Run("successful extraction", func(t *testing.T) {
		// Create a temporary directory for the test
		destDir := t.TempDir()

		// Create a test tar file
		tarPath := filepath.Join(destDir, "test.tar")
		if err := createTestTar(tarPath); err != nil {
			t.Fatalf("failed to create test tar: %v", err)
		}

		// Extract the tar
		if err := ExtractTar(tarPath, destDir); err != nil {
			t.Fatalf("ExtractTar failed: %v", err)
		}

		// Verify extracted files
		expectedFiles := []string{
			"testfile.txt",
			"subdir/testfile2.txt",
		}

		for _, file := range expectedFiles {
			fullPath := filepath.Join(destDir, file)
			if _, err := os.Stat(fullPath); err != nil {
				t.Errorf("expected file %s not found: %v", file, err)
			}
		}
	})

	t.Run("non-existent tar file", func(t *testing.T) {
		destDir := t.TempDir()
		err := ExtractTar("/non/existent/file.tar", destDir)
		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})

	t.Run("path traversal attack prevention", func(t *testing.T) {
		// Create a temporary directory for the test
		destDir := t.TempDir()

		// Create a malicious tar file with path traversal
		tarPath := filepath.Join(destDir, "malicious.tar")
		if err := createMaliciousTar(tarPath); err != nil {
			t.Fatalf("failed to create malicious tar: %v", err)
		}

		// Try to extract - should fail
		err := ExtractTar(tarPath, destDir)
		if err == nil {
			t.Error("expected error for path traversal, got nil")
		}
		if err.Error() != "tar entry escapes destination: ../../../etc/passwd" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("empty tar file", func(t *testing.T) {
		destDir := t.TempDir()

		// Create an empty tar file
		tarPath := filepath.Join(destDir, "empty.tar")
		f, err := os.Create(tarPath)
		if err != nil {
			t.Fatalf("failed to create empty tar: %v", err)
		}
		f.Close()

		// Extract should succeed with no files
		if err := ExtractTar(tarPath, destDir); err != nil {
			t.Fatalf("ExtractTar failed on empty tar: %v", err)
		}

		// Verify no files were created (except the tar itself)
		files, err := os.ReadDir(destDir)
		if err != nil {
			t.Fatalf("failed to read dest dir: %v", err)
		}
		if len(files) != 1 {
			t.Errorf("expected only the tar file, got %d files", len(files))
		}
	})
}

// createTestTar creates a test tar file with some sample content
func createTestTar(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	defer tw.Close()

	// Add a regular file
	if err := tw.WriteHeader(&tar.Header{
		Name:     "testfile.txt",
		Mode:     0644,
		Size:     int64(len("test content")),
		Typeflag: tar.TypeReg,
	}); err != nil {
		return err
	}
	if _, err := tw.Write([]byte("test content")); err != nil {
		return err
	}

	// Add a directory and file in it
	if err := tw.WriteHeader(&tar.Header{
		Name:     "subdir/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
	}); err != nil {
		return err
	}

	if err := tw.WriteHeader(&tar.Header{
		Name:     "subdir/testfile2.txt",
		Mode:     0644,
		Size:     int64(len("test content 2")),
		Typeflag: tar.TypeReg,
	}); err != nil {
		return err
	}
	if _, err := tw.Write([]byte("test content 2")); err != nil {
		return err
	}

	return nil
}

// createMaliciousTar creates a tar file with path traversal attempt
func createMaliciousTar(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	defer tw.Close()

	// Add a file that tries to escape the destination directory
	if err := tw.WriteHeader(&tar.Header{
		Name:     "../../../etc/passwd",
		Mode:     0644,
		Size:     int64(len("malicious content")),
		Typeflag: tar.TypeReg,
	}); err != nil {
		return err
	}
	if _, err := tw.Write([]byte("malicious content")); err != nil {
		return err
	}

	return nil
}

// Test helper function to verify file contents
func verifyFileContent(t *testing.T, path, expectedContent string) {
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}
	if string(content) != expectedContent {
		t.Errorf("file %s: expected content %q, got %q", path, expectedContent, string(content))
	}
}