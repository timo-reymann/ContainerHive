package vault

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetSecretWithDefaultConfiguration(t *testing.T) {
	// Set up test environment
	t.Setenv("VAULT_ADDR", "http://localhost:8200")
	defer os.Unsetenv("VAULT_ADDR")

	// Create a test HTTP server to mock Vault API
	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Verify the request has the correct token
		token := request.Header.Get("X-Vault-Token")
		if token != "test-cli-token" {
			writer.WriteHeader(http.StatusUnauthorized)
			writer.Write([]byte(`{"errors": ["invalid token"]}`))
			return
		}

		// Return a successful response
		writer.Write([]byte(`{ "data": {"data": { "password": "secret-password-123" }} }`))
	}))
	defer srv.Close()

	// Set up CLI token
	tempDir := t.TempDir()
	originalHomeDir := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHomeDir)

	cliToken := "test-cli-token"
	vaultTokenFile := filepath.Join(tempDir, ".vault-token")
	err := os.WriteFile(vaultTokenFile, []byte(cliToken), 0600)
	if err != nil {
		t.Fatalf("Failed to create test vault token file: %v", err)
	}
	os.Setenv("HOME", tempDir)

	// Override VAULT_ADDR to point to our test server
	t.Setenv("VAULT_ADDR", srv.URL)

	// Test the function
	secret, err := GetSecretWithDefaultConfiguration("secret/data/myapp", "password")
	if err != nil {
		t.Fatalf("GetSecretWithDefaultConfiguration() error = %v, want nil", err)
	}

	if secret != "secret-password-123" {
		t.Errorf("GetSecretWithDefaultConfiguration() = %v, want %v", secret, "secret-password-123")
	}
}

func TestGetSecretWithDefaultConfiguration_EnvToken(t *testing.T) {
	// Set up test environment
	t.Setenv("VAULT_ADDR", "http://localhost:8200")
	t.Setenv("VAULT_TOKEN", "test-env-token")
	defer func() {
		os.Unsetenv("VAULT_ADDR")
		os.Unsetenv("VAULT_TOKEN")
	}()

	// Create a test HTTP server to mock Vault API
	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Verify the request has the correct token
		token := request.Header.Get("X-Vault-Token")
		if token != "test-env-token" {
			writer.WriteHeader(http.StatusUnauthorized)
			writer.Write([]byte(`{"errors": ["invalid token"]}`))
			return
		}

		// Return a successful response
		writer.Write([]byte(`{ "data": {"data": { "api_key": "test-api-key-456" }} }`))
	}))
	defer srv.Close()

	// Override VAULT_ADDR to point to our test server
	t.Setenv("VAULT_ADDR", srv.URL)

	// Test the function
	secret, err := GetSecretWithDefaultConfiguration("secret/data/api", "api_key")
	if err != nil {
		t.Fatalf("GetSecretWithDefaultConfiguration() error = %v, want nil", err)
	}

	if secret != "test-api-key-456" {
		t.Errorf("GetSecretWithDefaultConfiguration() = %v, want %v", secret, "test-api-key-456")
	}
}

func TestGetSecretWithDefaultConfiguration_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		setupEnv      func(*testing.T)
		cleanupEnv    func()
		expectedError bool
		errorContains string
	}{
		{
			name: "VAULT_ADDR not set",
			setupEnv: func(t *testing.T) {
				os.Unsetenv("VAULT_ADDR")
				os.Unsetenv("VAULT_TOKEN")
				// Create a temporary home directory with a vault token file
				tempDir := t.TempDir()
				os.Setenv("HOME", tempDir)
				// Create a dummy vault token file so LookupToken() succeeds
				tokenFile := filepath.Join(tempDir, ".vault-token")
				err := os.WriteFile(tokenFile, []byte("dummy-token"), 0600)
				if err != nil {
					t.Fatalf("Failed to create test vault token file: %v", err)
				}
			},
			cleanupEnv: func() {
				os.Unsetenv("HOME")
			},
			expectedError: true,
			errorContains: "environment variable VAULT_ADDR not set",
		},
		{
			name: "no token available",
			setupEnv: func(t *testing.T) {
				os.Setenv("VAULT_ADDR", "http://localhost:8200")
				os.Unsetenv("VAULT_TOKEN")
				os.Setenv("HOME", "/nonexistent")
			},
			cleanupEnv: func() {
				os.Unsetenv("VAULT_ADDR")
				os.Unsetenv("HOME")
			},
			expectedError: true,
			errorContains: "open /nonexistent/.vault-token: no such file or directory",
		},
		{
			name: "Vault API returns 404",
			setupEnv: func(t *testing.T) {
				os.Setenv("VAULT_ADDR", "http://localhost:8200")
				os.Setenv("VAULT_TOKEN", "test-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("VAULT_ADDR")
				os.Unsetenv("VAULT_TOKEN")
			},
			expectedError: true,
			errorContains: "invalid HTTP status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment
			if tt.setupEnv != nil {
				defer tt.cleanupEnv()
				tt.setupEnv(t) // Pass t to setup function
			}

			// For API error cases, set up a test server
			if tt.name == "Vault API returns 404" {
				srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusNotFound)
					writer.Write([]byte(`{"errors": ["secret not found"]}`))
				}))
				defer srv.Close()
				os.Setenv("VAULT_ADDR", srv.URL)
			}

			// Test the function
			secret, err := GetSecretWithDefaultConfiguration("secret/data/test", "password")

			if tt.expectedError {
				if err == nil {
					t.Errorf("GetSecretWithDefaultConfiguration() error = nil, want non-nil")
				} else if tt.errorContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errorContains)) {
					t.Errorf("GetSecretWithDefaultConfiguration() error = %v, want error containing %q", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("GetSecretWithDefaultConfiguration() error = %v, want nil", err)
				}
			}

			if !tt.expectedError && secret != "" {
				t.Errorf("GetSecretWithDefaultConfiguration() = %v, want empty string", secret)
			}
		})
	}
}
