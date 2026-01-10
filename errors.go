package spotigo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SpotifyBaseException is a marker interface for all Spotify-specific errors.
//
// All Spotify error types implement this interface to allow type checking.
type SpotifyBaseException interface {
	error
	isSpotifyError()
}

// SpotifyError represents an error returned by the Spotify API.
//
// It includes the HTTP status code, error code, message, and optional reason.
// Use IsRetryable() to check if the error indicates a retryable condition.
//
// Example:
//
//	track, err := client.Track(ctx, trackID)
//	if err != nil {
//		if spotifyErr, ok := err.(*SpotifyError); ok {
//			if spotifyErr.IsRetryable() {
//				// Retry the request
//			}
//		}
//	}
type SpotifyError struct {
	HTTPStatus int
	Code       int
	URL        string // Request URL (separate from message)
	Method     string // HTTP method (optional)
	Message    string // Error message (without URL prefix)
	Reason     string
	Headers    map[string][]string
}

// Error implements the error interface with structured format
func (e *SpotifyError) Error() string {
	var parts []string
	if e.Method != "" {
		parts = append(parts, fmt.Sprintf("HTTP %s", e.Method))
	}
	if e.URL != "" {
		parts = append(parts, e.URL)
	}
	if len(parts) > 0 {
		parts = append(parts, ":")
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	if e.Reason != "" {
		parts = append(parts, fmt.Sprintf("(reason: %s)", e.Reason))
	}
	message := strings.Join(parts, " ")
	if message == "" {
		return fmt.Sprintf("http status: %d, code: %d", e.HTTPStatus, e.Code)
	}
	return fmt.Sprintf("http status: %d, code: %d - %s", e.HTTPStatus, e.Code, message)
}

// isSpotifyError marks this as a Spotify error
func (e *SpotifyError) isSpotifyError() {}

// IsRetryable returns true if the error indicates a retryable condition
func (e *SpotifyError) IsRetryable() bool {
	return e.HTTPStatus == 429 ||
		(e.HTTPStatus >= 500 && e.HTTPStatus < 600)
}

// RetryAfter extracts the Retry-After header value if present
func (e *SpotifyError) RetryAfter() (time.Duration, bool) {
	if e.Headers == nil {
		return 0, false
	}
	retryAfterValues, ok := e.Headers["Retry-After"]
	if !ok || len(retryAfterValues) == 0 {
		return 0, false
	}
	retryAfter := retryAfterValues[0]
	
	// Try parsing as integer seconds first
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return time.Duration(seconds) * time.Second, true
	}
	// Try parsing as HTTP-date
	if t, err := http.ParseTime(retryAfter); err == nil {
		delay := time.Until(t)
		if delay > 0 {
			return delay, true
		}
		// HTTP-date in past, return 0
		return 0, false
	}
	return 0, false
}

// SpotifyOAuthError represents an OAuth2 authentication error
type SpotifyOAuthError struct {
	ErrorType        string
	ErrorDescription string
	AdditionalFields map[string]interface{}
}

// Error implements the error interface
func (e *SpotifyOAuthError) Error() string {
	if e.ErrorDescription != "" {
		return fmt.Sprintf("error: %s, error_description: %s", e.ErrorType, e.ErrorDescription)
	}
	return fmt.Sprintf("error: %s", e.ErrorType)
}

// isSpotifyError marks this as a Spotify error
func (e *SpotifyOAuthError) isSpotifyError() {}

// SpotifyStateError represents a state mismatch error in OAuth flow
type SpotifyStateError struct {
	*SpotifyOAuthError
	LocalState  string
	RemoteState string
}

// Error implements the error interface
func (e *SpotifyStateError) Error() string {
	return fmt.Sprintf("State mismatch: expected %q, got %q. %s",
		e.LocalState, e.RemoteState, e.SpotifyOAuthError.Error())
}

// isSpotifyError marks this as a Spotify error
func (e *SpotifyStateError) isSpotifyError() {}

// ErrorResponse represents the JSON structure of Spotify error responses
type ErrorResponse struct {
	Error struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Reason  string `json:"reason"`
	} `json:"error"`
}

// WrapHTTPError wraps an HTTP error with Spotify error information.
// Always creates a SpotifyError for HTTP error status codes (>= 400), even if err is nil.
// If err is provided, it will be wrapped; otherwise, the SpotifyError is returned directly.
func WrapHTTPError(err error, statusCode int, method string, url string, body []byte, headers map[string][]string) error {
	// Always create error for HTTP error status codes (>= 400), even if err is nil
	if statusCode < 400 {
		// Not an error status code, only wrap if err is provided
		if err != nil {
			return err
		}
		return nil
	}

	spotifyErr := &SpotifyError{
		HTTPStatus: statusCode,
		Code:       -1,
		URL:        url,  // Structured field
		Method:     method, // HTTP method
		Message:    string(body), // Without URL prefix
		Headers:    headers,
	}

	// Try to parse JSON error response
	var errorResp ErrorResponse
	if jsonErr := json.Unmarshal(body, &errorResp); jsonErr == nil {
		spotifyErr.Code = errorResp.Error.Status
		if errorResp.Error.Message != "" {
			spotifyErr.Message = errorResp.Error.Message // Clean message
		}
		if errorResp.Error.Reason != "" {
			spotifyErr.Reason = errorResp.Error.Reason
		}
	} else if len(body) > 0 {
		spotifyErr.Message = string(body)
	}

	// If there's an underlying error, wrap it
	if err != nil {
		return fmt.Errorf("%w: %v", err, spotifyErr)
	}

	return spotifyErr
}

// WrapRetryError wraps errors that occur during retry attempts
func WrapRetryError(err error, url string, reason string) error {
	if err == nil {
		return nil
	}

	spotifyErr := &SpotifyError{
		HTTPStatus: 429, // Rate limit
		Code:       -1,
		URL:        url,
		Method:     "", // Not available in retry context
		Message:    "Max Retries",
		Reason:     reason,
	}

	// Wrap the underlying error
	return fmt.Errorf("%w: %v", err, spotifyErr)
}

// WrapJSONError wraps JSON decode errors with context
func WrapJSONError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to parse JSON response: %w", err)
}
