package registry

import (
	"fmt"
	"net/http"
	"testing"
)

func TestZotRegistry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping zot integration test")
	}

	t.Run("starts and responds to health check", func(t *testing.T) {
		reg := NewZotRegistry()
		if err := reg.Start(t.Context()); err != nil {
			t.Fatalf("failed to start zot: %v", err)
		}
		t.Cleanup(func() { reg.Stop(t.Context()) })

		addr := reg.Address()
		if addr == "" {
			t.Fatal("expected non-empty address")
		}

		resp, err := http.Get(fmt.Sprintf("http://%s/v2/", addr))
		if err != nil {
			t.Fatalf("failed to reach zot: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("is local", func(t *testing.T) {
		reg := NewZotRegistry()
		if !reg.IsLocal() {
			t.Error("expected IsLocal() to be true")
		}
	})
}
