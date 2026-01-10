package spotigo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// CacheHandler defines the interface for token cache implementations.
//
// Implementations store and retrieve OAuth2 tokens to avoid re-authentication.
// The library provides two implementations:
//   - FileCacheHandler for persistent file-based caching
//   - MemoryCacheHandler for in-memory caching
type CacheHandler interface {
	// GetCachedToken retrieves a cached token from storage
	// Returns (nil, nil) if token not found (not an error condition)
	// Returns (nil, error) if cache read fails
	GetCachedToken(ctx context.Context) (*TokenInfo, error)

	// SaveTokenToCache saves a token to cache storage
	// Returns error if cache write fails
	SaveTokenToCache(ctx context.Context, token *TokenInfo) error
}

// FileCacheHandler implements file-based token caching
type FileCacheHandler struct {
	CachePath string
	Username  string
}

const (
	// EnvCachePath is the environment variable for custom cache path
	EnvCachePath = "SPOTIPY_CACHE_PATH"
	// DefaultCacheFilename is the default cache filename
	DefaultCacheFilename = ".cache"
)

// NewFileCacheHandler creates a new file cache handler
func NewFileCacheHandler(cachePath, username string) (*FileCacheHandler, error) {
	handler := &FileCacheHandler{
		Username: username,
	}

	// Resolve cache path
	if cachePath == "" {
		// Check environment variable
		cachePath = os.Getenv(EnvCachePath)
	}

	if cachePath == "" {
		// Use default: .cache or .cache-{username}
		if username != "" {
			cachePath = fmt.Sprintf("%s-%s", DefaultCacheFilename, sanitizeUsername(username))
		} else {
			cachePath = DefaultCacheFilename
		}
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve cache path: %w", err)
	}

	handler.CachePath = absPath

	return handler, nil
}

// sanitizeUsername removes invalid filename characters from username
func sanitizeUsername(username string) string {
	// Remove invalid characters for filenames
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := username
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}

// GetCachedToken retrieves token from cache file
// Returns (nil, nil) if file doesn't exist (not an error)
// Returns (nil, error) for actual errors
func (f *FileCacheHandler) GetCachedToken(ctx context.Context) (*TokenInfo, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Try to open and read file
	data, err := os.ReadFile(f.CachePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist - not an error, just return nil
			log.Printf("cache does not exist at: %s", f.CachePath)
			return nil, nil
		}
		// Other file errors - log as error but return nil, nil per Spotipy behavior
		log.Printf("ERROR: Couldn't read cache at: %s: %v", f.CachePath, err)
		return nil, nil // Return nil, nil per Spotipy behavior (log error but don't return error)
	}

	// Parse JSON
	var tokenInfo TokenInfo
	if err := json.Unmarshal(data, &tokenInfo); err != nil {
		log.Printf("Couldn't decode JSON from cache at: %s: %v", f.CachePath, err)
		return nil, nil // Return nil, nil per Spotipy behavior
	}

	// Return copy to prevent external modification (consistent with MemoryCacheHandler)
	tokenCopy := tokenInfo
	return &tokenCopy, nil
}

// acquireLock attempts to acquire a file lock for cache writes
// Returns a function to release the lock, or an error if lock acquisition fails
func (f *FileCacheHandler) acquireLock(ctx context.Context) (func(), error) {
	lockPath := f.CachePath + ".lock"
	
	// Try to create lock file with exclusive flag (non-blocking)
	// O_CREATE|O_EXCL ensures only one process can create the file
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		if os.IsExist(err) {
			// Lock file exists - another process is writing
			// Wait briefly and retry a few times
			for i := 0; i < 5; i++ {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(100 * time.Millisecond):
				}
				lockFile, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
				if err == nil {
					break
				}
			}
			if err != nil {
				// Couldn't acquire lock after retries - proceed anyway
				// Atomic rename will handle the race condition
				log.Printf("Warning: Couldn't acquire cache lock at: %s (proceeding anyway): %v", lockPath, err)
				return func() {}, nil
			}
		} else {
			return nil, fmt.Errorf("failed to create lock file: %w", err)
		}
	}
	
	// Write process ID to lock file (helpful for debugging)
	lockFile.WriteString(fmt.Sprintf("%d\n", os.Getpid()))
	lockFile.Close()
	
	// Return release function
	return func() {
		os.Remove(lockPath)
	}, nil
}

// SaveTokenToCache saves token to cache file using atomic write pattern
func (f *FileCacheHandler) SaveTokenToCache(ctx context.Context, token *TokenInfo) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if token == nil {
		return fmt.Errorf("token is nil")
	}

	// Marshal to JSON
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(f.CachePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			log.Printf("Warning: Couldn't create cache directory: %v", err)
			// Continue anyway - file write might still work
		}
	}

	// Acquire lock for concurrent write protection
	releaseLock, err := f.acquireLock(ctx)
	if err != nil {
		log.Printf("Warning: Couldn't acquire cache lock: %v (proceeding anyway)", err)
		releaseLock = func() {} // No-op release function
	}
	defer releaseLock()

	// Write to temporary file first (atomic write pattern)
	tmpPath := f.CachePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		log.Printf("Warning: Couldn't write cache temp file at: %s: %v", tmpPath, err)
		// Cleanup temp file on error
		os.Remove(tmpPath)
		return nil // Don't fail - log warning and continue per Spotipy behavior
	}

	// Atomically rename temp file to final location
	// This is atomic on most filesystems and prevents corruption from crashes
	if err := os.Rename(tmpPath, f.CachePath); err != nil {
		log.Printf("Warning: Couldn't rename cache file from %s to %s: %v", tmpPath, f.CachePath, err)
		// Cleanup temp file on error
		os.Remove(tmpPath)
		return nil // Don't fail - log warning and continue per Spotipy behavior
	}

	// Set permissions to 0600 (after writing, as per Spotipy)
	if err := os.Chmod(f.CachePath, 0600); err != nil {
		log.Printf("Warning: Couldn't set cache permissions at: %s: %v", f.CachePath, err)
		// Don't fail - log warning and continue
	}

	return nil
}

// MemoryCacheHandler implements in-memory token caching
type MemoryCacheHandler struct {
	Token *TokenInfo
	mu    sync.RWMutex
}

// NewMemoryCacheHandler creates a new in-memory cache handler
func NewMemoryCacheHandler() *MemoryCacheHandler {
	return &MemoryCacheHandler{}
}

// GetCachedToken retrieves token from memory
func (m *MemoryCacheHandler) GetCachedToken(ctx context.Context) (*TokenInfo, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.Token == nil {
		return nil, nil // Return (nil, nil) when no token cached (not an error)
	}

	// Return copy to prevent external modification
	tokenCopy := *m.Token
	return &tokenCopy, nil
}

// SaveTokenToCache saves token to memory
func (m *MemoryCacheHandler) SaveTokenToCache(ctx context.Context, token *TokenInfo) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if token == nil {
		return fmt.Errorf("token is nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Store copy to prevent external modification
	tokenCopy := *token
	m.Token = &tokenCopy

	return nil
}
