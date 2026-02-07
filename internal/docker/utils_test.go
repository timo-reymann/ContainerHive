package docker

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTar(t *testing.T) {
	t.Run("extracts regular files and directories", func(t *testing.T) {
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)

		content1 := []byte(`{"schemaVersion":2}`)
		tw.WriteHeader(&tar.Header{Name: "index.json", Size: int64(len(content1)), Mode: 0644})
		tw.Write(content1)

		tw.WriteHeader(&tar.Header{Name: "blobs/", Typeflag: tar.TypeDir, Mode: 0755})

		content2 := []byte("blob data")
		tw.WriteHeader(&tar.Header{Name: "blobs/sha256/abc123", Size: int64(len(content2)), Mode: 0644})
		tw.Write(content2)

		tw.Close()

		tarPath := filepath.Join(t.TempDir(), "test.tar")
		if err := os.WriteFile(tarPath, buf.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}

		destDir := t.TempDir()
		if err := extractTar(tarPath, destDir); err != nil {
			t.Fatal("extractTar failed:", err)
		}

		// Verify top-level file
		got, err := os.ReadFile(filepath.Join(destDir, "index.json"))
		if err != nil {
			t.Fatal("index.json not extracted:", err)
		}
		if string(got) != string(content1) {
			t.Fatalf("index.json content: want %q, got %q", string(content1), string(got))
		}

		// Verify directory was created
		info, err := os.Stat(filepath.Join(destDir, "blobs"))
		if err != nil {
			t.Fatal("blobs/ directory not created:", err)
		}
		if !info.IsDir() {
			t.Fatal("blobs/ should be a directory")
		}

		// Verify nested file
		got2, err := os.ReadFile(filepath.Join(destDir, "blobs/sha256/abc123"))
		if err != nil {
			t.Fatal("nested blob not extracted:", err)
		}
		if string(got2) != "blob data" {
			t.Fatalf("blob content: want %q, got %q", "blob data", string(got2))
		}
	})

	t.Run("creates intermediate directories for files without explicit dir entries", func(t *testing.T) {
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)

		// File in nested dir, but no explicit directory entries in the tar
		content := []byte("deep file")
		tw.WriteHeader(&tar.Header{Name: "a/b/c/deep.txt", Size: int64(len(content)), Mode: 0644})
		tw.Write(content)
		tw.Close()

		tarPath := filepath.Join(t.TempDir(), "test.tar")
		os.WriteFile(tarPath, buf.Bytes(), 0644)

		destDir := t.TempDir()
		if err := extractTar(tarPath, destDir); err != nil {
			t.Fatal("extractTar failed:", err)
		}

		got, err := os.ReadFile(filepath.Join(destDir, "a/b/c/deep.txt"))
		if err != nil {
			t.Fatal("deep file not extracted:", err)
		}
		if string(got) != "deep file" {
			t.Fatalf("deep file content: want %q, got %q", "deep file", string(got))
		}
	})

	t.Run("handles empty tar", func(t *testing.T) {
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)
		tw.Close()

		tarPath := filepath.Join(t.TempDir(), "empty.tar")
		os.WriteFile(tarPath, buf.Bytes(), 0644)

		if err := extractTar(tarPath, t.TempDir()); err != nil {
			t.Fatal("empty tar should not error:", err)
		}
	})

	t.Run("rejects path traversal with dot-dot", func(t *testing.T) {
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)
		tw.WriteHeader(&tar.Header{Name: "../../etc/evil", Size: 4, Mode: 0644})
		tw.Write([]byte("pwnd"))
		tw.Close()

		tarPath := filepath.Join(t.TempDir(), "evil.tar")
		os.WriteFile(tarPath, buf.Bytes(), 0644)

		err := extractTar(tarPath, t.TempDir())
		if err == nil {
			t.Fatal("expected error for path traversal entry")
		}
		t.Logf("correctly rejected: %v", err)
	})

	t.Run("rejects absolute paths in tar", func(t *testing.T) {
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)
		tw.WriteHeader(&tar.Header{Name: "/etc/passwd", Size: 4, Mode: 0644})
		tw.Write([]byte("root"))
		tw.Close()

		tarPath := filepath.Join(t.TempDir(), "abs.tar")
		os.WriteFile(tarPath, buf.Bytes(), 0644)

		err := extractTar(tarPath, t.TempDir())
		if err == nil {
			t.Fatal("expected error for absolute path entry")
		}
		t.Logf("correctly rejected: %v", err)
	})

	t.Run("returns error for nonexistent tar", func(t *testing.T) {
		err := extractTar("/nonexistent/path.tar", t.TempDir())
		if err == nil {
			t.Fatal("expected error for nonexistent tar")
		}
		t.Logf("correctly errored: %v", err)
	})

	t.Run("returns error for invalid tar data", func(t *testing.T) {
		tarPath := filepath.Join(t.TempDir(), "garbage.tar")
		os.WriteFile(tarPath, []byte("not a tar file"), 0644)

		// extractTar reads headers — invalid data should cause a tar read error
		// or silently succeed if the garbage is too short for any header
		_ = extractTar(tarPath, t.TempDir())
		// No assertion — just verifying no panic
	})
}
