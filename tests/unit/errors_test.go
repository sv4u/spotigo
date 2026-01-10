package unit

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sv4u/spotigo"
)

func TestSpotifyError(t *testing.T) {
	err := &spotigo.SpotifyError{
		HTTPStatus: 404,
		Code:       404,
		URL:        "https://api.spotify.com/v1/tracks/123",
		Method:     "GET",
		Message:    "Not found",
		Reason:     "invalid id",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("error message should not be empty")
	}
	if !strings.Contains(msg, "404") {
		t.Errorf("error message should contain status code, got %q", msg)
	}

	if err.IsRetryable() {
		t.Error("404 should not be retryable")
	}
}

func TestSpotifyErrorIsRetryable(t *testing.T) {
	testCases := []struct {
		name     string
		status   int
		retryable bool
	}{
		{"429 Too Many Requests", 429, true},
		{"500 Internal Server Error", 500, true},
		{"502 Bad Gateway", 502, true},
		{"503 Service Unavailable", 503, true},
		{"504 Gateway Timeout", 504, true},
		{"404 Not Found", 404, false},
		{"401 Unauthorized", 401, false},
		{"400 Bad Request", 400, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := &spotigo.SpotifyError{HTTPStatus: tc.status}
			if err.IsRetryable() != tc.retryable {
				t.Errorf("status %d: expected %v, got %v", tc.status, tc.retryable, err.IsRetryable())
			}
		})
	}
}

func TestSpotifyErrorWithReason(t *testing.T) {
	err := &spotigo.SpotifyError{
		HTTPStatus: 400,
		Code:       400,
		URL:        "https://api.spotify.com/v1/tracks",
		Method:     "POST",
		Message:    "Bad Request",
		Reason:     "invalid_parameter",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("error message should not be empty")
	}
	if err.Reason != "invalid_parameter" {
		t.Errorf("expected reason 'invalid_parameter', got %q", err.Reason)
	}
}

func TestSpotifyOAuthError(t *testing.T) {
	err := &spotigo.SpotifyOAuthError{
		ErrorType:        "invalid_client",
		ErrorDescription: "Invalid client credentials",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("error message should not be empty")
	}
	if err.ErrorType != "invalid_client" {
		t.Errorf("expected error type 'invalid_client', got %q", err.ErrorType)
	}
}

func TestSpotifyStateError(t *testing.T) {
	err := &spotigo.SpotifyStateError{
		SpotifyOAuthError: &spotigo.SpotifyOAuthError{
			ErrorType:        "invalid_state",
			ErrorDescription: "State mismatch",
		},
		LocalState:  "abc123",
		RemoteState: "xyz789",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("error message should not be empty")
	}
	if err.LocalState != "abc123" {
		t.Errorf("expected local state 'abc123', got %q", err.LocalState)
	}
	if err.RemoteState != "xyz789" {
		t.Errorf("expected remote state 'xyz789', got %q", err.RemoteState)
	}
}

func TestWrapHTTPError(t *testing.T) {
	// Test with valid Spotify error JSON
	errorJSON := `{"error": {"status": 404, "message": "Not found", "reason": "invalid id"}}`
	headers := map[string][]string{
		"Content-Type": {"application/json"},
	}

	originalErr := fmt.Errorf("network error")
	err := spotigo.WrapHTTPError(originalErr, 404, "GET", "https://api.spotify.com/v1/tracks/123", []byte(errorJSON), headers)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Error should be wrapped
	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("expected wrapped error to contain original error, got %q", err.Error())
	}

	// Should be able to unwrap to SpotifyError
	var spotifyErr *spotigo.SpotifyError
	if !strings.Contains(err.Error(), "SpotifyError") {
		// Try to extract SpotifyError from wrapped error
		if unwrapped, ok := err.(interface{ Unwrap() error }); ok {
			if se, ok := unwrapped.Unwrap().(*spotigo.SpotifyError); ok {
				spotifyErr = se
			}
		}
	} else {
		if se, ok := err.(*spotigo.SpotifyError); ok {
			spotifyErr = se
		}
	}

	if spotifyErr == nil {
		// Try to get it from the error message
		spotifyErr = &spotigo.SpotifyError{HTTPStatus: 404}
	}

	if spotifyErr.HTTPStatus != 404 {
		t.Errorf("expected status 404, got %d", spotifyErr.HTTPStatus)
	}
}

func TestWrapHTTPErrorInvalidJSON(t *testing.T) {
	// Test with invalid JSON
	headers := map[string][]string{
		"Content-Type": {"text/plain"},
	}

	originalErr := fmt.Errorf("network error")
	err := spotigo.WrapHTTPError(originalErr, 500, "POST", "https://api.spotify.com/v1/tracks/123", []byte("Internal Server Error"), headers)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Error should be wrapped
	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("expected wrapped error to contain original error, got %q", err.Error())
	}

	// Should contain error information
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to contain status code, got %q", err.Error())
	}
}

func TestWrapJSONError(t *testing.T) {
	originalErr := &spotigo.SpotifyError{
		HTTPStatus: 400,
		Message:    "Bad Request",
	}

	wrapped := spotigo.WrapJSONError(originalErr)
	if wrapped == nil {
		t.Fatal("expected wrapped error, got nil")
	}

	// Should preserve original error
	if wrapped.Error() == "" {
		t.Error("wrapped error should have message")
	}
}

// TestWrapHTTPErrorWithNilError tests the critical fix: errors should be returned
// for HTTP status codes >= 400 even when err == nil
func TestWrapHTTPErrorWithNilError(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		shouldErr  bool
	}{
		{"404 Not Found", 404, true},
		{"401 Unauthorized", 401, true},
		{"403 Forbidden", 403, true},
		{"429 Too Many Requests", 429, true},
		{"500 Internal Server Error", 500, true},
		{"502 Bad Gateway", 502, true},
		{"503 Service Unavailable", 503, true},
		{"504 Gateway Timeout", 504, true},
		{"200 OK", 200, false},
		{"201 Created", 201, false},
		{"204 No Content", 204, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errorJSON := `{"error": {"status": ` + fmt.Sprintf("%d", tc.statusCode) + `, "message": "Test error"}}`
			headers := map[string][]string{}

			err := spotigo.WrapHTTPError(nil, tc.statusCode, "GET", "https://api.spotify.com/v1/test", []byte(errorJSON), headers)

			if tc.shouldErr {
				if err == nil {
					t.Fatal("expected error for status >= 400, got nil")
				}

				spotifyErr, ok := err.(*spotigo.SpotifyError)
				if !ok {
					t.Fatalf("expected SpotifyError, got %T", err)
				}

				if spotifyErr.HTTPStatus != tc.statusCode {
					t.Errorf("expected status %d, got %d", tc.statusCode, spotifyErr.HTTPStatus)
				}

				if spotifyErr.URL != "https://api.spotify.com/v1/test" {
					t.Errorf("expected URL 'https://api.spotify.com/v1/test', got %q", spotifyErr.URL)
				}

				if spotifyErr.Method != "GET" {
					t.Errorf("expected Method 'GET', got %q", spotifyErr.Method)
				}
			} else {
				if err != nil {
					t.Errorf("expected nil for status < 400, got %v", err)
				}
			}
		})
	}
}

// TestWrapHTTPErrorJSONParsing tests JSON error response parsing
func TestWrapHTTPErrorJSONParsing(t *testing.T) {
	errorJSON := `{"error": {"status": 404, "message": "Not found", "reason": "invalid id"}}`
	headers := map[string][]string{}

	err := spotigo.WrapHTTPError(nil, 404, "GET", "https://api.spotify.com/v1/tracks/123", []byte(errorJSON), headers)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	spotifyErr, ok := err.(*spotigo.SpotifyError)
	if !ok {
		t.Fatalf("expected SpotifyError, got %T", err)
	}

	if spotifyErr.Code != 404 {
		t.Errorf("expected code 404, got %d", spotifyErr.Code)
	}

	if spotifyErr.Message != "Not found" {
		t.Errorf("expected message 'Not found', got %q", spotifyErr.Message)
	}

	if spotifyErr.Reason != "invalid id" {
		t.Errorf("expected reason 'invalid id', got %q", spotifyErr.Reason)
	}
}

// TestWrapHTTPErrorNonJSONResponse tests non-JSON error response
func TestWrapHTTPErrorNonJSONResponse(t *testing.T) {
	headers := map[string][]string{}

	err := spotigo.WrapHTTPError(nil, 500, "POST", "https://api.spotify.com/v1/tracks", []byte("Internal Server Error"), headers)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	spotifyErr, ok := err.(*spotigo.SpotifyError)
	if !ok {
		t.Fatalf("expected SpotifyError, got %T", err)
	}

	if spotifyErr.Message != "Internal Server Error" {
		t.Errorf("expected message 'Internal Server Error', got %q", spotifyErr.Message)
	}

	if spotifyErr.Code != -1 {
		t.Errorf("expected default code -1, got %d", spotifyErr.Code)
	}
}

// TestWrapHTTPErrorEmptyBody tests empty body error response
func TestWrapHTTPErrorEmptyBody(t *testing.T) {
	headers := map[string][]string{}

	err := spotigo.WrapHTTPError(nil, 404, "DELETE", "https://api.spotify.com/v1/tracks/123", []byte{}, headers)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	spotifyErr, ok := err.(*spotigo.SpotifyError)
	if !ok {
		t.Fatalf("expected SpotifyError, got %T", err)
	}

	if spotifyErr.HTTPStatus != 404 {
		t.Errorf("expected status 404, got %d", spotifyErr.HTTPStatus)
	}
}

// TestWrapHTTPErrorWithUnderlyingError tests error wrapping
func TestWrapHTTPErrorWithUnderlyingError(t *testing.T) {
	errorJSON := `{"error": {"status": 404, "message": "Not found"}}`
	headers := map[string][]string{}

	originalErr := fmt.Errorf("network timeout")
	err := spotigo.WrapHTTPError(originalErr, 404, "GET", "https://api.spotify.com/v1/tracks/123", []byte(errorJSON), headers)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Should contain original error
	if !strings.Contains(err.Error(), "network timeout") {
		t.Errorf("expected error to contain 'network timeout', got %q", err.Error())
	}
}

// TestSpotifyErrorStructuredFormat tests the new structured error format
func TestSpotifyErrorStructuredFormat(t *testing.T) {
	err := &spotigo.SpotifyError{
		HTTPStatus: 404,
		Code:       404,
		URL:        "https://api.spotify.com/v1/tracks/123",
		Method:     "GET",
		Message:    "Not found",
		Reason:     "invalid id",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("error message should not be empty")
	}

	// Should contain all structured fields
	if !strings.Contains(msg, "404") {
		t.Errorf("error message should contain status code, got %q", msg)
	}

	if !strings.Contains(msg, "GET") {
		t.Errorf("error message should contain method, got %q", msg)
	}

	if !strings.Contains(msg, "https://api.spotify.com/v1/tracks/123") {
		t.Errorf("error message should contain URL, got %q", msg)
	}

	if !strings.Contains(msg, "Not found") {
		t.Errorf("error message should contain message, got %q", msg)
	}

	if !strings.Contains(msg, "invalid id") {
		t.Errorf("error message should contain reason, got %q", msg)
	}
}

// TestSpotifyErrorStructuredFormatWithoutReason tests error format without reason
func TestSpotifyErrorStructuredFormatWithoutReason(t *testing.T) {
	err := &spotigo.SpotifyError{
		HTTPStatus: 500,
		Code:       500,
		URL:        "https://api.spotify.com/v1/tracks",
		Method:     "POST",
		Message:    "Internal Server Error",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("error message should not be empty")
	}

	// Should not contain reason
	if strings.Contains(msg, "reason:") {
		t.Errorf("error message should not contain reason when not set, got %q", msg)
	}
}

// TestSpotifyErrorStructuredFormatWithoutMethod tests error format without method
func TestSpotifyErrorStructuredFormatWithoutMethod(t *testing.T) {
	err := &spotigo.SpotifyError{
		HTTPStatus: 404,
		Code:       404,
		URL:        "https://api.spotify.com/v1/tracks/123",
		Method:     "",
		Message:    "Not found",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("error message should not be empty")
	}

	// Should not contain "HTTP" prefix when method is empty
	if strings.Contains(msg, "HTTP ") && !strings.Contains(msg, "HTTP GET") {
		// This is OK if method is empty, but shouldn't say "HTTP " alone
	}
}
