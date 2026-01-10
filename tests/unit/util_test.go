package unit

import (
	"testing"

	"github.com/sv4u/spotigo"
)

func TestGetID(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		entityType  string
		expected    string
		expectError bool
	}{
		{"URI track", "spotify:track:6b2oQwSGFkzsMtQruIWm2p", "track", "6b2oQwSGFkzsMtQruIWm2p", false},
		{"URL track", "https://open.spotify.com/track/6b2oQwSGFkzsMtQruIWm2p", "track", "6b2oQwSGFkzsMtQruIWm2p", false},
		{"Raw ID", "6b2oQwSGFkzsMtQruIWm2p", "track", "6b2oQwSGFkzsMtQruIWm2p", false},
		{"URI artist", "spotify:artist:3jOstUTkEu2JkjvRdBA5Gu", "artist", "3jOstUTkEu2JkjvRdBA5Gu", false},
		{"Type mismatch", "spotify:track:6b2oQwSGFkzsMtQruIWm2p", "artist", "", true},
		{"Invalid URI", "spotify:invalid:123", "track", "", true},
		{"Invalid base62", "invalid-id-with-special-chars!", "track", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := spotigo.GetID(tc.input, tc.entityType)
			if tc.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if result != tc.expected {
					t.Errorf("expected %q, got %q", tc.expected, result)
				}
			}
		})
	}
}

func TestGetURI(t *testing.T) {
	testCases := []struct {
		name        string
		id          string
		entityType  string
		expected    string
		expectError bool
	}{
		{"Valid track", "6b2oQwSGFkzsMtQruIWm2p", "track", "spotify:track:6b2oQwSGFkzsMtQruIWm2p", false},
		{"Valid artist", "3jOstUTkEu2JkjvRdBA5Gu", "artist", "spotify:artist:3jOstUTkEu2JkjvRdBA5Gu", false},
		{"Invalid type", "6b2oQwSGFkzsMtQruIWm2p", "invalid", "", true},
		{"Invalid ID", "invalid-id!", "track", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := spotigo.GetURI(tc.id, tc.entityType)
			if tc.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if result != tc.expected {
					t.Errorf("expected %q, got %q", tc.expected, result)
				}
			}
		})
	}
}

func TestIsURI(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid track URI", "spotify:track:6b2oQwSGFkzsMtQruIWm2p", true},
		{"Valid artist URI", "spotify:artist:3jOstUTkEu2JkjvRdBA5Gu", true},
		{"Valid playlist URI", "spotify:playlist:37i9dQZF1DXcBWIGoYBM5M", true},
		{"Invalid URI", "spotify:invalid:123", false},
		{"Not a URI", "https://open.spotify.com/track/123", false},
		{"Raw ID", "6b2oQwSGFkzsMtQruIWm2p", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := spotigo.IsURI(tc.input)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestGetHostPort(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		expectedHost string
		expectedPort *int
	}{
		{"Host and port", "localhost:8080", "localhost", intPtr(8080)},
		{"Host only", "localhost", "localhost", nil},
		{"Empty", "", "", nil},
		{"Invalid port", "localhost:invalid", "localhost:invalid", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			host, port := spotigo.GetHostPort(tc.input)
			if host != tc.expectedHost {
				t.Errorf("expected host %q, got %q", tc.expectedHost, host)
			}
			if !equalIntPtr(port, tc.expectedPort) {
				t.Errorf("expected port %v, got %v", tc.expectedPort, port)
			}
		})
	}
}

func TestValidateCountryCode(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected bool
	}{
		{"Valid US", "US", true},
		{"Valid GB", "GB", true},
		{"Valid lowercase", "us", true}, // Should be case-insensitive
		{"Invalid code", "XX", false},
		{"Empty", "", false},
		{"Too long", "USA", false},
		{"Too short", "U", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := spotigo.ValidateCountryCode(tc.code)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func equalIntPtr(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
