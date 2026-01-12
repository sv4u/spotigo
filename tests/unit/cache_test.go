package unit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sv4u/spotigo"
)

func TestFileCacheHandler(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, ".cache")

	handler, err := spotigo.NewFileCacheHandler(cachePath, "testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token := &spotigo.TokenInfo{
		AccessToken: "test_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   int(time.Now().Unix()) + 3600,
	}

	ctx := context.Background()

	// Test save
	if err := handler.SaveTokenToCache(ctx, token); err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	// Test read
	cached, err := handler.GetCachedToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error reading: %v", err)
	}

	if cached == nil {
		t.Fatal("expected cached token, got nil")
	}

	if cached.AccessToken != token.AccessToken {
		t.Errorf("expected %q, got %q", token.AccessToken, cached.AccessToken)
	}

	// Test permissions
	info, err := os.Stat(cachePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestFileCacheHandlerNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, ".cache-nonexistent")

	handler, err := spotigo.NewFileCacheHandler(cachePath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()
	cached, err := handler.GetCachedToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return nil, nil when file doesn't exist (not an error)
	if cached != nil {
		t.Error("expected nil when file doesn't exist")
	}
}

func TestFileCacheHandlerInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, ".cache")

	handler, err := spotigo.NewFileCacheHandler(cachePath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Write invalid JSON
	if err := os.WriteFile(cachePath, []byte("invalid json"), 0600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()
	cached, err := handler.GetCachedToken(ctx)
	// Implementation may return error or handle gracefully
	// Check that it doesn't return a valid token for invalid JSON
	if err == nil && cached != nil {
		t.Error("expected error or nil token for invalid JSON, got valid token")
	}
	// If error is returned, that's fine
	// If nil token is returned without error, that's also acceptable (graceful handling)
}

func TestMemoryCacheHandler(t *testing.T) {
	handler := spotigo.NewMemoryCacheHandler()

	token := &spotigo.TokenInfo{
		AccessToken: "test_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   int(time.Now().Unix()) + 3600,
	}

	ctx := context.Background()

	// Test save
	if err := handler.SaveTokenToCache(ctx, token); err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	// Test read
	cached, err := handler.GetCachedToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error reading: %v", err)
	}

	if cached == nil {
		t.Fatal("expected cached token, got nil")
	}

	if cached.AccessToken != token.AccessToken {
		t.Errorf("expected %q, got %q", token.AccessToken, cached.AccessToken)
	}
}

func TestMemoryCacheHandlerNoToken(t *testing.T) {
	handler := spotigo.NewMemoryCacheHandler()

	ctx := context.Background()
	cached, err := handler.GetCachedToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return nil, nil when no token cached (not an error)
	if cached != nil {
		t.Error("expected nil when no token cached")
	}
}

func TestMemoryCacheHandlerIsolation(t *testing.T) {
	handler := spotigo.NewMemoryCacheHandler()

	token := &spotigo.TokenInfo{
		AccessToken: "original_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   int(time.Now().Unix()) + 3600,
	}

	ctx := context.Background()

	// Save token
	if err := handler.SaveTokenToCache(ctx, token); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Retrieve and modify
	cached, err := handler.GetCachedToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cached.AccessToken = "modified_token"

	// Retrieve again - should still have original
	cached2, err := handler.GetCachedToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cached2.AccessToken != "original_token" {
		t.Errorf("expected 'original_token', got %q (token was modified externally)", cached2.AccessToken)
	}
}

func TestMemoryCacheHandlerConcurrent(t *testing.T) {
	handler := spotigo.NewMemoryCacheHandler()

	ctx := context.Background()

	// Test concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			token := &spotigo.TokenInfo{
				AccessToken: "token_" + string(rune(id)),
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				ExpiresAt:   int(time.Now().Unix()) + 3600,
			}
			handler.SaveTokenToCache(ctx, token)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should be able to read without error
	_, err := handler.GetCachedToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestAtomicFileCacheWrite verifies atomic write pattern (temp file + rename)
func TestAtomicFileCacheWrite(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, ".cache")

	handler, err := spotigo.NewFileCacheHandler(cachePath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token := &spotigo.TokenInfo{
		AccessToken: "test_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   int(time.Now().Unix()) + 3600,
	}

	ctx := context.Background()

	// Save token - should use atomic write
	if err := handler.SaveTokenToCache(ctx, token); err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	// Verify temp file doesn't exist (should be cleaned up)
	tmpPath := cachePath + ".tmp"
	if _, err := os.Stat(tmpPath); err == nil {
		t.Error("temp file should not exist after successful write")
	}

	// Verify cache file exists and is valid
	cached, err := handler.GetCachedToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error reading: %v", err)
	}

	if cached == nil {
		t.Fatal("expected cached token, got nil")
	}

	if cached.AccessToken != token.AccessToken {
		t.Errorf("expected %q, got %q", token.AccessToken, cached.AccessToken)
	}
}

// TestAtomicFileCacheWriteConcurrent verifies file locking prevents concurrent write corruption
func TestAtomicFileCacheWriteConcurrent(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, ".cache")

	handler, err := spotigo.NewFileCacheHandler(cachePath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()

	// Concurrent writes from multiple goroutines
	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			token := &spotigo.TokenInfo{
				AccessToken: fmt.Sprintf("token_%d", id),
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				ExpiresAt:   int(time.Now().Unix()) + 3600,
			}
			done <- handler.SaveTokenToCache(ctx, token)
		}(i)
	}

	// Wait for all writes to complete
	var errors []error
	for i := 0; i < 10; i++ {
		if err := <-done; err != nil {
			errors = append(errors, err)
		}
	}

	// All writes should succeed (locking prevents corruption)
	if len(errors) > 0 {
		t.Errorf("unexpected errors during concurrent writes: %v", errors)
	}

	// Verify cache file is valid JSON (not corrupted)
	cached, err := handler.GetCachedToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error reading: %v", err)
	}

	if cached == nil {
		t.Fatal("expected cached token, got nil")
	}

	// Token should be one of the written tokens (last write wins)
	if cached.AccessToken == "" {
		t.Error("expected valid token, got empty")
	}
}

// TestAtomicFileCacheWriteLockFile verifies lock file is created and cleaned up
func TestAtomicFileCacheWriteLockFile(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, ".cache")
	lockPath := cachePath + ".lock"

	handler, err := spotigo.NewFileCacheHandler(cachePath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token := &spotigo.TokenInfo{
		AccessToken: "test_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   int(time.Now().Unix()) + 3600,
	}

	ctx := context.Background()

	// Save token
	if err := handler.SaveTokenToCache(ctx, token); err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	// Lock file should be cleaned up after write
	if _, err := os.Stat(lockPath); err == nil {
		t.Error("lock file should be cleaned up after write")
	}
}

// TestAtomicFileCacheWriteInterrupted verifies cache file is not corrupted if write is interrupted
func TestAtomicFileCacheWriteInterrupted(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, ".cache")

	handler, err := spotigo.NewFileCacheHandler(cachePath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Save initial token
	initialToken := &spotigo.TokenInfo{
		AccessToken: "initial_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   int(time.Now().Unix()) + 3600,
	}

	ctx := context.Background()
	if err := handler.SaveTokenToCache(ctx, initialToken); err != nil {
		t.Fatalf("unexpected error saving initial token: %v", err)
	}

	// Simulate interrupted write by creating a temp file but not renaming it
	tmpPath := cachePath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte("invalid json"), 0600); err != nil {
		t.Fatalf("unexpected error creating temp file: %v", err)
	}

	// Try to read cache - should still get initial token (not corrupted)
	cached, err := handler.GetCachedToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error reading: %v", err)
	}

	// Should get initial token, not corrupted temp file
	if cached != nil && cached.AccessToken != "initial_token" {
		t.Errorf("expected initial token, got %q (cache may be corrupted)", cached.AccessToken)
	}

	// Cleanup temp file
	os.Remove(tmpPath)
}

// TestFileCacheHandlerReturnsCopy verifies that FileCacheHandler returns a copy (consistent with MemoryCacheHandler)
func TestFileCacheHandlerReturnsCopy(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, ".cache")

	handler, err := spotigo.NewFileCacheHandler(cachePath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token := &spotigo.TokenInfo{
		AccessToken: "original_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   int(time.Now().Unix()) + 3600,
	}

	ctx := context.Background()

	// Save token
	if err := handler.SaveTokenToCache(ctx, token); err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	// Retrieve and modify
	cached, err := handler.GetCachedToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error reading: %v", err)
	}

	if cached == nil {
		t.Fatal("expected cached token, got nil")
	}

	cached.AccessToken = "modified_token"

	// Retrieve again - should still have original (copy prevents modification)
	cached2, err := handler.GetCachedToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error reading: %v", err)
	}

	if cached2 == nil {
		t.Fatal("expected cached token, got nil")
	}

	if cached2.AccessToken != "original_token" {
		t.Errorf("expected 'original_token', got %q (token was modified externally - copy not returned)", cached2.AccessToken)
	}
}

// TestFileCacheHandlerSanitizeUsername verifies that invalid filename characters are sanitized
func TestFileCacheHandlerSanitizeUsername(t *testing.T) {
	testCases := []struct {
		name           string
		username       string
		expectedSuffix string // Expected suffix in cache path
	}{
		{
			name:           "username with forward slash",
			username:       "user/name",
			expectedSuffix: "-user_name",
		},
		{
			name:           "username with backslash",
			username:       "user\\name",
			expectedSuffix: "-user_name",
		},
		{
			name:           "username with colon",
			username:       "user:name",
			expectedSuffix: "-user_name",
		},
		{
			name:           "username with asterisk",
			username:       "user*name",
			expectedSuffix: "-user_name",
		},
		{
			name:           "username with question mark",
			username:       "user?name",
			expectedSuffix: "-user_name",
		},
		{
			name:           "username with quotes",
			username:       "user\"name",
			expectedSuffix: "-user_name",
		},
		{
			name:           "username with angle brackets",
			username:       "user<name>",
			expectedSuffix: "-user_name_",
		},
		{
			name:           "username with pipe",
			username:       "user|name",
			expectedSuffix: "-user_name",
		},
		{
			name:           "username with multiple invalid chars",
			username:       "user/name:test*file?",
			expectedSuffix: "-user_name_test_file_",
		},
		{
			name:           "normal username",
			username:       "normaluser",
			expectedSuffix: "-normaluser",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create handler with username (will use default cache path with sanitized username)
			handler, err := spotigo.NewFileCacheHandler("", tc.username)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check that cache path ends with expected suffix
			if !strings.HasSuffix(handler.CachePath, tc.expectedSuffix) {
				t.Errorf("expected cache path to end with %q, got %q", tc.expectedSuffix, handler.CachePath)
			}

			// Verify no invalid characters in filename (not the full path)
			// Get just the filename from the path
			filename := filepath.Base(handler.CachePath)
			invalidChars := []string{"\\", ":", "*", "?", "\"", "<", ">", "|"}
			for _, char := range invalidChars {
				if strings.Contains(filename, char) {
					t.Errorf("cache filename contains invalid character %q: %s (full path: %s)", char, filename, handler.CachePath)
				}
			}
		})
	}
}

// TestFileCacheHandlerWithUsername verifies cache path generation with username
func TestFileCacheHandlerWithUsername(t *testing.T) {
	// Test with username - should use .cache-{sanitized_username}
	handler, err := spotigo.NewFileCacheHandler("", "testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Cache path should end with .cache-testuser
	if !strings.HasSuffix(handler.CachePath, ".cache-testuser") {
		t.Errorf("expected cache path to end with .cache-testuser, got %q", handler.CachePath)
	}

	// Test with empty username - should use .cache
	handler2, err := spotigo.NewFileCacheHandler("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasSuffix(handler2.CachePath, ".cache") {
		t.Errorf("expected cache path to end with .cache, got %q", handler2.CachePath)
	}
}
