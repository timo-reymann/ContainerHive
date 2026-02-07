package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// buildOCITar creates a minimal valid OCI image tar for testing.
// It constructs an OCI layout with index.json, oci-layout, a manifest blob,
// and a config blob. The image has no layers.
func buildOCITar(t *testing.T, imageName string) string {
	t.Helper()

	config := []byte(`{"architecture":"amd64","os":"linux","rootfs":{"type":"layers","diff_ids":[]}}`)
	configDigest := fmt.Sprintf("sha256:%x", sha256.Sum256(config))

	manifest, _ := json.Marshal(map[string]any{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.manifest.v1+json",
		"config": map[string]any{
			"mediaType": "application/vnd.oci.image.config.v1+json",
			"digest":    configDigest,
			"size":      len(config),
		},
		"layers": []any{},
	})
	manifestDigest := fmt.Sprintf("sha256:%x", sha256.Sum256(manifest))

	annotations := map[string]string{}
	if imageName != "" {
		annotations["io.containerd.image.name"] = imageName
	}

	index, _ := json.Marshal(map[string]any{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.index.v1+json",
		"manifests": []map[string]any{
			{
				"mediaType":   "application/vnd.oci.image.manifest.v1+json",
				"digest":      manifestDigest,
				"size":        len(manifest),
				"annotations": annotations,
			},
		},
	})

	ociLayout := []byte(`{"imageLayoutVersion":"1.0.0"}`)

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	writeEntry := func(name string, data []byte) {
		tw.WriteHeader(&tar.Header{Name: name, Size: int64(len(data)), Mode: 0644})
		tw.Write(data)
	}

	writeEntry("oci-layout", ociLayout)
	writeEntry("index.json", index)

	// Manifest blob: blobs/sha256/<hash>
	writeEntry(fmt.Sprintf("blobs/sha256/%x", sha256.Sum256(manifest)), manifest)

	// Config blob
	writeEntry(fmt.Sprintf("blobs/sha256/%x", sha256.Sum256(config)), config)

	tw.Close()

	p := filepath.Join(t.TempDir(), "image.tar")
	if err := os.WriteFile(p, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

// buildOCITarNoManifests creates an OCI tar with an empty manifests array.
func buildOCITarNoManifests(t *testing.T) string {
	t.Helper()

	index, _ := json.Marshal(map[string]any{
		"schemaVersion": 2,
		"manifests":     []any{},
	})
	ociLayout := []byte(`{"imageLayoutVersion":"1.0.0"}`)

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	writeEntry := func(name string, data []byte) {
		tw.WriteHeader(&tar.Header{Name: name, Size: int64(len(data)), Mode: 0644})
		tw.Write(data)
	}

	writeEntry("oci-layout", ociLayout)
	writeEntry("index.json", index)
	tw.Close()

	p := filepath.Join(t.TempDir(), "no-manifests.tar")
	if err := os.WriteFile(p, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadImageFromTar(t *testing.T) {
	// These tests exercise every error path in LoadImageFromTar that does not
	// require a running Docker daemon.

	client := &Client{} // nil docker client — fine for tests that fail before daemon.Write

	t.Run("returns error for nonexistent tar path", func(t *testing.T) {
		_, err := client.LoadImageFromTar(context.Background(), "/nonexistent/image.tar")
		if err == nil {
			t.Fatal("expected error for nonexistent tar")
		}
		t.Logf("got expected error: %v", err)
	})

	t.Run("returns error for invalid tar data", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "garbage.tar")
		os.WriteFile(p, []byte("not a tar file at all"), 0644)

		_, err := client.LoadImageFromTar(context.Background(), p)
		if err == nil {
			t.Fatal("expected error for invalid tar data")
		}
		t.Logf("got expected error: %v", err)
	})

	t.Run("returns error for tar missing oci-layout file", func(t *testing.T) {
		// Create a tar with only an index.json but no oci-layout
		index, _ := json.Marshal(map[string]any{
			"schemaVersion": 2,
			"manifests":     []any{},
		})

		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)
		tw.WriteHeader(&tar.Header{Name: "index.json", Size: int64(len(index)), Mode: 0644})
		tw.Write(index)
		tw.Close()

		p := filepath.Join(t.TempDir(), "no-layout.tar")
		os.WriteFile(p, buf.Bytes(), 0644)

		_, err := client.LoadImageFromTar(context.Background(), p)
		if err == nil {
			t.Fatal("expected error for tar without oci-layout")
		}
		t.Logf("got expected error: %v", err)
	})

	t.Run("returns error when no manifests in index", func(t *testing.T) {
		tarPath := buildOCITarNoManifests(t)

		_, err := client.LoadImageFromTar(context.Background(), tarPath)
		if err == nil {
			t.Fatal("expected error for empty manifests")
		}
		if err.Error() != "no manifests in OCI layout" {
			t.Fatalf("unexpected error message: %v", err)
		}
		t.Logf("got expected error: %v", err)
	})

	t.Run("returns error when image name annotation is missing", func(t *testing.T) {
		tarPath := buildOCITar(t, "") // empty image name → no annotation

		_, err := client.LoadImageFromTar(context.Background(), tarPath)
		if err == nil {
			t.Fatal("expected error for missing image name annotation")
		}
		t.Logf("got expected error: %v", err)
	})

	t.Run("returns error when image name is invalid", func(t *testing.T) {
		// Use an invalid Docker tag reference
		tarPath := buildOCITar(t, "INVALID:!!!")

		_, err := client.LoadImageFromTar(context.Background(), tarPath)
		if err == nil {
			t.Fatal("expected error for invalid image name")
		}
		t.Logf("got expected error: %v", err)
	})

	t.Run("returns error when daemon is not available", func(t *testing.T) {
		tarPath := buildOCITar(t, "test-image:latest")

		// Use a client pointing at a non-existent Docker daemon
		t.Setenv("DOCKER_HOST", "tcp://127.0.0.1:1")
		dockerClient, err := NewClient()
		if err != nil {
			t.Fatal("failed to create docker client:", err)
		}
		defer dockerClient.Close()

		_, err = dockerClient.LoadImageFromTar(context.Background(), tarPath)
		if err == nil {
			t.Fatal("expected error when Docker daemon is unreachable")
		}
		t.Logf("got expected error: %v", err)
	})
}

func TestNewClient(t *testing.T) {
	t.Run("creates client from environment", func(t *testing.T) {
		client, err := NewClient()
		if err != nil {
			t.Fatal("NewClient failed:", err)
		}
		defer client.Close()

		if client.docker == nil {
			t.Fatal("expected docker client to be non-nil")
		}
		t.Log("client created successfully")
	})

	t.Run("respects DOCKER_HOST from environment", func(t *testing.T) {
		t.Setenv("DOCKER_HOST", "tcp://127.0.0.1:9999")

		client, err := NewClient()
		if err != nil {
			t.Fatal("NewClient failed:", err)
		}
		defer client.Close()

		// Client is created but connection fails lazily on first use
		t.Log("client created with custom DOCKER_HOST")
	})
}

func TestClientClose(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	if err := client.Close(); err != nil {
		t.Fatal("Close failed:", err)
	}
	t.Log("client closed successfully")
}
