package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/sv4u/spotigo"
)

// TestCredentials holds test credentials
type TestCredentials struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Username     string
}

// GetTestCredentials retrieves test credentials from environment variables
// Returns nil if credentials are not available
func GetTestCredentials() *TestCredentials {
	clientID := os.Getenv("SPOTIPY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIPY_CLIENT_SECRET")
	redirectURI := os.Getenv("SPOTIPY_REDIRECT_URI")
	username := os.Getenv("SPOTIPY_CLIENT_USERNAME")

	if clientID == "" || clientSecret == "" {
		return nil
	}

	return &TestCredentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		Username:     username,
	}
}

// SkipIfNoCredentials skips the test if credentials are not available
func SkipIfNoCredentials(t *testing.T) {
	if GetTestCredentials() == nil {
		t.Skip("Skipping test: SPOTIPY_CLIENT_ID and SPOTIPY_CLIENT_SECRET environment variables not set")
	}
}

// NewTestClient creates a test client with client credentials auth
func NewTestClient(t *testing.T) (*spotigo.Client, error) {
	creds := GetTestCredentials()
	if creds == nil {
		return nil, fmt.Errorf("test credentials not available")
	}

	auth, err := spotigo.NewClientCredentials(creds.ClientID, creds.ClientSecret)
	if err != nil {
		return nil, err
	}
	cache := spotigo.NewMemoryCacheHandler()
	client, err := spotigo.NewClient(auth, spotigo.WithCacheHandler(cache))
	if err != nil {
		return nil, err
	}

	return client, nil
}

// NewTestClientWithUserAuth creates a test client with user authentication
// This requires a valid token from OAuth flow
func NewTestClientWithUserAuth(t *testing.T, token *spotigo.TokenInfo) (*spotigo.Client, error) {
	creds := GetTestCredentials()
	if creds == nil {
		return nil, fmt.Errorf("test credentials not available")
	}

	// Create a mock auth manager that returns the provided token
	auth := &MockAuthManager{
		Token: token,
	}

	client, err := spotigo.NewClient(auth)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// MockAuthManager is a simple mock auth manager for testing
type MockAuthManager struct {
	Token *spotigo.TokenInfo
}

func (m *MockAuthManager) GetAccessToken(ctx context.Context) (string, error) {
	if m.Token == nil {
		return "", fmt.Errorf("no token available")
	}
	return m.Token.AccessToken, nil
}

func (m *MockAuthManager) GetCachedToken(ctx context.Context) (*spotigo.TokenInfo, error) {
	if m.Token == nil {
		return nil, nil
	}
	return m.Token, nil
}

func (m *MockAuthManager) RefreshToken(ctx context.Context) error {
	// Mock implementation - no-op for testing
	return nil
}

// NewMockHTTPServer creates a mock HTTP server for testing
func NewMockHTTPServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// LoadFixture loads a JSON fixture file
func LoadFixture(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// AssertErrorType checks if an error is of a specific type
func AssertErrorType(t *testing.T, err error, expectedType interface{}) {
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	switch expectedType.(type) {
	case *spotigo.SpotifyError:
		if _, ok := err.(*spotigo.SpotifyError); !ok {
			t.Errorf("expected SpotifyError, got %T", err)
		}
	case *spotigo.SpotifyOAuthError:
		if _, ok := err.(*spotigo.SpotifyOAuthError); !ok {
			t.Errorf("expected SpotifyOAuthError, got %T", err)
		}
	case *spotigo.SpotifyStateError:
		if _, ok := err.(*spotigo.SpotifyStateError); !ok {
			t.Errorf("expected SpotifyStateError, got %T", err)
		}
	default:
		t.Errorf("unsupported error type: %T", expectedType)
	}
}

// AssertEqual performs a deep equality check
func AssertEqual(t *testing.T, expected, actual interface{}, msg string) {
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// Known test data (from Spotipy tests)
var (
	TestTrackURI   = "spotify:track:6b2oQwSGFkzsMtQruIWm2p" // Creep
	TestTrackID    = "6b2oQwSGFkzsMtQruIWm2p"
	TestTrackURL   = "http://open.spotify.com/track/6b2oQwSGFkzsMtQruIWm2p"
	TestArtistURI  = "spotify:artist:3jOstUTkEu2JkjvRdBA5Gu" // Weezer
	TestArtistID   = "3jOstUTkEu2JkjvRdBA5Gu"
	TestAlbumURI   = "spotify:album:04xe676vyiTeYNXw15o9jT" // Pinkerton
	TestAlbumID    = "04xe676vyiTeYNXw15o9jT"
	TestShowURI    = "spotify:show:5c26B28vZMN8PG0Nppmn5G"
	TestShowID     = "5c26B28vZMN8PG0Nppmn5G"
	TestPlaylistID = "2oCEWyyAPbZp9xhVSxZavx"
)

// CreateTokenResponse creates a mock token response
func CreateTokenResponse(accessToken string, expiresIn int) map[string]interface{} {
	return map[string]interface{}{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   expiresIn,
		"scope":        "user-read-private",
	}
}

// CreateErrorResponse creates a mock error response
func CreateErrorResponse(status int, message string, reason string) map[string]interface{} {
	return map[string]interface{}{
		"error": map[string]interface{}{
			"status":  status,
			"message": message,
			"reason":  reason,
		},
	}
}

// WriteJSONResponse writes a JSON response to the HTTP response writer
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// WaitForContext waits for context to be done or timeout
func WaitForContext(ctx context.Context, timeout time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(timeout):
		return fmt.Errorf("timeout after %v", timeout)
	}
}
