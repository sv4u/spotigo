// Package spotigo provides a Go client library for the Spotify Web API.
//
// This library provides full access to Spotify's music catalog, user data,
// and playback controls with type safety and idiomatic Go patterns.
//
// Quick Start:
//
//	auth := spotigo.NewClientCredentials("client_id", "client_secret")
//	client, err := spotigo.NewClient(auth)
//	if err != nil {
//		panic(err)
//	}
//
//	track, err := client.Track(context.Background(), "track_id")
//
// For more information, see:
//   - Spotify Web API: https://developer.spotify.com/documentation/web-api
//   - Examples: https://github.com/spotipy-dev/spotipy-examples
package spotigo

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultAPIPrefix is the default Spotify Web API base URL
	DefaultAPIPrefix = "https://api.spotify.com/v1/"
	// DefaultTimeout is the default request timeout
	DefaultTimeout = 5 * time.Second
	// DefaultMaxRetries is the default maximum number of retries
	DefaultMaxRetries = 3
)

// Logger defines a simple logging interface for the client.
// Implement this interface to provide custom logging behavior.
type Logger interface {
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

// DefaultLogger wraps Go's standard log package.
// Provides basic logging functionality with INFO, WARN, and ERROR levels.
// Debug logging is disabled by default.
type DefaultLogger struct{}

func (l *DefaultLogger) Debug(format string, v ...interface{}) {
	// Debug logging disabled by default
}

func (l *DefaultLogger) Info(format string, v ...interface{}) {
	log.Printf("[INFO] "+format, v...)
}

func (l *DefaultLogger) Warn(format string, v ...interface{}) {
	log.Printf("[WARN] "+format, v...)
}

func (l *DefaultLogger) Error(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries       int
	StatusRetries    int
	StatusForcelist  []int
	BackoffFactor    float64
	RetryAfterHeader bool
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:       DefaultMaxRetries,
		StatusRetries:    3,
		StatusForcelist:  []int{429, 500, 502, 503, 504},
		BackoffFactor:    0.3,
		RetryAfterHeader: true,
	}
}

// Client is the main Spotify API client.
//
// It provides methods for accessing all Spotify Web API endpoints.
// All methods accept context.Context for cancellation and timeouts.
//
// Example:
//
//	auth := spotigo.NewClientCredentials("client_id", "client_secret")
//	client, err := spotigo.NewClient(auth)
//	if err != nil {
//		panic(err)
//	}
//
//	track, err := client.Track(ctx, "4iV5W9uYEdYUVa79Axb7Rh")
type Client struct {
	HTTPClient     *http.Client      // Custom HTTP client (optional)
	AuthManager    AuthManager       // Authentication manager (required)
	CacheHandler   CacheHandler      // Token cache handler (optional)
	APIPrefix      string            // API base URL (default: https://api.spotify.com/v1/)
	Language       string            // Language for localized responses
	RetryConfig    *RetryConfig      // Retry configuration
	RequestTimeout time.Duration     // Request timeout
	Logger         Logger            // Logger for debugging
	Proxies        map[string]string // HTTP proxies
	MaxRetries     int               // Maximum retry attempts
	CountryCodes   []string          // Supported country codes (ISO 3166-1 alpha-2)
}

// ClientOption is a functional option for client configuration.
// Use With* functions to configure the client.
type ClientOption func(*Client)

// NewClient creates a new Spotify API client.
//
// The authManager parameter is required and provides authentication for API requests.
// Use ClientOption functions to customize the client behavior.
//
// Example:
//
//	auth := spotigo.NewClientCredentials("client_id", "client_secret")
//	client, err := spotigo.NewClient(auth,
//		spotigo.WithCacheHandler(cache),
//		spotigo.WithLanguage("en"),
//	)
func NewClient(authManager AuthManager, opts ...ClientOption) (*Client, error) {
	if authManager == nil {
		return nil, fmt.Errorf("auth manager is required")
	}

	client := &Client{
		AuthManager:    authManager,
		APIPrefix:      DefaultAPIPrefix,
		RetryConfig:    DefaultRetryConfig(),
		RequestTimeout: DefaultTimeout,
		MaxRetries:     DefaultMaxRetries,
		Logger:         &DefaultLogger{},
		CountryCodes:   getDefaultCountryCodes(),
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	// Initialize HTTP client if not provided
	if client.HTTPClient == nil {
		client.HTTPClient = &http.Client{
			Timeout: client.RequestTimeout,
		}
	}

	return client, nil
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// WithCacheHandler sets a cache handler
func WithCacheHandler(handler CacheHandler) ClientOption {
	return func(c *Client) {
		c.CacheHandler = handler
	}
}

// WithLanguage sets the Accept-Language header
func WithLanguage(lang string) ClientOption {
	return func(c *Client) {
		c.Language = lang
	}
}

// WithRetryConfig sets the retry configuration
func WithRetryConfig(config *RetryConfig) ClientOption {
	return func(c *Client) {
		c.RetryConfig = config
	}
}

// WithRequestTimeout sets the request timeout
func WithRequestTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.RequestTimeout = timeout
		if c.HTTPClient != nil {
			c.HTTPClient.Timeout = timeout
		}
	}
}

// WithLogger sets a custom logger
func WithLogger(logger Logger) ClientOption {
	return func(c *Client) {
		c.Logger = logger
	}
}

// WithAPIPrefix sets a custom API prefix
func WithAPIPrefix(prefix string) ClientOption {
	return func(c *Client) {
		c.APIPrefix = prefix
	}
}

// getDefaultCountryCodes returns the list of supported country codes
// Uses the shared SupportedCountryCodes map from util.go
func getDefaultCountryCodes() []string {
	codes := make([]string, 0, len(SupportedCountryCodes))
	for code := range SupportedCountryCodes {
		codes = append(codes, code)
	}
	return codes
}

// validatePaginationParams validates limit and offset parameters
func validatePaginationParams(limit, offset int) error {
	if limit < 0 {
		return fmt.Errorf("limit must be non-negative, got %d", limit)
	}
	if offset < 0 {
		return fmt.Errorf("offset must be non-negative, got %d", offset)
	}
	return nil
}

// validateMarketParameter validates market parameter (country code or "from_token")
func validateMarketParameter(market string) error {
	if market == "" {
		return nil // Empty is valid (will use default)
	}
	// "from_token" is a special value allowed by Spotify API
	if market == "from_token" {
		return nil
	}
	if !ValidateCountryCode(market) {
		return fmt.Errorf("invalid country code: %s", market)
	}
	return nil
}

// _internal_call performs the core HTTP request with retry logic
func (c *Client) _internal_call(
	ctx context.Context,
	method string,
	urlStr string,
	params url.Values,
	body interface{},
	result interface{},
) error {
	// Build full URL
	fullURL := c.buildURL(urlStr, params)

	// Retry loop
	var lastErr error
	for attempt := 0; attempt <= c.RetryConfig.MaxRetries; attempt++ {
		// Check context cancellation before retry attempt
		select {
		case <-ctx.Done():
			return fmt.Errorf("request cancelled after %d retry attempts: %w", attempt, ctx.Err())
		default:
		}

		// Refresh token before each attempt to ensure we have a valid token
		// This is especially important during long retry delays (e.g., 429 Retry-After)
		token, err := c.AuthManager.GetAccessToken(ctx)
		if err != nil {
			return fmt.Errorf("failed to get access token: %w", err)
		}

		// Create request with fresh token
		req, err := c.createRequest(ctx, method, fullURL, body, token, params)
		if err != nil {
			return err
		}

		// Log request
		c.logRequest(req, body)

		// Execute request
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			if !c.shouldRetry(err, attempt) {
				return fmt.Errorf("request failed: %w", err)
			}
			// Calculate backoff and retry
			delay := c.calculateBackoffDelay(attempt)
			c.logRetry(attempt, delay, err)
			
			// Check context cancellation before sleeping
			select {
			case <-ctx.Done():
				return fmt.Errorf("request cancelled after %d retry attempts: %w", attempt, ctx.Err())
			case <-time.After(delay):
				// Continue retry
			}
			continue
		}

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			if !c.shouldRetry(err, attempt) {
				return lastErr
			}
			continue
		}

		// Check for errors
		if resp.StatusCode >= 400 {
			spotifyErr := c.parseErrorResponse(resp.StatusCode, method, resp.Header, respBody, fullURL)

			// Check if retryable
			if c.shouldRetryStatus(resp.StatusCode, attempt) {
				delay := c.calculateRetryDelay(resp.StatusCode, resp.Header, attempt)
				c.logRetry(attempt, delay, spotifyErr)
				
				// Check context cancellation before sleeping
				select {
				case <-ctx.Done():
					return fmt.Errorf("request cancelled after %d retry attempts: %w", attempt, ctx.Err())
				case <-time.After(delay):
					// Continue retry
				}
				lastErr = spotifyErr
				continue
			}

			return spotifyErr
		}

		// Decode successful response
		if result != nil {
			if len(respBody) == 0 {
				// Empty response - valid for 204 No Content
				if resp.StatusCode == 204 {
					return nil
				}
				// For other status codes, result may have zero values
				// Continue to unmarshal (will result in zero values)
			}
			if err := json.Unmarshal(respBody, result); err != nil {
				return WrapJSONError(err)
			}
		}

		// Log success
		c.logResponse(resp.StatusCode, respBody)

		return nil
	}

	// Max retries exceeded
	return WrapRetryError(lastErr, fullURL, "Max retries exceeded")
}

// buildURL constructs the full URL from base URL and parameters
func (c *Client) buildURL(urlStr string, params url.Values) string {
	// If URL is absolute, use as-is
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
		if len(params) > 0 {
			parsedURL, err := url.Parse(urlStr)
			if err != nil {
				// Log error but return original URL to avoid breaking callers
				// This should rarely happen with valid URLs
				log.Printf("Warning: Failed to parse URL %q: %v", urlStr, err)
				return urlStr
			}
			if parsedURL != nil {
				// Merge existing query params with new params
				existing := parsedURL.Query()
				for k, v := range params {
					existing[k] = v
				}
				parsedURL.RawQuery = existing.Encode()
				return parsedURL.String()
			}
		}
		return urlStr
	}

	// Relative URL - prepend API prefix
	fullURL := c.APIPrefix + strings.TrimPrefix(urlStr, "/")

	// Add query parameters
	if len(params) > 0 {
		if strings.Contains(fullURL, "?") {
			fullURL += "&" + params.Encode()
		} else {
			fullURL += "?" + params.Encode()
		}
	}

	return fullURL
}

// createRequest creates an HTTP request with proper headers and body
func (c *Client) createRequest(ctx context.Context, method, urlStr string, body interface{}, token string, params url.Values) (*http.Request, error) {
	var reqBody io.Reader
	contentType := "application/json"

	// Handle custom content type from params
	if params != nil {
		if customContentType := params.Get("content_type"); customContentType != "" {
			contentType = customContentType
			params.Del("content_type") // Remove from query params
		}
	}

	// Encode body
	if body != nil {
		if contentType == "application/json" {
			// JSON encode
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewReader(jsonData)
		} else {
			// Form-encoded or other content type
			if formData, ok := body.(url.Values); ok {
				reqBody = strings.NewReader(formData.Encode())
			} else if strData, ok := body.(string); ok {
				reqBody = strings.NewReader(strData)
			} else if byteData, ok := body.([]byte); ok {
				reqBody = bytes.NewReader(byteData)
			} else {
				return nil, fmt.Errorf("unsupported body type for content type %s", contentType)
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", contentType)
	}
	if c.Language != "" {
		req.Header.Set("Accept-Language", c.Language)
	}

	return req, nil
}

// shouldRetry determines if a network error should be retried
func (c *Client) shouldRetry(err error, attempt int) bool {
	if attempt >= c.RetryConfig.MaxRetries {
		return false
	}
	// Retry on network errors (context timeout, connection errors, etc.)
	return true
}

// shouldRetryStatus determines if an HTTP status code should be retried
func (c *Client) shouldRetryStatus(statusCode, attempt int) bool {
	if attempt >= c.RetryConfig.StatusRetries {
		return false
	}
	// Check if status code is in retry list
	for _, code := range c.RetryConfig.StatusForcelist {
		if statusCode == code {
			return true
		}
	}
	return false
}

// calculateBackoffDelay calculates exponential backoff delay
func (c *Client) calculateBackoffDelay(attempt int) time.Duration {
	delay := time.Duration(float64(attempt+1) * c.RetryConfig.BackoffFactor * float64(time.Second))
	if delay > 30*time.Second {
		delay = 30 * time.Second
	}
	return delay
}

// calculateRetryDelay calculates retry delay, using Retry-After header if available
func (c *Client) calculateRetryDelay(statusCode int, headers http.Header, attempt int) time.Duration {
	// For 429, try to use Retry-After header
	if statusCode == 429 && c.RetryConfig.RetryAfterHeader {
		if retryAfter := headers.Get("Retry-After"); retryAfter != "" {
			// Try parsing as integer seconds first
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				return time.Duration(seconds) * time.Second
			}
			// Try parsing as HTTP-date
			if t, err := http.ParseTime(retryAfter); err == nil {
				delay := time.Until(t)
				if delay > 0 {
					return delay
				}
			}
		}
	}

	// Use exponential backoff
	return c.calculateBackoffDelay(attempt)
}

// parseErrorResponse parses error response from Spotify API
func (c *Client) parseErrorResponse(statusCode int, method string, headers http.Header, body []byte, requestURL string) error {
	return WrapHTTPError(nil, statusCode, method, requestURL, body, headers)
}

// logRequest logs the request details
func (c *Client) logRequest(req *http.Request, body interface{}) {
	if c.Logger == nil {
		return
	}
	c.Logger.Debug("Request: %s %s", req.Method, req.URL.String())
	if body != nil {
		// Sanitize sensitive data if needed
		c.Logger.Debug("Request body: %v", body)
	}
}

// logResponse logs the response details
func (c *Client) logResponse(statusCode int, body []byte) {
	if c.Logger == nil {
		return
	}
	c.Logger.Debug("Response: %d", statusCode)
	if len(body) > 0 {
		c.Logger.Debug("Response body: %s", string(body))
	}
}

// logRetry logs retry attempts
func (c *Client) logRetry(attempt int, delay time.Duration, err error) {
	if c.Logger == nil {
		return
	}
	if spotifyErr, ok := err.(*SpotifyError); ok && spotifyErr.HTTPStatus == 429 {
		c.Logger.Warn("Your application has reached a rate/request limit. Retry will occur after: %.0f s", delay.Seconds())
	} else {
		c.Logger.Warn("Retry attempt %d after %v: %v", attempt+1, delay, err)
	}
}

// ============================================================================
// Category 1: Tracks & Artists
// ============================================================================

// Track retrieves a single track by ID, URI, or URL.
//
// The trackID parameter can be:
//   - A Spotify URI: "spotify:track:4iV5W9uYEdYUVa79Axb7Rh"
//   - A Spotify URL: "https://open.spotify.com/track/4iV5W9uYEdYUVa79Axb7Rh"
//   - A raw track ID: "4iV5W9uYEdYUVa79Axb7Rh"
//
// The optional market parameter restricts results to a specific country.
// If not provided, results may vary by user's country.
//
// Example:
//
//	track, err := client.Track(ctx, "4iV5W9uYEdYUVa79Axb7Rh", "US")
//	if err != nil {
//		// Handle error
//	}
//	fmt.Println(track.Name)
//
// See also: Tracks for retrieving multiple tracks.
func (c *Client) Track(ctx context.Context, trackID string, market ...string) (*Track, error) {
	id, err := GetID(trackID, "track")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if len(market) > 0 && market[0] != "" {
		if err := validateMarketParameter(market[0]); err != nil {
			return nil, err
		}
		params.Set("market", market[0])
	}

	var result Track
	if err := c._get(ctx, fmt.Sprintf("tracks/%s", id), params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Tracks retrieves multiple tracks by IDs, URIs, or URLs
// Maximum 50 tracks per request
func (c *Client) Tracks(ctx context.Context, trackIDs []string, market ...string) (*TracksResponse, error) {
	if len(trackIDs) > 50 {
		return nil, fmt.Errorf("maximum 50 tracks per request")
	}

	ids := make([]string, len(trackIDs))
	for i, id := range trackIDs {
		extracted, err := GetID(id, "track")
		if err != nil {
			return nil, err
		}
		ids[i] = extracted
	}

	params := url.Values{}
	params.Set("ids", strings.Join(ids, ","))
	if len(market) > 0 && market[0] != "" {
		if err := validateMarketParameter(market[0]); err != nil {
			return nil, err
		}
		params.Set("market", market[0])
	}

	var result TracksResponse
	if err := c._get(ctx, "tracks", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Artist retrieves a single artist by ID, URI, or URL
func (c *Client) Artist(ctx context.Context, artistID string) (*Artist, error) {
	id, err := GetID(artistID, "artist")
	if err != nil {
		return nil, err
	}

	var result Artist
	if err := c._get(ctx, fmt.Sprintf("artists/%s", id), nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Artists retrieves multiple artists by IDs, URIs, or URLs
func (c *Client) Artists(ctx context.Context, artistIDs []string) (*ArtistsResponse, error) {
	if len(artistIDs) > 50 {
		return nil, fmt.Errorf("maximum 50 artists per request")
	}

	ids := make([]string, len(artistIDs))
	for i, id := range artistIDs {
		extracted, err := GetID(id, "artist")
		if err != nil {
			return nil, err
		}
		ids[i] = extracted
	}

	params := url.Values{}
	params.Set("ids", strings.Join(ids, ","))

	var result ArtistsResponse
	if err := c._get(ctx, "artists", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ArtistAlbumsOptions holds options for ArtistAlbums
type ArtistAlbumsOptions struct {
	IncludeGroups []string // album, single, appears_on, compilation
	Country       string   // ISO 3166-1 alpha-2 country code
	Limit         int      // Default: 20, Max: 50
	Offset        int      // Default: 0
}

// ArtistAlbums retrieves albums by an artist
func (c *Client) ArtistAlbums(ctx context.Context, artistID string, opts *ArtistAlbumsOptions) (*Paging[SimplifiedAlbum], error) {
	id, err := GetID(artistID, "artist")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if len(opts.IncludeGroups) > 0 {
			params.Set("include_groups", strings.Join(opts.IncludeGroups, ","))
		}
		if opts.Country != "" {
			params.Set("country", opts.Country)
		}
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result Paging[SimplifiedAlbum]
	if err := c._get(ctx, fmt.Sprintf("artists/%s/albums", id), params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ArtistTopTracks retrieves top tracks for an artist by country
func (c *Client) ArtistTopTracks(ctx context.Context, artistID, market string) (*TracksResponse, error) {
	id, err := GetID(artistID, "artist")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if market != "" {
		if err := validateMarketParameter(market); err != nil {
			return nil, err
		}
		params.Set("market", market)
	} else {
		params.Set("market", "US") // Default
	}

	var result TracksResponse
	if err := c._get(ctx, fmt.Sprintf("artists/%s/top-tracks", id), params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ArtistRelatedArtists retrieves artists related to an artist
// Note: This endpoint may be deprecated by Spotify
func (c *Client) ArtistRelatedArtists(ctx context.Context, artistID string) (*ArtistsResponse, error) {
	id, err := GetID(artistID, "artist")
	if err != nil {
		return nil, err
	}

	var result ArtistsResponse
	if err := c._get(ctx, fmt.Sprintf("artists/%s/related-artists", id), nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ============================================================================
// Category 2: Albums
// ============================================================================

// Album retrieves a single album by ID, URI, or URL
func (c *Client) Album(ctx context.Context, albumID string, market ...string) (*Album, error) {
	id, err := GetID(albumID, "album")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if len(market) > 0 && market[0] != "" {
		if err := validateMarketParameter(market[0]); err != nil {
			return nil, err
		}
		params.Set("market", market[0])
	}

	var result Album
	if err := c._get(ctx, fmt.Sprintf("albums/%s", id), params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Albums retrieves multiple albums by IDs, URIs, or URLs
// Maximum 20 albums per request
func (c *Client) Albums(ctx context.Context, albumIDs []string, market ...string) (*AlbumsResponse, error) {
	if len(albumIDs) > 20 {
		return nil, fmt.Errorf("maximum 20 albums per request")
	}

	ids := make([]string, len(albumIDs))
	for i, id := range albumIDs {
		extracted, err := GetID(id, "album")
		if err != nil {
			return nil, err
		}
		ids[i] = extracted
	}

	params := url.Values{}
	params.Set("ids", strings.Join(ids, ","))
	if len(market) > 0 && market[0] != "" {
		if err := validateMarketParameter(market[0]); err != nil {
			return nil, err
		}
		params.Set("market", market[0])
	}

	var result AlbumsResponse
	if err := c._get(ctx, "albums", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// AlbumTracksOptions holds options for AlbumTracks
type AlbumTracksOptions struct {
	Market string // ISO 3166-1 alpha-2 country code
	Limit  int    // Default: 20, Max: 50
	Offset int    // Default: 0
}

// AlbumTracks retrieves tracks from an album
func (c *Client) AlbumTracks(ctx context.Context, albumID string, opts *AlbumTracksOptions) (*Paging[SimplifiedTrack], error) {
	id, err := GetID(albumID, "album")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.Market != "" {
			if err := validateMarketParameter(opts.Market); err != nil {
				return nil, err
			}
			params.Set("market", opts.Market)
		}
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result Paging[SimplifiedTrack]
	if err := c._get(ctx, fmt.Sprintf("albums/%s/tracks", id), params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ============================================================================
// Category 5: Search
// ============================================================================

// SearchOptions holds options for search
type SearchOptions struct {
	Market          string // ISO 3166-1 alpha-2 country code or "from_token"
	Limit           int    // Default: 10, Min: 1, Max: 50
	Offset          int    // Default: 0
	IncludeExternal string // "audio" to include external audio content
}

// Search performs a search query
// searchType: comma-separated types: 'artist', 'album', 'track', 'playlist', 'show', 'episode', 'audiobook'
func (c *Client) Search(ctx context.Context, query, searchType string, opts *SearchOptions) (*SearchResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if searchType == "" {
		searchType = "track" // Default
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("type", searchType)

	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.Market != "" {
			if err := validateMarketParameter(opts.Market); err != nil {
				return nil, err
			}
			params.Set("market", opts.Market)
		}
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			if opts.Limit < 1 {
				opts.Limit = 1
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "10") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
		if opts.IncludeExternal != "" {
			params.Set("include_external", opts.IncludeExternal)
		}
	} else {
		params.Set("limit", "10") // Default
	}

	var result SearchResponse
	if err := c._get(ctx, "search", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ============================================================================
// Category 6: Playlists
// ============================================================================

// PlaylistOptions holds options for playlist retrieval
type PlaylistOptions struct {
	Fields          string // Comma-separated field list
	AdditionalTypes string // Comma-separated types: track, episode
	Market          string // ISO 3166-1 alpha-2 country code
}

// Playlist retrieves a playlist by ID
func (c *Client) Playlist(ctx context.Context, playlistID string, opts *PlaylistOptions) (*Playlist, error) {
	id, err := GetID(playlistID, "playlist")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if opts != nil {
		if opts.Fields != "" {
			params.Set("fields", opts.Fields)
		}
		if opts.AdditionalTypes != "" {
			params.Set("additional_types", opts.AdditionalTypes)
		}
		if opts.Market != "" {
			if err := validateMarketParameter(opts.Market); err != nil {
				return nil, err
			}
			params.Set("market", opts.Market)
		}
	}

	var result Playlist
	if err := c._get(ctx, fmt.Sprintf("playlists/%s", id), params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// PlaylistTracksOptions holds options for playlist tracks
type PlaylistTracksOptions struct {
	Fields          string // Comma-separated field list
	Limit           int    // Default: 100, Max: 100
	Offset          int    // Default: 0
	Market          string // ISO 3166-1 alpha-2 country code
	AdditionalTypes string // Comma-separated types: track, episode
}

// PlaylistTracks retrieves tracks from a playlist
func (c *Client) PlaylistTracks(ctx context.Context, playlistID string, opts *PlaylistTracksOptions) (*Paging[PlaylistTrack], error) {
	id, err := GetID(playlistID, "playlist")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.Fields != "" {
			params.Set("fields", opts.Fields)
		}
		if opts.Limit > 0 {
			if opts.Limit > 100 {
				opts.Limit = 100
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "100") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
		if opts.Market != "" {
			if err := validateMarketParameter(opts.Market); err != nil {
				return nil, err
			}
			params.Set("market", opts.Market)
		}
		if opts.AdditionalTypes != "" {
			params.Set("additional_types", opts.AdditionalTypes)
		}
	} else {
		params.Set("limit", "100") // Default
	}

	var result Paging[PlaylistTrack]
	if err := c._get(ctx, fmt.Sprintf("playlists/%s/tracks", id), params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CurrentUserPlaylistsOptions holds options for current user playlists
type CurrentUserPlaylistsOptions struct {
	Limit  int // Default: 20, Max: 50
	Offset int // Default: 0
}

// CurrentUserPlaylists retrieves playlists for the current user
func (c *Client) CurrentUserPlaylists(ctx context.Context, opts *CurrentUserPlaylistsOptions) (*Paging[SimplifiedPlaylist], error) {
	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result Paging[SimplifiedPlaylist]
	if err := c._get(ctx, "me/playlists", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UserPlaylistCreate creates a new playlist for a user
func (c *Client) UserPlaylistCreate(ctx context.Context, userID string, opts *CreatePlaylistOptions) (*Playlist, error) {
	if opts == nil {
		return nil, fmt.Errorf("options are required")
	}

	if opts.Name == "" {
		return nil, fmt.Errorf("playlist name is required")
	}

	var result Playlist
	if err := c._post(ctx, fmt.Sprintf("users/%s/playlists", userID), nil, opts, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// PlaylistAddItemsRequest represents the request body for adding items to a playlist
// Format matches Spotify Web API specification: {"uris": [...], "position": int}
// uris: Array of Spotify URIs (spotify:track:... or spotify:episode:...)
// position: Optional position to insert items (0-based index)
type PlaylistAddItemsRequest struct {
	URIs     []string `json:"uris"`
	Position *int     `json:"position,omitempty"`
}

// PlaylistAddItems adds items (tracks or episodes) to a playlist
// items: list of track/episode URIs, URLs, or IDs (empty array is accepted by API but will have no effect)
// position: optional position to insert items (0-based)
func (c *Client) PlaylistAddItems(ctx context.Context, playlistID string, items []string, position ...int) (*PlaylistSnapshotID, error) {
	id, err := GetID(playlistID, "playlist")
	if err != nil {
		return nil, err
	}

	if len(items) > 100 {
		return nil, fmt.Errorf("maximum 100 items per request")
	}

	// Convert items to URIs, collecting invalid items
	uris := make([]string, 0, len(items))
	invalidItems := make([]string, 0)

	for _, item := range items {
		var uri string
		var err error

		// Check if already a URI
		if IsURI(item) {
			uris = append(uris, item)
			continue
		}

		// Check if it's a URL - extract ID and determine type
		if strings.Contains(item, "spotify.com") {
			// Extract type from URL path
			if strings.Contains(item, "/episode/") {
				id, err := GetID(item, "episode")
				if err == nil {
					uri, err = GetURI(id, "episode")
					if err == nil {
						uris = append(uris, uri)
						continue
					}
				}
			} else {
				// Default to track for URLs
				id, err := GetID(item, "track")
				if err == nil {
					uri, err = GetURI(id, "track")
					if err == nil {
						uris = append(uris, uri)
						continue
					}
				}
			}
		}

		// Raw ID - try track first (most common), then episode
		uri, err = GetURI(item, "track")
		if err != nil {
			uri, err = GetURI(item, "episode")
			if err != nil {
				// Invalid item - collect for error reporting
				invalidItems = append(invalidItems, item)
				continue
			}
		}
		uris = append(uris, uri)
	}

	// If all items are invalid, return error
	if len(invalidItems) > 0 && len(uris) == 0 {
		return nil, fmt.Errorf("all items invalid: %v", invalidItems)
	}

	reqBody := PlaylistAddItemsRequest{
		URIs: uris,
	}
	if len(position) > 0 {
		pos := position[0]
		if pos < 0 {
			return nil, fmt.Errorf("position must be non-negative, got %d", pos)
		}
		reqBody.Position = &pos
	}

	var result PlaylistSnapshotID
	if err := c._post(ctx, fmt.Sprintf("playlists/%s/tracks", id), nil, reqBody, &result); err != nil {
		return nil, err
	}

	// If there were invalid items but we processed valid ones, return result with error
	if len(invalidItems) > 0 {
		return &result, fmt.Errorf("some items could not be converted to URIs (processed %d valid items): %v", len(uris), invalidItems)
	}

	return &result, nil
}

// PlaylistReplaceItems replaces all items in a playlist
// items: list of track/episode URIs, URLs, or IDs
func (c *Client) PlaylistReplaceItems(ctx context.Context, playlistID string, items []string) (*PlaylistSnapshotID, error) {
	id, err := GetID(playlistID, "playlist")
	if err != nil {
		return nil, err
	}

	// Convert items to URIs (similar to PlaylistAddItems)
	uris := make([]string, 0, len(items))
	for _, item := range items {
		var uri string
		var convErr error

		// Check if already a URI
		if IsURI(item) {
			uris = append(uris, item)
			continue
		}

		// Check if it's a URL - extract ID and determine type
		if strings.Contains(item, "spotify.com") {
			if strings.Contains(item, "/episode/") {
				extractedID, err := GetID(item, "episode")
				if err == nil {
					uri, convErr = GetURI(extractedID, "episode")
					if convErr == nil {
						uris = append(uris, uri)
						continue
					}
				}
			} else {
				// Default to track for URLs
				extractedID, err := GetID(item, "track")
				if err == nil {
					uri, convErr = GetURI(extractedID, "track")
					if convErr == nil {
						uris = append(uris, uri)
						continue
					}
				}
			}
		}

		// Raw ID - try track first, then episode
		uri, convErr = GetURI(item, "track")
		if convErr != nil {
			uri, convErr = GetURI(item, "episode")
			if convErr != nil {
				return nil, fmt.Errorf("failed to convert item to URI: %w", convErr)
			}
		}
		uris = append(uris, uri)
	}

	reqBody := map[string]interface{}{
		"uris": uris,
	}

	var result PlaylistSnapshotID
	if err := c._put(ctx, fmt.Sprintf("playlists/%s/tracks", id), nil, reqBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// PlaylistReorderItems reorders items in a playlist
func (c *Client) PlaylistReorderItems(ctx context.Context, playlistID string, opts *ReorderItemsOptions) (*PlaylistSnapshotID, error) {
	id, err := GetID(playlistID, "playlist")
	if err != nil {
		return nil, err
	}

	if opts == nil {
		return nil, fmt.Errorf("options are required")
	}

	var result PlaylistSnapshotID
	if err := c._put(ctx, fmt.Sprintf("playlists/%s/tracks", id), nil, opts, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// PlaylistRemoveItemsRequest represents the request body for removing items from a playlist
// Format matches Spotify Web API specification: {"tracks": [...], "snapshot_id": "..."}
// tracks: Array of PlaylistItemToRemove objects
// snapshot_id: Optional playlist snapshot ID for optimistic concurrency
type PlaylistRemoveItemsRequest struct {
	Tracks     []PlaylistItemToRemove `json:"tracks"`
	SnapshotID *string                `json:"snapshot_id,omitempty"`
}

// PlaylistRemoveItems removes items from a playlist
// Note: Spotify DELETE with body requires special handling
func (c *Client) PlaylistRemoveItems(ctx context.Context, playlistID string, items []PlaylistItemToRemove, snapshotID ...string) (*PlaylistSnapshotID, error) {
	id, err := GetID(playlistID, "playlist")
	if err != nil {
		return nil, err
	}

	// Validate items
	for i, item := range items {
		if item.URI == "" {
			return nil, fmt.Errorf("item at index %d has empty URI", i)
		}
	}

	reqBody := PlaylistRemoveItemsRequest{
		Tracks: items,
	}
	if len(snapshotID) > 0 && snapshotID[0] != "" {
		reqBody.SnapshotID = &snapshotID[0]
	}

	// DELETE with body - use _internal_call directly
	var result PlaylistSnapshotID
	if err := c._internal_call(ctx, "DELETE", fmt.Sprintf("playlists/%s/tracks", id), nil, reqBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ChangePlaylistDetailsOptions holds options for changing playlist details
type ChangePlaylistDetailsOptions struct {
	Name          *string `json:"name,omitempty"`
	Public        *bool   `json:"public,omitempty"`
	Collaborative *bool   `json:"collaborative,omitempty"`
	Description   *string `json:"description,omitempty"`
}

// ReorderItemsOptions holds options for reordering playlist items
type ReorderItemsOptions struct {
	RangeStart  int     `json:"range_start"`
	InsertBefore int    `json:"insert_before"`
	RangeLength *int    `json:"range_length,omitempty"`
	SnapshotID  *string `json:"snapshot_id,omitempty"`
}

// CreatePlaylistOptions holds options for creating a playlist
type CreatePlaylistOptions struct {
	Name          string `json:"name"`
	Public        *bool  `json:"public,omitempty"`
	Collaborative *bool  `json:"collaborative,omitempty"`
	Description   string `json:"description,omitempty"`
}

// PlaylistChangeDetails changes playlist details
func (c *Client) PlaylistChangeDetails(ctx context.Context, playlistID string, opts *ChangePlaylistDetailsOptions) error {
	id, err := GetID(playlistID, "playlist")
	if err != nil {
		return err
	}

	if opts == nil {
		return fmt.Errorf("options are required")
	}

	if err := c._put(ctx, fmt.Sprintf("playlists/%s", id), nil, opts, nil); err != nil {
		return err
	}

	return nil
}

// PlaylistCoverImage retrieves the playlist cover image
func (c *Client) PlaylistCoverImage(ctx context.Context, playlistID string) ([]Image, error) {
	id, err := GetID(playlistID, "playlist")
	if err != nil {
		return nil, err
	}

	var result []Image
	if err := c._get(ctx, fmt.Sprintf("playlists/%s/images", id), nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// PlaylistUploadCoverImage uploads a custom cover image for a playlist
// imageData: JPEG image data (base64 encoded, max 256KB)
func (c *Client) PlaylistUploadCoverImage(ctx context.Context, playlistID string, imageData []byte) error {
	id, err := GetID(playlistID, "playlist")
	if err != nil {
		return err
	}

	// Validate image size (max 256KB)
	const maxImageSize = 256 * 1024
	if len(imageData) > maxImageSize {
		return fmt.Errorf("image size exceeds maximum of 256KB: %d bytes", len(imageData))
	}

	// Validate that it's a JPEG (check magic bytes)
	if len(imageData) < 2 || (imageData[0] != 0xFF || imageData[1] != 0xD8) {
		return fmt.Errorf("image must be in JPEG format")
	}

	// Base64 encode the image data
	encoded := base64.StdEncoding.EncodeToString(imageData)

	// Set content type for image/jpeg
	params := url.Values{}
	params.Set("content_type", "image/jpeg")

	// Send as string body (base64 encoded)
	if err := c._put(ctx, fmt.Sprintf("playlists/%s/images", id), params, encoded, nil); err != nil {
		return err
	}

	return nil
}

// ============================================================================
// Category 7: User Profile
// ============================================================================

// CurrentUser retrieves the current user's profile
func (c *Client) CurrentUser(ctx context.Context) (*User, error) {
	var result User
	if err := c._get(ctx, "me", nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// User retrieves a user's profile by ID
func (c *Client) User(ctx context.Context, userID string) (*PublicUser, error) {
	var result PublicUser
	if err := c._get(ctx, fmt.Sprintf("users/%s", userID), nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ============================================================================
// Category 8: Saved Content
// ============================================================================

// SavedTracksOptions holds options for saved tracks
type SavedTracksOptions struct {
	Market string // ISO 3166-1 alpha-2 country code
	Limit  int    // Default: 20, Max: 50
	Offset int    // Default: 0
}

// CurrentUserSavedTracks retrieves user's saved tracks
func (c *Client) CurrentUserSavedTracks(ctx context.Context, opts *SavedTracksOptions) (*Paging[SavedTrack], error) {
	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.Market != "" {
			if err := validateMarketParameter(opts.Market); err != nil {
				return nil, err
			}
			params.Set("market", opts.Market)
		}
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result Paging[SavedTrack]
	if err := c._get(ctx, "me/tracks", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CurrentUserSavedTracksAdd adds tracks to user's library
// CurrentUserSavedTracksAdd saves tracks for the current user
// trackIDs: list of track IDs, URIs, or URLs (empty array is accepted by API but will have no effect)
func (c *Client) CurrentUserSavedTracksAdd(ctx context.Context, trackIDs []string) error {
	if len(trackIDs) > 50 {
		return fmt.Errorf("maximum 50 tracks per request")
	}

	ids := make([]string, len(trackIDs))
	for i, id := range trackIDs {
		extracted, err := GetID(id, "track")
		if err != nil {
			return err
		}
		ids[i] = extracted
	}

	body := map[string]interface{}{
		"ids": ids,
	}

	return c._put(ctx, "me/tracks", nil, body, nil)
}

// CurrentUserSavedTracksDelete removes tracks from user's library
// CurrentUserSavedTracksDelete removes saved tracks for the current user
// trackIDs: list of track IDs, URIs, or URLs (empty array is accepted by API but will have no effect)
func (c *Client) CurrentUserSavedTracksDelete(ctx context.Context, trackIDs []string) error {
	if len(trackIDs) > 50 {
		return fmt.Errorf("maximum 50 tracks per request")
	}

	ids := make([]string, len(trackIDs))
	for i, id := range trackIDs {
		extracted, err := GetID(id, "track")
		if err != nil {
			return err
		}
		ids[i] = extracted
	}

	body := map[string]interface{}{
		"ids": ids,
	}

	return c._internal_call(ctx, "DELETE", "me/tracks", nil, body, nil)
}

// CurrentUserSavedTracksContains checks if tracks are saved
func (c *Client) CurrentUserSavedTracksContains(ctx context.Context, trackIDs []string) ([]bool, error) {
	if len(trackIDs) > 50 {
		return nil, fmt.Errorf("maximum 50 tracks per request")
	}

	ids := make([]string, len(trackIDs))
	for i, id := range trackIDs {
		extracted, err := GetID(id, "track")
		if err != nil {
			return nil, err
		}
		ids[i] = extracted
	}

	params := url.Values{}
	params.Set("ids", strings.Join(ids, ","))

	var result []bool
	if err := c._get(ctx, "me/tracks/contains", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// SavedAlbumsOptions holds options for saved albums
type SavedAlbumsOptions struct {
	Limit  int // Default: 20, Max: 50
	Offset int // Default: 0
}

// CurrentUserSavedAlbums retrieves user's saved albums
func (c *Client) CurrentUserSavedAlbums(ctx context.Context, opts *SavedAlbumsOptions) (*Paging[SavedAlbum], error) {
	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result Paging[SavedAlbum]
	if err := c._get(ctx, "me/albums", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CurrentUserSavedAlbumsAdd adds albums to user's library
func (c *Client) CurrentUserSavedAlbumsAdd(ctx context.Context, albumIDs []string) error {
	if len(albumIDs) > 50 {
		return fmt.Errorf("maximum 50 albums per request")
	}

	ids := make([]string, len(albumIDs))
	for i, id := range albumIDs {
		extracted, err := GetID(id, "album")
		if err != nil {
			return err
		}
		ids[i] = extracted
	}

	body := map[string]interface{}{
		"ids": ids,
	}

	return c._put(ctx, "me/albums", nil, body, nil)
}

// CurrentUserSavedAlbumsDelete removes albums from user's library
func (c *Client) CurrentUserSavedAlbumsDelete(ctx context.Context, albumIDs []string) error {
	if len(albumIDs) > 50 {
		return fmt.Errorf("maximum 50 albums per request")
	}

	ids := make([]string, len(albumIDs))
	for i, id := range albumIDs {
		extracted, err := GetID(id, "album")
		if err != nil {
			return err
		}
		ids[i] = extracted
	}

	body := map[string]interface{}{
		"ids": ids,
	}

	return c._internal_call(ctx, "DELETE", "me/albums", nil, body, nil)
}

// CurrentUserSavedAlbumsContains checks if albums are saved
func (c *Client) CurrentUserSavedAlbumsContains(ctx context.Context, albumIDs []string) ([]bool, error) {
	if len(albumIDs) > 50 {
		return nil, fmt.Errorf("maximum 50 albums per request")
	}

	ids := make([]string, len(albumIDs))
	for i, id := range albumIDs {
		extracted, err := GetID(id, "album")
		if err != nil {
			return nil, err
		}
		ids[i] = extracted
	}

	params := url.Values{}
	params.Set("ids", strings.Join(ids, ","))

	var result []bool
	if err := c._get(ctx, "me/albums/contains", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// SavedEpisodesOptions holds options for saved episodes
type SavedEpisodesOptions struct {
	Market string // ISO 3166-1 alpha-2 country code
	Limit  int    // Default: 20, Max: 50
	Offset int    // Default: 0
}

// CurrentUserSavedEpisodes retrieves user's saved episodes
func (c *Client) CurrentUserSavedEpisodes(ctx context.Context, opts *SavedEpisodesOptions) (*Paging[SavedEpisode], error) {
	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.Market != "" {
			if err := validateMarketParameter(opts.Market); err != nil {
				return nil, err
			}
			params.Set("market", opts.Market)
		}
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result Paging[SavedEpisode]
	if err := c._get(ctx, "me/episodes", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CurrentUserSavedEpisodesAdd adds episodes to user's library
func (c *Client) CurrentUserSavedEpisodesAdd(ctx context.Context, episodeIDs []string) error {
	if len(episodeIDs) > 50 {
		return fmt.Errorf("maximum 50 episodes per request")
	}

	ids := make([]string, len(episodeIDs))
	for i, id := range episodeIDs {
		extracted, err := GetID(id, "episode")
		if err != nil {
			return err
		}
		ids[i] = extracted
	}

	body := map[string]interface{}{
		"ids": ids,
	}

	return c._put(ctx, "me/episodes", nil, body, nil)
}

// CurrentUserSavedEpisodesDelete removes episodes from user's library
func (c *Client) CurrentUserSavedEpisodesDelete(ctx context.Context, episodeIDs []string) error {
	if len(episodeIDs) > 50 {
		return fmt.Errorf("maximum 50 episodes per request")
	}

	ids := make([]string, len(episodeIDs))
	for i, id := range episodeIDs {
		extracted, err := GetID(id, "episode")
		if err != nil {
			return err
		}
		ids[i] = extracted
	}

	body := map[string]interface{}{
		"ids": ids,
	}

	return c._internal_call(ctx, "DELETE", "me/episodes", nil, body, nil)
}

// CurrentUserSavedEpisodesContains checks if episodes are saved
func (c *Client) CurrentUserSavedEpisodesContains(ctx context.Context, episodeIDs []string) ([]bool, error) {
	if len(episodeIDs) > 50 {
		return nil, fmt.Errorf("maximum 50 episodes per request")
	}

	ids := make([]string, len(episodeIDs))
	for i, id := range episodeIDs {
		extracted, err := GetID(id, "episode")
		if err != nil {
			return nil, err
		}
		ids[i] = extracted
	}

	params := url.Values{}
	params.Set("ids", strings.Join(ids, ","))

	var result []bool
	if err := c._get(ctx, "me/episodes/contains", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// SavedShowsOptions holds options for saved shows
type SavedShowsOptions struct {
	Limit  int // Default: 20, Max: 50
	Offset int // Default: 0
}

// CurrentUserSavedShows retrieves user's saved shows
func (c *Client) CurrentUserSavedShows(ctx context.Context, opts *SavedShowsOptions) (*Paging[SavedShow], error) {
	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result Paging[SavedShow]
	if err := c._get(ctx, "me/shows", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CurrentUserSavedShowsAdd adds shows to user's library
func (c *Client) CurrentUserSavedShowsAdd(ctx context.Context, showIDs []string) error {
	if len(showIDs) > 50 {
		return fmt.Errorf("maximum 50 shows per request")
	}

	ids := make([]string, len(showIDs))
	for i, id := range showIDs {
		extracted, err := GetID(id, "show")
		if err != nil {
			return err
		}
		ids[i] = extracted
	}

	body := map[string]interface{}{
		"ids": ids,
	}

	return c._put(ctx, "me/shows", nil, body, nil)
}

// CurrentUserSavedShowsDelete removes shows from user's library
func (c *Client) CurrentUserSavedShowsDelete(ctx context.Context, showIDs []string) error {
	if len(showIDs) > 50 {
		return fmt.Errorf("maximum 50 shows per request")
	}

	ids := make([]string, len(showIDs))
	for i, id := range showIDs {
		extracted, err := GetID(id, "show")
		if err != nil {
			return err
		}
		ids[i] = extracted
	}

	body := map[string]interface{}{
		"ids": ids,
	}

	return c._internal_call(ctx, "DELETE", "me/shows", nil, body, nil)
}

// CurrentUserSavedShowsContains checks if shows are saved
func (c *Client) CurrentUserSavedShowsContains(ctx context.Context, showIDs []string) ([]bool, error) {
	if len(showIDs) > 50 {
		return nil, fmt.Errorf("maximum 50 shows per request")
	}

	ids := make([]string, len(showIDs))
	for i, id := range showIDs {
		extracted, err := GetID(id, "show")
		if err != nil {
			return nil, err
		}
		ids[i] = extracted
	}

	params := url.Values{}
	params.Set("ids", strings.Join(ids, ","))

	var result []bool
	if err := c._get(ctx, "me/shows/contains", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// ============================================================================
// Category 9: Following
// ============================================================================

// FollowedArtistsOptions holds options for followed artists
type FollowedArtistsOptions struct {
	Type  string // "artist" (required)
	After string // Cursor for pagination
	Limit int    // Default: 20, Max: 50
}

// CurrentUserFollowedArtists retrieves user's followed artists
func (c *Client) CurrentUserFollowedArtists(ctx context.Context, opts *FollowedArtistsOptions) (*CursorPaging[Artist], error) {
	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters (limit only, no offset for cursor-based pagination)
		if err := validatePaginationParams(opts.Limit, 0); err != nil {
			return nil, err
		}
		
		params.Set("type", opts.Type)
		if opts.After != "" {
			params.Set("after", opts.After)
		}
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
	} else {
		params.Set("type", "artist")
		params.Set("limit", "20") // Default
	}

	var result struct {
		Artists CursorPaging[Artist] `json:"artists"`
	}
	if err := c._get(ctx, "me/following", params, &result); err != nil {
		return nil, err
	}

	return &result.Artists, nil
}

// CurrentUserFollowingArtists checks if user follows artists
func (c *Client) CurrentUserFollowingArtists(ctx context.Context, artistIDs []string) ([]bool, error) {
	if len(artistIDs) > 50 {
		return nil, fmt.Errorf("maximum 50 artists per request")
	}

	ids := make([]string, len(artistIDs))
	for i, id := range artistIDs {
		extracted, err := GetID(id, "artist")
		if err != nil {
			return nil, err
		}
		ids[i] = extracted
	}

	params := url.Values{}
	params.Set("type", "artist")
	params.Set("ids", strings.Join(ids, ","))

	var result []bool
	if err := c._get(ctx, "me/following/contains", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// CurrentUserFollowingUsers checks if user follows users
func (c *Client) CurrentUserFollowingUsers(ctx context.Context, userIDs []string) ([]bool, error) {
	if len(userIDs) > 50 {
		return nil, fmt.Errorf("maximum 50 users per request")
	}

	params := url.Values{}
	params.Set("type", "user")
	params.Set("ids", strings.Join(userIDs, ","))

	var result []bool
	if err := c._get(ctx, "me/following/contains", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// UserFollowArtists follows artists
// UserFollowArtists follows one or more artists
// artistIDs: list of artist IDs, URIs, or URLs (empty array is accepted by API but will have no effect)
func (c *Client) UserFollowArtists(ctx context.Context, artistIDs []string) error {
	if len(artistIDs) > 50 {
		return fmt.Errorf("maximum 50 artists per request")
	}

	ids := make([]string, len(artistIDs))
	for i, id := range artistIDs {
		extracted, err := GetID(id, "artist")
		if err != nil {
			return err
		}
		ids[i] = extracted
	}

	body := map[string]interface{}{
		"ids":  ids,
		"type": "artist",
	}

	return c._put(ctx, "me/following", nil, body, nil)
}

// UserFollowUsers follows one or more users
// userIDs: list of user IDs, URIs, or URLs (empty array is accepted by API but will have no effect)
func (c *Client) UserFollowUsers(ctx context.Context, userIDs []string) error {
	if len(userIDs) > 50 {
		return fmt.Errorf("maximum 50 users per request")
	}

	body := map[string]interface{}{
		"ids":  userIDs,
		"type": "user",
	}

	return c._put(ctx, "me/following", nil, body, nil)
}

// UserUnfollowArtists unfollows artists
// UserUnfollowArtists unfollows one or more artists
// artistIDs: list of artist IDs, URIs, or URLs (empty array is accepted by API but will have no effect)
func (c *Client) UserUnfollowArtists(ctx context.Context, artistIDs []string) error {
	if len(artistIDs) > 50 {
		return fmt.Errorf("maximum 50 artists per request")
	}

	ids := make([]string, len(artistIDs))
	for i, id := range artistIDs {
		extracted, err := GetID(id, "artist")
		if err != nil {
			return err
		}
		ids[i] = extracted
	}

	body := map[string]interface{}{
		"ids":  ids,
		"type": "artist",
	}

	return c._internal_call(ctx, "DELETE", "me/following", nil, body, nil)
}

// UserUnfollowUsers unfollows users
// UserUnfollowUsers unfollows one or more users
// userIDs: list of user IDs, URIs, or URLs (empty array is accepted by API but will have no effect)
func (c *Client) UserUnfollowUsers(ctx context.Context, userIDs []string) error {
	if len(userIDs) > 50 {
		return fmt.Errorf("maximum 50 users per request")
	}

	body := map[string]interface{}{
		"ids":  userIDs,
		"type": "user",
	}

	return c._internal_call(ctx, "DELETE", "me/following", nil, body, nil)
}

// CurrentUserFollowPlaylist follows a playlist
func (c *Client) CurrentUserFollowPlaylist(ctx context.Context, playlistID string, public ...bool) error {
	id, err := GetID(playlistID, "playlist")
	if err != nil {
		return err
	}

	body := map[string]interface{}{}
	if len(public) > 0 {
		body["public"] = public[0]
	}

	return c._put(ctx, fmt.Sprintf("playlists/%s/followers", id), nil, body, nil)
}

// CurrentUserUnfollowPlaylist unfollows a playlist
func (c *Client) CurrentUserUnfollowPlaylist(ctx context.Context, playlistID string) error {
	id, err := GetID(playlistID, "playlist")
	if err != nil {
		return err
	}

	return c._internal_call(ctx, "DELETE", fmt.Sprintf("playlists/%s/followers", id), nil, nil, nil)
}

// PlaylistIsFollowing checks if users follow a playlist
func (c *Client) PlaylistIsFollowing(ctx context.Context, playlistID string, userIDs []string) ([]bool, error) {
	if len(userIDs) > 5 {
		return nil, fmt.Errorf("maximum 5 users per request")
	}

	id, err := GetID(playlistID, "playlist")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("ids", strings.Join(userIDs, ","))

	var result []bool
	if err := c._get(ctx, fmt.Sprintf("playlists/%s/followers/contains", id), params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// ============================================================================
// Category 10: User Data (Top Tracks/Artists, Recently Played)
// ============================================================================

// TopItemsOptions holds options for top items
type TopItemsOptions struct {
	TimeRange string // "short_term", "medium_term", "long_term" (default: "medium_term")
	Limit     int    // Default: 20, Max: 50
	Offset    int    // Default: 0
}

// CurrentUserTopTracks retrieves user's top tracks
func (c *Client) CurrentUserTopTracks(ctx context.Context, opts *TopItemsOptions) (*Paging[Track], error) {
	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.TimeRange != "" {
			params.Set("time_range", opts.TimeRange)
		}
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result Paging[Track]
	if err := c._get(ctx, "me/top/tracks", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CurrentUserTopArtists retrieves user's top artists
func (c *Client) CurrentUserTopArtists(ctx context.Context, opts *TopItemsOptions) (*Paging[Artist], error) {
	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.TimeRange != "" {
			params.Set("time_range", opts.TimeRange)
		}
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result Paging[Artist]
	if err := c._get(ctx, "me/top/artists", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// RecentlyPlayedOptions holds options for recently played
type RecentlyPlayedOptions struct {
	Limit  int    // Default: 20, Max: 50
	After  *int64 // Unix timestamp in milliseconds
	Before *int64 // Unix timestamp in milliseconds
}

// CurrentUserRecentlyPlayed retrieves user's recently played tracks
func (c *Client) CurrentUserRecentlyPlayed(ctx context.Context, opts *RecentlyPlayedOptions) (*CursorPaging[PlayHistoryItem], error) {
	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters (limit only, no offset for cursor-based pagination)
		if err := validatePaginationParams(opts.Limit, 0); err != nil {
			return nil, err
		}
		
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.After != nil {
			params.Set("after", fmt.Sprintf("%d", *opts.After))
		}
		if opts.Before != nil {
			params.Set("before", fmt.Sprintf("%d", *opts.Before))
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result struct {
		Items   []PlayHistoryItem `json:"items"`
		Cursors *Cursors          `json:"cursors"`
		Next    *string           `json:"next"`
		Href    string            `json:"href"`
		Limit   int               `json:"limit"`
		Total   int               `json:"total"`
	}
	if err := c._get(ctx, "me/player/recently-played", params, &result); err != nil {
		return nil, err
	}

	// Convert to CursorPaging format
	cursorPaging := &CursorPaging[PlayHistoryItem]{
		Href:    result.Href,
		Items:   result.Items,
		Limit:   result.Limit,
		Total:   result.Total,
		Next:    result.Next,
		Cursors: result.Cursors,
	}

	return cursorPaging, nil
}

// ============================================================================
// Category 11: Browse (Categories, Featured Playlists, New Releases)
// ============================================================================

// BrowseCategoriesOptions holds options for browse categories
type BrowseCategoriesOptions struct {
	Country string // ISO 3166-1 alpha-2 country code
	Locale  string // ISO 639-1 language code and ISO 3166-1 alpha-2 country code
	Limit   int    // Default: 20, Max: 50
	Offset  int    // Default: 0
}

// BrowseCategories retrieves browse categories
func (c *Client) BrowseCategories(ctx context.Context, opts *BrowseCategoriesOptions) (*Paging[Category], error) {
	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.Country != "" {
			params.Set("country", opts.Country)
		}
		if opts.Locale != "" {
			params.Set("locale", opts.Locale)
		}
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result struct {
		Categories Paging[Category] `json:"categories"`
	}
	if err := c._get(ctx, "browse/categories", params, &result); err != nil {
		return nil, err
	}

	return &result.Categories, nil
}

// BrowseCategory retrieves a single category
func (c *Client) BrowseCategory(ctx context.Context, categoryID string, opts *BrowseCategoriesOptions) (*Category, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Country != "" {
			params.Set("country", opts.Country)
		}
		if opts.Locale != "" {
			params.Set("locale", opts.Locale)
		}
	}

	var result Category
	if err := c._get(ctx, fmt.Sprintf("browse/categories/%s", categoryID), params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// FeaturedPlaylistsOptions holds options for featured playlists
type FeaturedPlaylistsOptions struct {
	Country   string // ISO 3166-1 alpha-2 country code
	Locale    string // ISO 639-1 language code
	Limit     int    // Default: 20, Max: 50
	Offset    int    // Default: 0
	Timestamp string // ISO 8601 timestamp
}

// FeaturedPlaylistsResponse represents featured playlists response
type FeaturedPlaylistsResponse struct {
	Message   string                     `json:"message"`
	Playlists Paging[SimplifiedPlaylist] `json:"playlists"`
}

// BrowseFeaturedPlaylists retrieves featured playlists
func (c *Client) BrowseFeaturedPlaylists(ctx context.Context, opts *FeaturedPlaylistsOptions) (*FeaturedPlaylistsResponse, error) {
	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.Country != "" {
			params.Set("country", opts.Country)
		}
		if opts.Locale != "" {
			params.Set("locale", opts.Locale)
		}
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
		if opts.Timestamp != "" {
			params.Set("timestamp", opts.Timestamp)
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result FeaturedPlaylistsResponse
	if err := c._get(ctx, "browse/featured-playlists", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// NewReleasesOptions holds options for new releases
type NewReleasesOptions struct {
	Country string // ISO 3166-1 alpha-2 country code
	Limit   int    // Default: 20, Max: 50
	Offset  int    // Default: 0
}

// NewReleasesResponse represents new releases response
type NewReleasesResponse struct {
	Albums Paging[SimplifiedAlbum] `json:"albums"`
}

// BrowseNewReleases retrieves new releases
func (c *Client) BrowseNewReleases(ctx context.Context, opts *NewReleasesOptions) (*NewReleasesResponse, error) {
	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.Country != "" {
			params.Set("country", opts.Country)
		}
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result NewReleasesResponse
	if err := c._get(ctx, "browse/new-releases", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CategoryPlaylistsOptions holds options for category playlists
type CategoryPlaylistsOptions struct {
	Country string // ISO 3166-1 alpha-2 country code
	Limit   int    // Default: 20, Max: 50
	Offset  int    // Default: 0
}

// CategoryPlaylistsResponse represents category playlists response
type CategoryPlaylistsResponse struct {
	Playlists Paging[SimplifiedPlaylist] `json:"playlists"`
}

// BrowseCategoryPlaylists retrieves playlists for a category
func (c *Client) BrowseCategoryPlaylists(ctx context.Context, categoryID string, opts *CategoryPlaylistsOptions) (*CategoryPlaylistsResponse, error) {
	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.Country != "" {
			params.Set("country", opts.Country)
		}
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result CategoryPlaylistsResponse
	if err := c._get(ctx, fmt.Sprintf("browse/categories/%s/playlists", categoryID), params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ============================================================================
// Category 12: Recommendations
// ============================================================================

// RecommendationsOptions holds options for recommendations
type RecommendationsOptions struct {
	SeedArtists            []string // Up to 5 artist IDs
	SeedGenres             []string // Up to 5 genre names
	SeedTracks             []string // Up to 5 track IDs
	Limit                  int      // Default: 20, Max: 100
	Market                 string   // ISO 3166-1 alpha-2 country code
	MinAcousticness        *float64
	MaxAcousticness        *float64
	TargetAcousticness     *float64
	MinDanceability        *float64
	MaxDanceability        *float64
	TargetDanceability     *float64
	MinDurationMs          *int
	MaxDurationMs          *int
	TargetDurationMs       *int
	MinEnergy              *float64
	MaxEnergy              *float64
	TargetEnergy           *float64
	MinInstrumentalness    *float64
	MaxInstrumentalness    *float64
	TargetInstrumentalness *float64
	MinKey                 *int
	MaxKey                 *int
	TargetKey              *int
	MinLiveness            *float64
	MaxLiveness            *float64
	TargetLiveness         *float64
	MinLoudness            *float64
	MaxLoudness            *float64
	TargetLoudness         *float64
	MinMode                *int
	MaxMode                *int
	TargetMode             *int
	MinPopularity          *int
	MaxPopularity          *int
	TargetPopularity       *int
	MinSpeechiness         *float64
	MaxSpeechiness         *float64
	TargetSpeechiness      *float64
	MinTempo               *float64
	MaxTempo               *float64
	TargetTempo            *float64
	MinTimeSignature       *int
	MaxTimeSignature       *int
	TargetTimeSignature    *int
	MinValence             *float64
	MaxValence             *float64
	TargetValence          *float64
}

// Recommendations retrieves track recommendations
func (c *Client) Recommendations(ctx context.Context, opts *RecommendationsOptions) (*RecommendationsResponse, error) {
	if opts == nil {
		return nil, fmt.Errorf("options are required")
	}

	// Validate seed parameters (at least one required, max 5 total)
	totalSeeds := len(opts.SeedArtists) + len(opts.SeedGenres) + len(opts.SeedTracks)
	if totalSeeds == 0 {
		return nil, fmt.Errorf("at least one seed (artist, genre, or track) is required")
	}
	if totalSeeds > 5 {
		return nil, fmt.Errorf("maximum 5 seeds total (artists + genres + tracks)")
	}
	if len(opts.SeedArtists) > 5 {
		return nil, fmt.Errorf("maximum 5 seed artists")
	}
	if len(opts.SeedGenres) > 5 {
		return nil, fmt.Errorf("maximum 5 seed genres")
	}
	if len(opts.SeedTracks) > 5 {
		return nil, fmt.Errorf("maximum 5 seed tracks")
	}

	params := url.Values{}
	if len(opts.SeedArtists) > 0 {
		ids := make([]string, len(opts.SeedArtists))
		for i, id := range opts.SeedArtists {
			extracted, err := GetID(id, "artist")
			if err != nil {
				return nil, err
			}
			ids[i] = extracted
		}
		params.Set("seed_artists", strings.Join(ids, ","))
	}
	if len(opts.SeedGenres) > 0 {
		params.Set("seed_genres", strings.Join(opts.SeedGenres, ","))
	}
	if len(opts.SeedTracks) > 0 {
		ids := make([]string, len(opts.SeedTracks))
		for i, id := range opts.SeedTracks {
			extracted, err := GetID(id, "track")
			if err != nil {
				return nil, err
			}
			ids[i] = extracted
		}
		params.Set("seed_tracks", strings.Join(ids, ","))
	}

	if opts.Limit > 0 {
		if opts.Limit > 100 {
			opts.Limit = 100
		}
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	} else {
		params.Set("limit", "20") // Default
	}
	if opts.Market != "" {
		params.Set("market", opts.Market)
	}

	// Add audio feature parameters
	addFloatParam := func(key string, val *float64) {
		if val != nil {
			params.Set(key, fmt.Sprintf("%.2f", *val))
		}
	}
	addIntParam := func(key string, val *int) {
		if val != nil {
			params.Set(key, fmt.Sprintf("%d", *val))
		}
	}

	addFloatParam("min_acousticness", opts.MinAcousticness)
	addFloatParam("max_acousticness", opts.MaxAcousticness)
	addFloatParam("target_acousticness", opts.TargetAcousticness)
	addFloatParam("min_danceability", opts.MinDanceability)
	addFloatParam("max_danceability", opts.MaxDanceability)
	addFloatParam("target_danceability", opts.TargetDanceability)
	addIntParam("min_duration_ms", opts.MinDurationMs)
	addIntParam("max_duration_ms", opts.MaxDurationMs)
	addIntParam("target_duration_ms", opts.TargetDurationMs)
	addFloatParam("min_energy", opts.MinEnergy)
	addFloatParam("max_energy", opts.MaxEnergy)
	addFloatParam("target_energy", opts.TargetEnergy)
	addFloatParam("min_instrumentalness", opts.MinInstrumentalness)
	addFloatParam("max_instrumentalness", opts.MaxInstrumentalness)
	addFloatParam("target_instrumentalness", opts.TargetInstrumentalness)
	addIntParam("min_key", opts.MinKey)
	addIntParam("max_key", opts.MaxKey)
	addIntParam("target_key", opts.TargetKey)
	addFloatParam("min_liveness", opts.MinLiveness)
	addFloatParam("max_liveness", opts.MaxLiveness)
	addFloatParam("target_liveness", opts.TargetLiveness)
	addFloatParam("min_loudness", opts.MinLoudness)
	addFloatParam("max_loudness", opts.MaxLoudness)
	addFloatParam("target_loudness", opts.TargetLoudness)
	addIntParam("min_mode", opts.MinMode)
	addIntParam("max_mode", opts.MaxMode)
	addIntParam("target_mode", opts.TargetMode)
	addIntParam("min_popularity", opts.MinPopularity)
	addIntParam("max_popularity", opts.MaxPopularity)
	addIntParam("target_popularity", opts.TargetPopularity)
	addFloatParam("min_speechiness", opts.MinSpeechiness)
	addFloatParam("max_speechiness", opts.MaxSpeechiness)
	addFloatParam("target_speechiness", opts.TargetSpeechiness)
	addFloatParam("min_tempo", opts.MinTempo)
	addFloatParam("max_tempo", opts.MaxTempo)
	addFloatParam("target_tempo", opts.TargetTempo)
	addIntParam("min_time_signature", opts.MinTimeSignature)
	addIntParam("max_time_signature", opts.MaxTimeSignature)
	addIntParam("target_time_signature", opts.TargetTimeSignature)
	addFloatParam("min_valence", opts.MinValence)
	addFloatParam("max_valence", opts.MaxValence)
	addFloatParam("target_valence", opts.TargetValence)

	var result RecommendationsResponse
	if err := c._get(ctx, "recommendations", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// RecommendationGenreSeeds retrieves available genre seeds for recommendations
// Note: This endpoint may be deprecated by Spotify
func (c *Client) RecommendationGenreSeeds(ctx context.Context) ([]string, error) {
	var result struct {
		Genres []string `json:"genres"`
	}
	if err := c._get(ctx, "recommendations/available-genre-seeds", nil, &result); err != nil {
		return nil, err
	}

	return result.Genres, nil
}

// ============================================================================
// Category 13: Audio Features
// ============================================================================

// AudioFeatures retrieves audio features for a track
func (c *Client) AudioFeatures(ctx context.Context, trackID string) (*AudioFeatures, error) {
	id, err := GetID(trackID, "track")
	if err != nil {
		return nil, err
	}

	var result AudioFeatures
	if err := c._get(ctx, fmt.Sprintf("audio-features/%s", id), nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// AudioFeaturesMultiple retrieves audio features for multiple tracks
func (c *Client) AudioFeaturesMultiple(ctx context.Context, trackIDs []string) ([]AudioFeatures, error) {
	if len(trackIDs) > 100 {
		return nil, fmt.Errorf("maximum 100 tracks per request")
	}

	ids := make([]string, len(trackIDs))
	for i, id := range trackIDs {
		extracted, err := GetID(id, "track")
		if err != nil {
			return nil, err
		}
		ids[i] = extracted
	}

	params := url.Values{}
	params.Set("ids", strings.Join(ids, ","))

	var result struct {
		AudioFeatures []AudioFeatures `json:"audio_features"`
	}
	if err := c._get(ctx, "audio-features", params, &result); err != nil {
		return nil, err
	}

	return result.AudioFeatures, nil
}

// AudioAnalysis retrieves detailed audio analysis for a track
func (c *Client) AudioAnalysis(ctx context.Context, trackID string) (*AudioAnalysis, error) {
	id, err := GetID(trackID, "track")
	if err != nil {
		return nil, err
	}

	var result AudioAnalysis
	if err := c._get(ctx, fmt.Sprintf("audio-analysis/%s", id), nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ============================================================================
// Category 14: Player Control
// ============================================================================

// CurrentlyPlayingOptions holds options for currently playing
type CurrentlyPlayingOptions struct {
	Market          string // ISO 3166-1 alpha-2 country code
	AdditionalTypes string // Comma-separated: track, episode
}

// CurrentUserPlayingTrack retrieves currently playing track/episode
func (c *Client) CurrentUserPlayingTrack(ctx context.Context, opts *CurrentlyPlayingOptions) (*CurrentlyPlaying, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Market != "" {
			if err := validateMarketParameter(opts.Market); err != nil {
				return nil, err
			}
			params.Set("market", opts.Market)
		}
		if opts.AdditionalTypes != "" {
			params.Set("additional_types", opts.AdditionalTypes)
		}
	}

	var result CurrentlyPlaying
	if err := c._get(ctx, "me/player/currently-playing", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CurrentUserPlaybackState retrieves current playback state
func (c *Client) CurrentUserPlaybackState(ctx context.Context, opts *CurrentlyPlayingOptions) (*PlaybackState, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Market != "" {
			if err := validateMarketParameter(opts.Market); err != nil {
				return nil, err
			}
			params.Set("market", opts.Market)
		}
		if opts.AdditionalTypes != "" {
			params.Set("additional_types", opts.AdditionalTypes)
		}
	}

	var result PlaybackState
	if err := c._get(ctx, "me/player", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CurrentUserDevices retrieves user's available devices
func (c *Client) CurrentUserDevices(ctx context.Context) ([]Device, error) {
	var result struct {
		Devices []Device `json:"devices"`
	}
	if err := c._get(ctx, "me/player/devices", nil, &result); err != nil {
		return nil, err
	}

	return result.Devices, nil
}

// TransferPlaybackOptions holds options for transferring playback
type TransferPlaybackOptions struct {
	Play bool // Whether to start playback
}

// CurrentUserTransferPlayback transfers playback to a device
func (c *Client) CurrentUserTransferPlayback(ctx context.Context, deviceIDs []string, opts *TransferPlaybackOptions) error {
	if len(deviceIDs) == 0 {
		return fmt.Errorf("at least one device ID is required")
	}

	body := map[string]interface{}{
		"device_ids": deviceIDs,
	}
	if opts != nil {
		body["play"] = opts.Play
	}

	return c._put(ctx, "me/player", nil, body, nil)
}

// StartPlaybackOptions holds options for starting playback
type StartPlaybackOptions struct {
	DeviceID   string                 // Device ID
	ContextURI string                 // Spotify URI of context (album, artist, playlist)
	URIs       []string               // Array of Spotify URIs
	Offset     map[string]interface{} // Offset object
	PositionMs *int                   // Position in milliseconds
}

// CurrentUserStartPlayback starts or resumes playback
func (c *Client) CurrentUserStartPlayback(ctx context.Context, opts *StartPlaybackOptions) error {
	params := url.Values{}
	body := map[string]interface{}{}

	if opts != nil {
		if opts.DeviceID != "" {
			params.Set("device_id", opts.DeviceID)
		}
		if opts.ContextURI != "" {
			body["context_uri"] = opts.ContextURI
		}
		if len(opts.URIs) > 0 {
			body["uris"] = opts.URIs
		}
		if opts.Offset != nil {
			body["offset"] = opts.Offset
		}
		if opts.PositionMs != nil {
			body["position_ms"] = *opts.PositionMs
		}
	}

	return c._put(ctx, "me/player/play", params, body, nil)
}

// PausePlaybackOptions holds options for pausing playback
type PausePlaybackOptions struct {
	DeviceID string // Device ID
}

// CurrentUserPausePlayback pauses playback
func (c *Client) CurrentUserPausePlayback(ctx context.Context, opts *PausePlaybackOptions) error {
	params := url.Values{}
	if opts != nil && opts.DeviceID != "" {
		params.Set("device_id", opts.DeviceID)
	}

	return c._put(ctx, "me/player/pause", params, nil, nil)
}

// SeekToPositionOptions holds options for seeking
type SeekToPositionOptions struct {
	PositionMs int    // Position in milliseconds
	DeviceID   string // Device ID
}

// CurrentUserSeekToPosition seeks to position in currently playing track
func (c *Client) CurrentUserSeekToPosition(ctx context.Context, opts *SeekToPositionOptions) error {
	if opts == nil {
		return fmt.Errorf("options are required")
	}

	params := url.Values{}
	params.Set("position_ms", fmt.Sprintf("%d", opts.PositionMs))
	if opts.DeviceID != "" {
		params.Set("device_id", opts.DeviceID)
	}

	return c._put(ctx, "me/player/seek", params, nil, nil)
}

// SetRepeatModeOptions holds options for setting repeat mode
type SetRepeatModeOptions struct {
	State    string // "track", "context", "off"
	DeviceID string // Device ID
}

// CurrentUserSetRepeatMode sets repeat mode
func (c *Client) CurrentUserSetRepeatMode(ctx context.Context, opts *SetRepeatModeOptions) error {
	if opts == nil {
		return fmt.Errorf("options are required")
	}

	params := url.Values{}
	params.Set("state", opts.State)
	if opts.DeviceID != "" {
		params.Set("device_id", opts.DeviceID)
	}

	return c._put(ctx, "me/player/repeat", params, nil, nil)
}

// SetVolumeOptions holds options for setting volume
type SetVolumeOptions struct {
	VolumePercent int    // 0-100
	DeviceID      string // Device ID
}

// CurrentUserSetVolume sets playback volume
func (c *Client) CurrentUserSetVolume(ctx context.Context, opts *SetVolumeOptions) error {
	if opts == nil {
		return fmt.Errorf("options are required")
	}

	if opts.VolumePercent < 0 || opts.VolumePercent > 100 {
		return fmt.Errorf("volume must be between 0 and 100")
	}

	params := url.Values{}
	params.Set("volume_percent", fmt.Sprintf("%d", opts.VolumePercent))
	if opts.DeviceID != "" {
		params.Set("device_id", opts.DeviceID)
	}

	return c._put(ctx, "me/player/volume", params, nil, nil)
}

// ToggleShuffleOptions holds options for toggling shuffle
type ToggleShuffleOptions struct {
	State    bool   // Shuffle state
	DeviceID string // Device ID
}

// CurrentUserToggleShuffle toggles shuffle mode
func (c *Client) CurrentUserToggleShuffle(ctx context.Context, opts *ToggleShuffleOptions) error {
	if opts == nil {
		return fmt.Errorf("options are required")
	}

	params := url.Values{}
	params.Set("state", fmt.Sprintf("%t", opts.State))
	if opts.DeviceID != "" {
		params.Set("device_id", opts.DeviceID)
	}

	return c._put(ctx, "me/player/shuffle", params, nil, nil)
}

// CurrentUserSkipToNext skips to next track
func (c *Client) CurrentUserSkipToNext(ctx context.Context, deviceID ...string) error {
	params := url.Values{}
	if len(deviceID) > 0 && deviceID[0] != "" {
		params.Set("device_id", deviceID[0])
	}

	return c._post(ctx, "me/player/next", params, nil, nil)
}

// CurrentUserSkipToPrevious skips to previous track
func (c *Client) CurrentUserSkipToPrevious(ctx context.Context, deviceID ...string) error {
	params := url.Values{}
	if len(deviceID) > 0 && deviceID[0] != "" {
		params.Set("device_id", deviceID[0])
	}

	return c._post(ctx, "me/player/previous", params, nil, nil)
}

// CurrentUserQueue retrieves the user's current playback queue
func (c *Client) CurrentUserQueue(ctx context.Context) (*QueueResponse, error) {
	var result QueueResponse
	if err := c._get(ctx, "me/player/queue", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CurrentUserAddToQueue adds an item (track or episode) to the user's playback queue
// uri: track or episode URI, URL, or ID
// deviceID: optional device ID to target
func (c *Client) CurrentUserAddToQueue(ctx context.Context, uri string, deviceID ...string) error {
	params := url.Values{}

	// Convert uri to proper format if needed
	// Check if already a URI
	if !IsURI(uri) {
		// Try to convert from URL or ID
		if strings.Contains(uri, "spotify.com") {
			// Extract ID from URL and determine type
			if strings.Contains(uri, "/episode/") {
				id, err := GetID(uri, "episode")
				if err == nil {
					uri, err = GetURI(id, "episode")
					if err != nil {
						return fmt.Errorf("failed to convert episode URL to URI: %w", err)
					}
				}
			} else {
				// Default to track for URLs
				id, err := GetID(uri, "track")
				if err == nil {
					uri, err = GetURI(id, "track")
					if err != nil {
						return fmt.Errorf("failed to convert track URL to URI: %w", err)
					}
				}
			}
		} else {
			// Raw ID - try track first, then episode
			var err error
			uri, err = GetURI(uri, "track")
			if err != nil {
				uri, err = GetURI(uri, "episode")
				if err != nil {
					return fmt.Errorf("invalid URI, URL, or ID: %w", err)
				}
			}
		}
	}

	params.Set("uri", uri)

	if len(deviceID) > 0 && deviceID[0] != "" {
		params.Set("device_id", deviceID[0])
	}

	return c._post(ctx, "me/player/queue", params, nil, nil)
}

// ============================================================================
// Category 15: Markets
// ============================================================================

// AvailableMarkets retrieves list of available markets
func (c *Client) AvailableMarkets(ctx context.Context) ([]string, error) {
	var result struct {
		Markets []string `json:"markets"`
	}
	if err := c._get(ctx, "markets", nil, &result); err != nil {
		return nil, err
	}

	return result.Markets, nil
}

// ============================================================================
// Category 3: Shows & Episodes
// ============================================================================

// Show retrieves a single show by ID, URI, or URL
func (c *Client) Show(ctx context.Context, showID string, market ...string) (*Show, error) {
	id, err := GetID(showID, "show")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if len(market) > 0 && market[0] != "" {
		if err := validateMarketParameter(market[0]); err != nil {
			return nil, err
		}
		params.Set("market", market[0])
	}

	var result Show
	if err := c._get(ctx, fmt.Sprintf("shows/%s", id), params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Shows retrieves multiple shows by IDs, URIs, or URLs
func (c *Client) Shows(ctx context.Context, showIDs []string, market ...string) (*ShowsResponse, error) {
	if len(showIDs) > 50 {
		return nil, fmt.Errorf("maximum 50 shows per request")
	}

	ids := make([]string, len(showIDs))
	for i, id := range showIDs {
		extracted, err := GetID(id, "show")
		if err != nil {
			return nil, err
		}
		ids[i] = extracted
	}

	params := url.Values{}
	params.Set("ids", strings.Join(ids, ","))
	if len(market) > 0 && market[0] != "" {
		if err := validateMarketParameter(market[0]); err != nil {
			return nil, err
		}
		params.Set("market", market[0])
	}

	var result ShowsResponse
	if err := c._get(ctx, "shows", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ShowEpisodesOptions holds options for ShowEpisodes
type ShowEpisodesOptions struct {
	Market string // ISO 3166-1 alpha-2 country code
	Limit  int    // Default: 20, Max: 50
	Offset int    // Default: 0
}

// ShowEpisodes retrieves episodes from a show
func (c *Client) ShowEpisodes(ctx context.Context, showID string, opts *ShowEpisodesOptions) (*Paging[SimplifiedEpisode], error) {
	id, err := GetID(showID, "show")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.Market != "" {
			if err := validateMarketParameter(opts.Market); err != nil {
				return nil, err
			}
			params.Set("market", opts.Market)
		}
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result Paging[SimplifiedEpisode]
	if err := c._get(ctx, fmt.Sprintf("shows/%s/episodes", id), params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Episode retrieves a single episode by ID, URI, or URL
func (c *Client) Episode(ctx context.Context, episodeID string, market ...string) (*Episode, error) {
	id, err := GetID(episodeID, "episode")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if len(market) > 0 && market[0] != "" {
		if err := validateMarketParameter(market[0]); err != nil {
			return nil, err
		}
		params.Set("market", market[0])
	}

	var result Episode
	if err := c._get(ctx, fmt.Sprintf("episodes/%s", id), params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Episodes retrieves multiple episodes by IDs, URIs, or URLs
func (c *Client) Episodes(ctx context.Context, episodeIDs []string, market ...string) (*EpisodesResponse, error) {
	if len(episodeIDs) > 50 {
		return nil, fmt.Errorf("maximum 50 episodes per request")
	}

	ids := make([]string, len(episodeIDs))
	for i, id := range episodeIDs {
		extracted, err := GetID(id, "episode")
		if err != nil {
			return nil, err
		}
		ids[i] = extracted
	}

	params := url.Values{}
	params.Set("ids", strings.Join(ids, ","))
	if len(market) > 0 && market[0] != "" {
		if err := validateMarketParameter(market[0]); err != nil {
			return nil, err
		}
		params.Set("market", market[0])
	}

	var result EpisodesResponse
	if err := c._get(ctx, "episodes", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ============================================================================
// Category 4: Audiobooks
// ============================================================================

// GetAudiobook retrieves a single audiobook by ID, URI, or URL
func (c *Client) GetAudiobook(ctx context.Context, audiobookID string, market ...string) (*Audiobook, error) {
	id, err := GetID(audiobookID, "audiobook")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if len(market) > 0 && market[0] != "" {
		if err := validateMarketParameter(market[0]); err != nil {
			return nil, err
		}
		params.Set("market", market[0])
	}

	var result Audiobook
	if err := c._get(ctx, fmt.Sprintf("audiobooks/%s", id), params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetAudiobooks retrieves multiple audiobooks by IDs, URIs, or URLs
func (c *Client) GetAudiobooks(ctx context.Context, audiobookIDs []string, market ...string) (*AudiobooksResponse, error) {
	if len(audiobookIDs) > 50 {
		return nil, fmt.Errorf("maximum 50 audiobooks per request")
	}

	ids := make([]string, len(audiobookIDs))
	for i, id := range audiobookIDs {
		extracted, err := GetID(id, "audiobook")
		if err != nil {
			return nil, err
		}
		ids[i] = extracted
	}

	params := url.Values{}
	params.Set("ids", strings.Join(ids, ","))
	if len(market) > 0 && market[0] != "" {
		if err := validateMarketParameter(market[0]); err != nil {
			return nil, err
		}
		params.Set("market", market[0])
	}

	var result AudiobooksResponse
	if err := c._get(ctx, "audiobooks", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// AudiobookChaptersOptions holds options for GetAudiobookChapters
type AudiobookChaptersOptions struct {
	Market string // ISO 3166-1 alpha-2 country code
	Limit  int    // Default: 20, Max: 50
	Offset int    // Default: 0
}

// GetAudiobookChapters retrieves chapters from an audiobook
func (c *Client) GetAudiobookChapters(ctx context.Context, audiobookID string, opts *AudiobookChaptersOptions) (*Paging[Chapter], error) {
	id, err := GetID(audiobookID, "audiobook")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if opts != nil {
		// Validate pagination parameters
		if err := validatePaginationParams(opts.Limit, opts.Offset); err != nil {
			return nil, err
		}
		
		if opts.Market != "" {
			if err := validateMarketParameter(opts.Market); err != nil {
				return nil, err
			}
			params.Set("market", opts.Market)
		}
		if opts.Limit > 0 {
			if opts.Limit > 50 {
				opts.Limit = 50
			}
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			params.Set("limit", "20") // Default
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	} else {
		params.Set("limit", "20") // Default
	}

	var result Paging[Chapter]
	if err := c._get(ctx, fmt.Sprintf("audiobooks/%s/chapters", id), params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// _get performs a GET request
func (c *Client) _get(ctx context.Context, urlStr string, params url.Values, result interface{}) error {
	return c._internal_call(ctx, "GET", urlStr, params, nil, result)
}

// _post performs a POST request
func (c *Client) _post(ctx context.Context, urlStr string, params url.Values, body interface{}, result interface{}) error {
	return c._internal_call(ctx, "POST", urlStr, params, body, result)
}

// _put performs a PUT request
func (c *Client) _put(ctx context.Context, urlStr string, params url.Values, body interface{}, result interface{}) error {
	return c._internal_call(ctx, "PUT", urlStr, params, body, result)
}

// _delete performs a DELETE request
func (c *Client) _delete(ctx context.Context, urlStr string, params url.Values, result interface{}) error {
	return c._internal_call(ctx, "DELETE", urlStr, params, nil, result)
}

// PaginationHelper is an interface for paginated results
type PaginationHelper interface {
	GetNext() *string
	GetPrevious() *string
}

// Next retrieves the next page from a paginated result
// Returns (nil, nil) if no next page available (not an error, matching Spotipy behavior)
func (c *Client) Next(ctx context.Context, paging interface{}) (interface{}, error) {
	var nextURL string

	// Try to extract next URL using type assertion
	switch p := paging.(type) {
	case interface{ GetNext() *string }:
		next := p.GetNext()
		if next == nil || *next == "" {
			return nil, nil // No next page - not an error
		}
		nextURL = *next
	case map[string]interface{}:
		if nextVal, ok := p["next"]; ok {
			if nextStr, ok := nextVal.(string); ok && nextStr != "" {
				nextURL = nextStr
			} else {
				return nil, nil // No next page
			}
		} else {
			return nil, nil // No next page
		}
	default:
		// Try reflection or return error
		return nil, fmt.Errorf("unsupported pagination type")
	}

	if nextURL == "" {
		return nil, nil // No next page - not an error
	}

	// Make request to absolute URL (use as-is, don't construct)
	// We need to create a new result of the same type
	// For now, use a generic approach - caller should provide result type
	var result map[string]interface{}
	if err := c._get(ctx, nextURL, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Previous retrieves the previous page from a paginated result
// Returns (nil, nil) if no previous page available (not an error, matching Spotipy behavior)
func (c *Client) Previous(ctx context.Context, paging interface{}) (interface{}, error) {
	var prevURL string

	// Try to extract previous URL
	switch p := paging.(type) {
	case interface{ GetPrevious() *string }:
		prev := p.GetPrevious()
		if prev == nil || *prev == "" {
			return nil, nil // No previous page - not an error
		}
		prevURL = *prev
	case map[string]interface{}:
		if prevVal, ok := p["previous"]; ok {
			if prevStr, ok := prevVal.(string); ok && prevStr != "" {
				prevURL = prevStr
			} else {
				return nil, nil // No previous page
			}
		} else {
			return nil, nil // No previous page
		}
	default:
		return nil, fmt.Errorf("unsupported pagination type")
	}

	if prevURL == "" {
		return nil, nil // No previous page - not an error
	}

	// Make request to absolute URL
	var result map[string]interface{}
	if err := c._get(ctx, prevURL, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// NextGeneric retrieves the next page from a paginated result with type safety using generics
// Returns (nil, nil) if no next page available (not an error)
// This is a type-safe version of Next. The old Next method is kept for backward compatibility
// but will be removed after comprehensive testing.
func NextGeneric[T any](c *Client, ctx context.Context, paging interface{ GetNext() *string }) (*Paging[T], error) {
	next := paging.GetNext()
	if next == nil || *next == "" {
		return nil, nil
	}

	var result Paging[T]
	if err := c._get(ctx, *next, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PreviousGeneric retrieves the previous page from a paginated result with type safety using generics
// Returns (nil, nil) if no previous page available (not an error)
// This is a type-safe version of Previous. The old Previous method is kept for backward compatibility
// but will be removed after comprehensive testing.
func PreviousGeneric[T any](c *Client, ctx context.Context, paging interface{ GetPrevious() *string }) (*Paging[T], error) {
	prev := paging.GetPrevious()
	if prev == nil || *prev == "" {
		return nil, nil
	}

	var result Paging[T]
	if err := c._get(ctx, *prev, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// NextCursor retrieves the next page from a cursor-based paginated result with type safety
// Returns (nil, nil) if no next page available (not an error)
func NextCursor[T any](c *Client, ctx context.Context, paging interface{ GetNext() *string }) (*CursorPaging[T], error) {
	next := paging.GetNext()
	if next == nil || *next == "" {
		return nil, nil
	}

	var result CursorPaging[T]
	if err := c._get(ctx, *next, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PreviousCursor retrieves the previous page from a cursor-based paginated result with type safety
// Returns (nil, nil) if no previous page available (not an error)
func PreviousCursor[T any](c *Client, ctx context.Context, paging interface{ GetPrevious() *string }) (*CursorPaging[T], error) {
	prev := paging.GetPrevious()
	if prev == nil || *prev == "" {
		return nil, nil
	}

	var result CursorPaging[T]
	if err := c._get(ctx, *prev, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
