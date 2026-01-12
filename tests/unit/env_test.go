package unit

import (
	"os"
	"path/filepath"
	"testing"

	spotigotests "github.com/sv4u/spotigo/tests"
)

func TestLoadEnvFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	
	// Create a minimal go.mod file to simulate project root
	goModPath := filepath.Join(tmpDir, "go.mod")
	goModContent := "module test\n\ngo 1.23\n"
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create go.mod file: %v", err)
	}
	
	// Create a .env file with test values
	envContent := `# Test comment
SPOTIGO_CLIENT_ID=test_client_id_from_env
SPOTIGO_CLIENT_SECRET=test_secret_from_env
SPOTIGO_REDIRECT_URI=http://localhost:8080/callback
SPOTIGO_CLIENT_USERNAME=test_user

# Another comment
# Empty line above
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() {
		// Restore original working directory
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Clear any existing environment variables
	os.Unsetenv("SPOTIGO_CLIENT_ID")
	os.Unsetenv("SPOTIGO_CLIENT_SECRET")
	os.Unsetenv("SPOTIGO_REDIRECT_URI")
	os.Unsetenv("SPOTIGO_CLIENT_USERNAME")

	// Load credentials (this should load from .env)
	creds := spotigotests.GetTestCredentials()

	if creds == nil {
		t.Fatal("expected credentials to be loaded from .env file, got nil")
	}

	if creds.ClientID != "test_client_id_from_env" {
		t.Errorf("expected ClientID to be 'test_client_id_from_env', got %q", creds.ClientID)
	}

	if creds.ClientSecret != "test_secret_from_env" {
		t.Errorf("expected ClientSecret to be 'test_secret_from_env', got %q", creds.ClientSecret)
	}

	if creds.RedirectURI != "http://localhost:8080/callback" {
		t.Errorf("expected RedirectURI to be 'http://localhost:8080/callback', got %q", creds.RedirectURI)
	}

	if creds.Username != "test_user" {
		t.Errorf("expected Username to be 'test_user', got %q", creds.Username)
	}
}

func TestLoadEnvFileWithQuotedValues(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a minimal go.mod file to simulate project root
	goModPath := filepath.Join(tmpDir, "go.mod")
	goModContent := "module test\n\ngo 1.23\n"
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create go.mod file: %v", err)
	}

	// Create a .env file with quoted values
	envContent := `SPOTIGO_CLIENT_ID="quoted_client_id"
SPOTIGO_CLIENT_SECRET='quoted_secret'
SPOTIGO_REDIRECT_URI="http://localhost:8080/callback"
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Clear any existing environment variables
	os.Unsetenv("SPOTIGO_CLIENT_ID")
	os.Unsetenv("SPOTIGO_CLIENT_SECRET")
	os.Unsetenv("SPOTIGO_REDIRECT_URI")

	// Load credentials
	creds := spotigotests.GetTestCredentials()

	if creds == nil {
		t.Fatal("expected credentials to be loaded from .env file, got nil")
	}

	// Check that quotes were removed
	if creds.ClientID != "quoted_client_id" {
		t.Errorf("expected ClientID to be 'quoted_client_id' (without quotes), got %q", creds.ClientID)
	}

	if creds.ClientSecret != "quoted_secret" {
		t.Errorf("expected ClientSecret to be 'quoted_secret' (without quotes), got %q", creds.ClientSecret)
	}
}

func TestEnvironmentVariablesTakePrecedence(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a minimal go.mod file to simulate project root
	goModPath := filepath.Join(tmpDir, "go.mod")
	goModContent := "module test\n\ngo 1.23\n"
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create go.mod file: %v", err)
	}

	// Create a .env file
	envContent := `SPOTIGO_CLIENT_ID=env_file_client_id
SPOTIGO_CLIENT_SECRET=env_file_secret
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
		// Clean up environment variables
		os.Unsetenv("SPOTIGO_CLIENT_ID")
		os.Unsetenv("SPOTIGO_CLIENT_SECRET")
	}()

	// Set environment variables (these should take precedence)
	os.Setenv("SPOTIGO_CLIENT_ID", "env_var_client_id")
	os.Setenv("SPOTIGO_CLIENT_SECRET", "env_var_secret")

	// Load credentials
	creds := spotigotests.GetTestCredentials()

	if creds == nil {
		t.Fatal("expected credentials to be loaded, got nil")
	}

	// Environment variables should take precedence over .env file
	if creds.ClientID != "env_var_client_id" {
		t.Errorf("expected ClientID to be 'env_var_client_id' (from env var), got %q", creds.ClientID)
	}

	if creds.ClientSecret != "env_var_secret" {
		t.Errorf("expected ClientSecret to be 'env_var_secret' (from env var), got %q", creds.ClientSecret)
	}
}

func TestGetTestCredentialsReturnsNilWhenMissing(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Create a temporary directory without .env file
	tmpDir := t.TempDir()

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Clear all environment variables
	os.Unsetenv("SPOTIGO_CLIENT_ID")
	os.Unsetenv("SPOTIGO_CLIENT_SECRET")
	os.Unsetenv("SPOTIGO_REDIRECT_URI")
	os.Unsetenv("SPOTIGO_CLIENT_USERNAME")

	// Load credentials - should return nil when no credentials available
	creds := spotigotests.GetTestCredentials()

	if creds != nil {
		t.Errorf("expected credentials to be nil when not available, got %+v", creds)
	}
}

func TestGetTestCredentialsWithPartialEnv(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Create a temporary directory
	tmpDir := t.TempDir()

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
		// Clean up
		os.Unsetenv("SPOTIGO_CLIENT_ID")
		os.Unsetenv("SPOTIGO_CLIENT_SECRET")
	}()

	// Set only one required variable - should return nil
	os.Setenv("SPOTIGO_CLIENT_ID", "test_id")
	os.Unsetenv("SPOTIGO_CLIENT_SECRET")

	creds := spotigotests.GetTestCredentials()

	if creds != nil {
		t.Errorf("expected credentials to be nil when ClientSecret is missing, got %+v", creds)
	}
}
