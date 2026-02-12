package vault

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(`{ "data": {"data": { "password": "password-val" }} }`))
	}))
	secret, err := getSecret(srv.URL, "token", "path/to/secret", "password")
	if err != nil {
		t.Fatal(err)
	}
	if secret != "password-val" {
		t.Fatal("Failed to query")
	}
	defer srv.Close()
}

func TestGetSecret_Success(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		path          string
		field         string
		expectedValue string
	}{
		{
			name:          "simple secret",
			responseBody:  `{ "data": {"data": { "password": "my-secret-password" }} }`,
			path:          "secret/data/myapp",
			field:         "password",
			expectedValue: "my-secret-password",
		},
		{
			name:          "complex secret with multiple fields",
			responseBody:  `{ "data": {"data": { "username": "admin", "password": "admin123", "api_key": "abc123" }} }`,
			path:          "secret/data/api",
			field:         "api_key",
			expectedValue: "abc123",
		},
		{
			name:          "secret with special characters",
			responseBody:  `{ "data": {"data": { "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." }} }`,
			path:          "secret/data/tokens",
			field:         "token",
			expectedValue: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.Write([]byte(tt.responseBody))
			}))
			defer srv.Close()

			secret, err := getSecret(srv.URL, "test-token", tt.path, tt.field)
			if err != nil {
				t.Fatalf("getSecret() error = %v, want nil", err)
			}

			if secret != tt.expectedValue {
				t.Errorf("getSecret() = %v, want %v", secret, tt.expectedValue)
			}
		})
	}
}

func TestGetSecret_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		path          string
		field         string
		expectError   bool
		errorContains string
	}{
		{
			name:          "HTTP 404 Not Found",
			statusCode:    http.StatusNotFound,
			responseBody:  `{ "errors": ["secret not found"] }`,
			path:          "secret/data/nonexistent",
			field:         "password",
			expectError:   true,
			errorContains: "invalid HTTP status: 404",
		},
		{
			name:          "HTTP 403 Forbidden",
			statusCode:    http.StatusForbidden,
			responseBody:  `{ "errors": ["permission denied"] }`,
			path:          "secret/data/restricted",
			field:         "password",
			expectError:   true,
			errorContains: "invalid HTTP status: 403",
		},
		{
			name:          "HTTP 500 Internal Server Error",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{ "errors": ["internal server error"] }`,
			path:          "secret/data/error",
			field:         "password",
			expectError:   true,
			errorContains: "invalid HTTP status: 500",
		},
		{
			name:          "missing field in response",
			statusCode:    http.StatusOK,
			responseBody:  `{ "data": {"data": { "username": "admin" }} }`,
			path:          "secret/data/myapp",
			field:         "password",
			expectError:   true,
			errorContains: "no field 'password' in secret",
		},
		{
			name:          "malformed JSON response",
			statusCode:    http.StatusOK,
			responseBody:  `{ "data": {"data": { "password": "unclosed-brace" }`,
			path:          "secret/data/myapp",
			field:         "password",
			expectError:   true,
			errorContains: "unexpected end of JSON input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(tt.statusCode)
				writer.Write([]byte(tt.responseBody))
			}))
			defer srv.Close()

			secret, err := getSecret(srv.URL, "test-token", tt.path, tt.field)

			if tt.expectError {
				if err == nil {
					t.Errorf("getSecret() error = nil, want non-nil")
				} else if tt.errorContains != "" && !containsError(err, tt.errorContains) {
					t.Errorf("getSecret() error = %v, want error containing %q", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("getSecret() error = %v, want nil", err)
				}
			}

			if !tt.expectError && secret != "" {
				t.Errorf("getSecret() = %v, want empty string", secret)
			}
		})
	}
}

func TestGetSecret_RequestValidation(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		field         string
		expectPanic   bool
		expectError   bool
		errorContains string
		useMockServer bool
	}{
		{
			name:        "empty path causes panic",
			path:        "",
			field:       "password",
			expectPanic: true,
		},
		{
			name:        "path without slash causes panic",
			path:        "single",
			field:       "password",
			expectPanic: true,
		},
		{
			name:          "empty field",
			path:          "secret/data/myapp",
			field:         "",
			expectError:   true,
			errorContains: "no field '' in secret for path 'secret/data/myapp'",
			useMockServer: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				// Test that this causes a panic
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("getSecret() did not panic as expected")
					}
				}()
			}

			var secret string
			var err error

			if tt.useMockServer {
				// Set up a mock server for this test
				srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					writer.Write([]byte(`{ "data": {"data": { "password": "test-value" }} }`))
				}))
				defer srv.Close()

				secret, err = getSecret(srv.URL, "test-token", tt.path, tt.field)
			} else {
				// This will fail before making HTTP request due to path parsing
				secret, err = getSecret("http://example.com", "test-token", tt.path, tt.field)
			}

			if tt.expectPanic {
				t.Errorf("getSecret() should have panicked but completed")
			} else if tt.expectError {
				if err == nil {
					t.Errorf("getSecret() error = nil, want non-nil")
				} else if tt.errorContains != "" && !containsError(err, tt.errorContains) {
					t.Errorf("getSecret() error = %v, want error containing %q", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("getSecret() error = %v, want nil", err)
				}
			}

			if !tt.expectPanic && !tt.expectError && secret != "" {
				t.Errorf("getSecret() = %v, want empty string", secret)
			}
		})
	}
}

// Helper function to check if error message contains expected substring
func containsError(err error, substring string) bool {
	return err != nil && containsString(err.Error(), substring)
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || containsString(s[1:], substr)))
}
