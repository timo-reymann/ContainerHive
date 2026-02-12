package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLookupCliToken_HomeDirError(t *testing.T) {
	originalHomeDir := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHomeDir)

	os.Setenv("HOME", "")
	_, err := lookupCliToken()
	if err == nil {
		t.Fatal("lookupCliToken() error = nil, want non-nil")
	}
}

func TestLookupCliToken_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalHomeDir := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHomeDir)

	os.Setenv("HOME", tempDir)
	_, err := lookupCliToken()
	if err == nil {
		t.Fatal("lookupCliToken() error = nil, want non-nil")
	}
}

func TestLookupCliToken_Success(t *testing.T) {
	tempDir := t.TempDir()
	originalHomeDir := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHomeDir)

	expectedToken := "cli-token-12345"
	vaultTokenFile := filepath.Join(tempDir, ".vault-token")
	err := os.WriteFile(vaultTokenFile, []byte(expectedToken), 0600)
	if err != nil {
		t.Fatalf("Failed to create test vault token file: %v", err)
	}

	os.Setenv("HOME", tempDir)
	token, err := lookupCliToken()
	if err != nil {
		t.Fatalf("lookupCliToken() error = %v, want nil", err)
	}
	if token != expectedToken {
		t.Fatalf("lookupCliToken() = %v, want %v", token, expectedToken)
	}
}

func TestLookupCliToken_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	originalHomeDir := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHomeDir)

	vaultTokenFile := filepath.Join(tempDir, ".vault-token")
	err := os.WriteFile(vaultTokenFile, []byte(""), 0600)
	if err != nil {
		t.Fatalf("Failed to create test vault token file: %v", err)
	}

	os.Setenv("HOME", tempDir)
	token, err := lookupCliToken()
	if err != nil {
		t.Fatalf("lookupCliToken() error = %v, want nil", err)
	}
	if token != "" {
		t.Fatalf("lookupCliToken() = %v, want empty string", token)
	}
}

func TestLookupCliToken_MultiLineFile(t *testing.T) {
	tempDir := t.TempDir()
	originalHomeDir := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHomeDir)

	expectedToken := "cli-token-12345\n"
	vaultTokenFile := filepath.Join(tempDir, ".vault-token")
	err := os.WriteFile(vaultTokenFile, []byte(expectedToken), 0600)
	if err != nil {
		t.Fatalf("Failed to create test vault token file: %v", err)
	}

	os.Setenv("HOME", tempDir)
	token, err := lookupCliToken()
	if err != nil {
		t.Fatalf("lookupCliToken() error = %v, want nil", err)
	}
	if token != expectedToken {
		t.Fatalf("lookupCliToken() = %v, want %v", token, expectedToken)
	}
}

func TestLookupEnvToken(t *testing.T) {
	expectedToken := "env-token"
	os.Setenv("VAULT_TOKEN", expectedToken)
	defer os.Unsetenv("VAULT_TOKEN")

	token := lookupEnvToken()
	if token != expectedToken {
		t.Fatalf("lookupEnvToken() = %v, want %v", token, expectedToken)
	}
}

func TestLookupEnvToken_Empty(t *testing.T) {
	os.Setenv("VAULT_TOKEN", "")
	defer os.Unsetenv("VAULT_TOKEN")

	token := lookupEnvToken()
	if token != "" {
		t.Fatalf("lookupEnvToken() = %v, want empty string", token)
	}
}

func TestLookupEnvToken_NotSet(t *testing.T) {
	os.Unsetenv("VAULT_TOKEN")

	token := lookupEnvToken()
	if token != "" {
		t.Fatalf("lookupEnvToken() = %v, want empty string", token)
	}
}

func TestLookupToken_EnvToken(t *testing.T) {
	expectedToken := "env-token"
	os.Setenv("VAULT_TOKEN", expectedToken)
	defer os.Unsetenv("VAULT_TOKEN")

	token, err := LookupToken()
	if err != nil {
		t.Fatalf("LookupToken() error = %v, want nil", err)
	}
	if token != expectedToken {
		t.Fatalf("LookupToken() = %v, want %v", token, expectedToken)
	}
}

func TestLookupToken_CliTokenFallback(t *testing.T) {
	tempDir := t.TempDir()
	originalHomeDir := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHomeDir)

	expectedToken := "cli-fallback-token"
	vaultTokenFile := filepath.Join(tempDir, ".vault-token")
	err := os.WriteFile(vaultTokenFile, []byte(expectedToken), 0600)
	if err != nil {
		t.Fatalf("Failed to create test vault token file: %v", err)
	}

	os.Unsetenv("VAULT_TOKEN")
	os.Setenv("HOME", tempDir)

	token, err := LookupToken()
	if err != nil {
		t.Fatalf("LookupToken() error = %v, want nil", err)
	}
	if token != expectedToken {
		t.Fatalf("LookupToken() = %v, want %v", token, expectedToken)
	}
}

func TestLookupToken_EnvPriorityOverCli(t *testing.T) {
	tempDir := t.TempDir()
	originalHomeDir := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHomeDir)

	expectedToken := "env-priority-token"
	cliToken := "cli-should-not-be-used"

	// Set up CLI token
	vaultTokenFile := filepath.Join(tempDir, ".vault-token")
	err := os.WriteFile(vaultTokenFile, []byte(cliToken), 0600)
	if err != nil {
		t.Fatalf("Failed to create test vault token file: %v", err)
	}

	// Set up env token (should take priority)
	os.Setenv("VAULT_TOKEN", expectedToken)
	defer os.Unsetenv("VAULT_TOKEN")
	os.Setenv("HOME", tempDir)

	token, err := LookupToken()
	if err != nil {
		t.Fatalf("LookupToken() error = %v, want nil", err)
	}
	if token != expectedToken {
		t.Fatalf("LookupToken() = %v, want %v (env should take priority)", token, expectedToken)
	}
}

func TestLookupToken_NoToken(t *testing.T) {
	os.Unsetenv("VAULT_TOKEN")
	originalHomeDir := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHomeDir)

	os.Setenv("HOME", "")
	_, err := LookupToken()
	if err == nil {
		t.Fatal("LookupToken() error = nil, want non-nil")
	}
}

func TestLookupToken_CliTokenFileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalHomeDir := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHomeDir)

	os.Unsetenv("VAULT_TOKEN")
	os.Setenv("HOME", tempDir)

	_, err := LookupToken()
	if err == nil {
		t.Fatal("LookupToken() error = nil, want non-nil")
	}
}
