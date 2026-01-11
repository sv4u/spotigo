package spotigo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sort"
	"strings"
	"time"
)

const (
	// AuthURL is the Spotify OAuth2 authorization endpoint
	AuthURL = "https://accounts.spotify.com/authorize"
	// TokenURL is the Spotify OAuth2 token endpoint
	TokenURL = "https://accounts.spotify.com/api/token"
)

// Environment variable names for OAuth credentials
const (
	EnvClientID     = "SPOTIGO_CLIENT_ID"
	EnvClientSecret = "SPOTIGO_CLIENT_SECRET"
	EnvRedirectURI  = "SPOTIGO_REDIRECT_URI"
	EnvUsername     = "SPOTIGO_CLIENT_USERNAME"
)

// AuthManager defines the interface for authentication managers.
//
// Implementations handle OAuth2 flows and token management for the Spotify API.
// The library provides several implementations:
//   - ClientCredentials for client-only authentication
//   - SpotifyOAuth for authorization code flow
//   - SpotifyPKCE for PKCE flow
//   - SpotifyImplicitGrant for implicit grant flow (deprecated)
type AuthManager interface {
	// GetAccessToken retrieves the current access token (returns token string, not TokenInfo)
	GetAccessToken(ctx context.Context) (string, error)
	// GetCachedToken retrieves the cached token info (separate method for getting full token info)
	GetCachedToken(ctx context.Context) (*TokenInfo, error)
	// RefreshToken refreshes the access token if expired
	RefreshToken(ctx context.Context) error
}

// TokenInfo represents OAuth2 token information returned by Spotify.
//
// It includes the access token, refresh token (if available), expiration time,
// and granted scopes. The ExpiresAt field is calculated from ExpiresIn.
type TokenInfo struct {
	AccessToken      string                 `json:"access_token"`
	TokenType        string                 `json:"token_type"`
	ExpiresIn        int                    `json:"expires_in"`
	ExpiresAt        int                    `json:"expires_at"` // Custom field, calculated
	RefreshToken     string                 `json:"refresh_token,omitempty"`
	Scope            string                 `json:"scope"`
	AdditionalFields map[string]interface{} `json:"-"`
}

// SpotifyAuthBase provides base functionality for all auth managers
type SpotifyAuthBase struct {
	ClientID        string
	ClientSecret    string
	RedirectURI     string
	Scope           string
	HTTPClient      *http.Client
	TokenInfo       *TokenInfo
	CacheHandler    CacheHandler // Will be defined in cache.go
	Proxies         map[string]string
	RequestsTimeout time.Duration
}

// ensureValue checks if a value is provided, otherwise gets it from environment
// Returns error if value is not found in either location
func ensureValue(value, envKey, envVar string) (string, error) {
	if value != "" {
		return value, nil
	}
	envValue := os.Getenv(envVar)
	if envValue != "" {
		return envValue, nil
	}
	return "", &SpotifyOAuthError{
		ErrorType:        "missing_parameter",
		ErrorDescription: fmt.Sprintf("No %s. Pass it or set a %s environment variable.", envKey, envVar),
	}
}

// newHTTPClient creates a properly configured HTTP client with explicit transport settings
// This ensures consistent behavior across Go versions, especially for OAuth token requests
// The explicit configuration helps avoid issues with default HTTP client behavior changes between Go versions
func newHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		// Explicitly configure connection timeouts with sufficient time for TLS handshake
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second, // Increased from default to allow for TLS handshake
			KeepAlive: 30 * time.Second,
		}).DialContext,
		// Configure TLS handshake timeout explicitly
		TLSHandshakeTimeout: 10 * time.Second,
		// Ensure response headers timeout is set
		ResponseHeaderTimeout: 10 * time.Second,
		// HTTP/2 is enabled by default for HTTPS, which is fine
		// The explicit timeouts above help prevent connection resets
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// NewSpotifyAuthBase creates a new base auth manager with environment variable fallback
func NewSpotifyAuthBase(clientID, clientSecret, redirectURI, scope string) (*SpotifyAuthBase, error) {
	// Ensure required values (with environment variable fallback)
	var err error
	clientID, err = ensureValue(clientID, "client_id", EnvClientID)
	if err != nil {
		return nil, err
	}
	clientSecret, err = ensureValue(clientSecret, "client_secret", EnvClientSecret)
	if err != nil {
		return nil, err
	}
	redirectURI, err = ensureValue(redirectURI, "redirect_uri", EnvRedirectURI)
	if err != nil {
		return nil, err
	}

	base := &SpotifyAuthBase{
		ClientID:        clientID,
		ClientSecret:    clientSecret,
		RedirectURI:     redirectURI,
		HTTPClient:      newHTTPClient(5 * time.Second),
		RequestsTimeout: 5 * time.Second,
	}

	// Normalize scope if provided
	if scope != "" {
		base.Scope = NormalizeScope(scope)
	}

	return base, nil
}

// IsTokenExpired checks if token expires within 60 seconds
func (b *SpotifyAuthBase) IsTokenExpired(tokenInfo *TokenInfo) bool {
	if tokenInfo == nil || tokenInfo.ExpiresAt == 0 {
		return true
	}
	now := int(time.Now().Unix())
	return tokenInfo.ExpiresAt-now < 60
}

// NormalizeScope converts scope input to normalized space-separated string
// Accepts string (comma-separated), slice of strings, or empty/nil
func NormalizeScope(scope interface{}) string {
	var scopes []string

	switch s := scope.(type) {
	case string:
		if s == "" {
			return ""
		}
		// Split by commas (not spaces) as per Spotipy behavior
		scopes = strings.Split(s, ",")
		// Trim whitespace from each scope
		for i, scope := range scopes {
			scopes[i] = strings.TrimSpace(scope)
		}
	case []string:
		scopes = s
	case nil:
		return ""
	default:
		// Unsupported type - return empty string (could panic, but being lenient)
		return ""
	}

	// Remove empty strings
	var validScopes []string
	for _, scope := range scopes {
		if scope != "" {
			validScopes = append(validScopes, scope)
		}
	}

	if len(validScopes) == 0 {
		return ""
	}

	// Remove duplicates using a map
	seen := make(map[string]bool)
	unique := []string{}
	for _, scope := range validScopes {
		if !seen[scope] {
			seen[scope] = true
			unique = append(unique, scope)
		}
	}

	// Sort for consistency
	sort.Strings(unique)

	// Join with spaces (not commas)
	return strings.Join(unique, " ")
}

// IsScopeSubset checks if requested scopes are subset of granted scopes
// Both scopes should be space-separated strings
// Returns true if all requested scopes are in granted scopes
func (b *SpotifyAuthBase) IsScopeSubset(requested, granted string) bool {
	// Convert to sets (maps) by splitting on spaces
	requestedSet := make(map[string]bool)
	grantedSet := make(map[string]bool)

	// Handle empty strings
	if requested != "" {
		for _, scope := range strings.Fields(requested) {
			requestedSet[scope] = true
		}
	}
	if granted != "" {
		for _, scope := range strings.Fields(granted) {
			grantedSet[scope] = true
		}
	}

	// Check if all requested scopes are in granted set
	for scope := range requestedSet {
		if !grantedSet[scope] {
			return false
		}
	}

	return true
}

// GetAuthHeader generates Basic authentication header
// Format: "Basic {base64(client_id:client_secret)}"
func (b *SpotifyAuthBase) GetAuthHeader() string {
	credentials := fmt.Sprintf("%s:%s", b.ClientID, b.ClientSecret)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encoded)
}

// ParseTokenResponse parses token response from Spotify and adds expires_at field
func (b *SpotifyAuthBase) ParseTokenResponse(body []byte) (*TokenInfo, error) {
	var tokenInfo TokenInfo
	if err := json.Unmarshal(body, &tokenInfo); err != nil {
		return nil, WrapJSONError(err)
	}

	// Add expires_at field (calculated from expires_in)
	tokenInfo = *b.AddCustomValuesToTokenInfo(&tokenInfo)

	return &tokenInfo, nil
}

// AddCustomValuesToTokenInfo adds expires_at field calculated from expires_in
func (b *SpotifyAuthBase) AddCustomValuesToTokenInfo(tokenInfo *TokenInfo) *TokenInfo {
	if tokenInfo.ExpiresIn > 0 {
		now := int64(time.Now().Unix())
		expiresAt := now + int64(tokenInfo.ExpiresIn)
		// Check for overflow (max int32 value)
		maxInt := int64(1<<31 - 1)
		if expiresAt > maxInt {
			// Token expires too far in future, clamp to max int
			tokenInfo.ExpiresAt = int(maxInt)
		} else {
			tokenInfo.ExpiresAt = int(expiresAt)
		}
	}
	return tokenInfo
}

// HandleOAuthError parses OAuth errors from HTTP responses (JSON or text)
func HandleOAuthError(httpError error, body []byte) error {
	if httpError == nil {
		return nil
	}

	// Try to parse JSON error response
	var oauthErr struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}

	if jsonErr := json.Unmarshal(body, &oauthErr); jsonErr == nil {
		// Successfully parsed JSON
		if oauthErr.Error != "" {
			return &SpotifyOAuthError{
				ErrorType:        oauthErr.Error,
				ErrorDescription: oauthErr.ErrorDescription,
			}
		}
	}

	// If JSON parse failed or error field is empty, use body as text
	errorText := strings.TrimSpace(string(body))
	if errorText == "" {
		errorText = "Unknown OAuth error"
	}

	return &SpotifyOAuthError{
		ErrorType:        "oauth_error",
		ErrorDescription: errorText,
	}
}

// Close closes HTTP client connections (cleanup method)
func (b *SpotifyAuthBase) Close() {
	if b.HTTPClient != nil {
		// HTTP client doesn't need explicit close in Go, but we can set it to nil
		// If using a custom transport with connection pooling, it will be garbage collected
		b.HTTPClient = nil
	}
}

// ClientCredentials implements the Client Credentials OAuth2 flow
type ClientCredentials struct {
	*SpotifyAuthBase
}

// NewClientCredentials creates a new Client Credentials auth manager
// Client Credentials flow doesn't require redirect URI
func NewClientCredentials(clientID, clientSecret string) (*ClientCredentials, error) {
	// Ensure client ID and secret (redirect URI not needed for this flow)
	var err error
	clientID, err = ensureValue(clientID, "client_id", EnvClientID)
	if err != nil {
		return nil, err
	}
	clientSecret, err = ensureValue(clientSecret, "client_secret", EnvClientSecret)
	if err != nil {
		return nil, err
	}

	base := &SpotifyAuthBase{
		ClientID:        clientID,
		ClientSecret:    clientSecret,
		RedirectURI:     "", // Not needed for Client Credentials
		HTTPClient:      newHTTPClient(5 * time.Second),
		RequestsTimeout: 5 * time.Second,
	}

	return &ClientCredentials{SpotifyAuthBase: base}, nil
}

// GetAccessToken retrieves or refreshes the access token
func (c *ClientCredentials) GetAccessToken(ctx context.Context) (string, error) {
	// Check cache first (if cache handler is set)
	if c.CacheHandler != nil {
		cachedToken, err := c.CacheHandler.GetCachedToken(ctx)
		if err == nil && cachedToken != nil {
			// Check if token is expired
			if !c.IsTokenExpired(cachedToken) {
				c.TokenInfo = cachedToken
				return cachedToken.AccessToken, nil
			}
		}
	}

	// Check if we have a non-expired token in memory
	if c.TokenInfo != nil && !c.IsTokenExpired(c.TokenInfo) {
		return c.TokenInfo.AccessToken, nil
	}

	// Request new token with retry logic for transient network errors
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context cancellation before retry attempt
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("request cancelled: %w", ctx.Err())
		default:
		}

		// Request new token
		data := url.Values{}
		data.Set("grant_type", "client_credentials")

		req, err := http.NewRequestWithContext(ctx, "POST", TokenURL, strings.NewReader(data.Encode()))
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", c.GetAuthHeader())

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to execute request: %w", err)
			// Retry on network errors (transient connection issues)
			if attempt < maxRetries && isTransientNetworkError(err) {
				// Exponential backoff: 100ms, 200ms, 400ms
				backoff := time.Duration(100*(1<<uint(attempt))) * time.Millisecond
				select {
				case <-ctx.Done():
					return "", fmt.Errorf("request cancelled: %w", ctx.Err())
				case <-time.After(backoff):
					continue
				}
			}
			return "", lastErr
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			// Retry on read errors
			if attempt < maxRetries {
				backoff := time.Duration(100*(1<<uint(attempt))) * time.Millisecond
				select {
				case <-ctx.Done():
					return "", fmt.Errorf("request cancelled: %w", ctx.Err())
				case <-time.After(backoff):
					continue
				}
			}
			return "", lastErr
		}

		// Handle HTTP errors
		if resp.StatusCode >= 400 {
			// Retry on server errors (5xx) but not client errors (4xx)
			if resp.StatusCode >= 500 && attempt < maxRetries {
				lastErr = HandleOAuthError(fmt.Errorf("HTTP %d", resp.StatusCode), body)
				backoff := time.Duration(100*(1<<uint(attempt))) * time.Millisecond
				select {
				case <-ctx.Done():
					return "", fmt.Errorf("request cancelled: %w", ctx.Err())
				case <-time.After(backoff):
					continue
				}
			}
			return "", HandleOAuthError(fmt.Errorf("HTTP %d", resp.StatusCode), body)
		}

		// Parse token response
		tokenInfo, err := c.ParseTokenResponse(body)
		if err != nil {
			return "", err
		}

		// Store token
		c.TokenInfo = tokenInfo

		// Save to cache if cache handler is set
		if c.CacheHandler != nil {
			_ = c.CacheHandler.SaveTokenToCache(ctx, tokenInfo) // Ignore cache errors
		}

		return tokenInfo.AccessToken, nil
	}

	return "", fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// isTransientNetworkError checks if an error is a transient network error that should be retried
func isTransientNetworkError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Check for common transient network errors
	transientErrors := []string{
		"connection reset",
		"connection terminated",
		"upstream connect error",
		"disconnect/reset before headers",
		"EOF",
		"broken pipe",
		"timeout",
		"temporary failure",
		"no such host",
	}
	for _, transientErr := range transientErrors {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(transientErr)) {
			return true
		}
	}
	return false
}

// GetCachedToken returns the cached token info
func (c *ClientCredentials) GetCachedToken(ctx context.Context) (*TokenInfo, error) {
	// Check cache handler first
	if c.CacheHandler != nil {
		cachedToken, err := c.CacheHandler.GetCachedToken(ctx)
		if err == nil && cachedToken != nil {
			c.TokenInfo = cachedToken
			return cachedToken, nil
		}
	}

	// Fall back to in-memory token
	if c.TokenInfo == nil {
		return nil, fmt.Errorf("no token cached")
	}
	return c.TokenInfo, nil
}

// RefreshToken refreshes the access token
func (c *ClientCredentials) RefreshToken(ctx context.Context) error {
	// Client Credentials flow doesn't have refresh tokens, so we just request a new one
	_, err := c.GetAccessToken(ctx)
	return err
}

// SpotifyOAuth implements the Authorization Code OAuth2 flow
type SpotifyOAuth struct {
	*SpotifyAuthBase
	State       string
	OpenBrowser bool
	ShowDialog  bool
}

// NewSpotifyOAuth creates a new Authorization Code auth manager
func NewSpotifyOAuth(clientID, clientSecret, redirectURI, scope string) (*SpotifyOAuth, error) {
	base, err := NewSpotifyAuthBase(clientID, clientSecret, redirectURI, scope)
	if err != nil {
		return nil, err
	}
	return &SpotifyOAuth{
		SpotifyAuthBase: base,
		OpenBrowser:     true,
	}, nil
}

// GetAuthURL generates the authorization URL
func (o *SpotifyOAuth) GetAuthURL(state string, showDialog bool) (string, error) {
	params := url.Values{}
	params.Set("client_id", o.ClientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", o.RedirectURI)

	if o.Scope != "" {
		params.Set("scope", o.Scope)
	}

	// Use provided state or stored state
	useState := state
	if useState == "" {
		useState = o.State
	}
	if useState != "" {
		params.Set("state", useState)
		o.State = useState // Store for validation
	}

	if showDialog || o.ShowDialog {
		params.Set("show_dialog", "true")
	}

	return fmt.Sprintf("%s?%s", AuthURL, params.Encode()), nil
}

// GetAccessToken retrieves or refreshes the access token
func (o *SpotifyOAuth) GetAccessToken(ctx context.Context) (string, error) {
	// Check cache first
	if o.CacheHandler != nil {
		cachedToken, err := o.CacheHandler.GetCachedToken(ctx)
		if err == nil && cachedToken != nil {
			// Check if token is expired
			if !o.IsTokenExpired(cachedToken) {
				o.TokenInfo = cachedToken
				return cachedToken.AccessToken, nil
			}
			// Try to refresh if we have a refresh token
			if cachedToken.RefreshToken != "" {
				if err := o.RefreshToken(ctx); err == nil {
					return o.TokenInfo.AccessToken, nil
				}
			}
		}
	}

	// Check in-memory token
	if o.TokenInfo != nil {
		if !o.IsTokenExpired(o.TokenInfo) {
			return o.TokenInfo.AccessToken, nil
		}
		// Try to refresh
		if o.TokenInfo.RefreshToken != "" {
			if err := o.RefreshToken(ctx); err == nil {
				return o.TokenInfo.AccessToken, nil
			}
		}
	}

	// No valid token, user must authorize first
	return "", &SpotifyOAuthError{
		ErrorType:        "no_token",
		ErrorDescription: "No access token available. User must authorize first.",
	}
}

// GetCachedToken returns the cached token info
func (o *SpotifyOAuth) GetCachedToken(ctx context.Context) (*TokenInfo, error) {
	if o.CacheHandler != nil {
		cachedToken, err := o.CacheHandler.GetCachedToken(ctx)
		if err == nil && cachedToken != nil {
			o.TokenInfo = cachedToken
			return cachedToken, nil
		}
	}

	if o.TokenInfo == nil {
		return nil, fmt.Errorf("no token cached")
	}
	return o.TokenInfo, nil
}

// GetAuthorizationCode performs the interactive authorization flow
func (o *SpotifyOAuth) GetAuthorizationCode(ctx context.Context, openBrowser bool) (string, error) {
	// Generate state if not set
	if o.State == "" {
		state, err := GenerateRandomState()
		if err != nil {
			return "", fmt.Errorf("failed to generate state: %w", err)
		}
		o.State = state
	}

	// Get authorization URL
	authURL, err := o.GetAuthURL(o.State, o.ShowDialog)
	if err != nil {
		return "", err
	}

	// Parse redirect URI to check if we should start local server
	redirectURL, err := url.Parse(o.RedirectURI)
	if err != nil {
		return "", fmt.Errorf("invalid redirect URI: %w", err)
	}

	host, port := GetHostPort(redirectURL.Host)

	// Check for deprecated localhost usage
	if host == "localhost" {
		log.Printf("Warning: Using 'localhost' as a redirect URI is being deprecated. Use a loopback IP address such as 127.0.0.1 to ensure your app remains functional.")
	}

	// Check for deprecated HTTP usage
	if redirectURL.Scheme == "http" && host != "127.0.0.1" && host != "localhost" {
		log.Printf("Warning: Redirect URIs using HTTP are being deprecated. To ensure your app remains functional, use HTTPS instead.")
	}

	// Check if we should start local server
	shouldStartServer := (openBrowser || o.OpenBrowser) &&
		(host == "127.0.0.1" || host == "localhost") &&
		redirectURL.Scheme == "http" &&
		port != nil

	if !shouldStartServer && port == nil && (host == "127.0.0.1" || host == "localhost") {
		log.Printf("Warning: Using '%s' as redirect URI without a port. Specify a port (e.g. '%s:8080') to allow automatic retrieval of authentication code instead of having to copy and paste the URL your browser is redirected to.", host, host)
	}

	// Open browser if requested
	if openBrowser || o.OpenBrowser {
		if err := openBrowserURL(authURL); err != nil {
			log.Printf("Warning: Failed to open browser: %v", err)
		}
	}

	// Start local server if conditions are met
	if shouldStartServer {
		return o.startLocalServer(ctx, *port)
	}

	// Manual flow - return URL for user to visit
	return "", fmt.Errorf("manual authorization required. Visit: %s", authURL)
}

// ExchangeCode exchanges authorization code for tokens
func (o *SpotifyOAuth) ExchangeCode(ctx context.Context, code string) error {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", o.RedirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", o.GetAuthHeader())

	resp, err := o.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Handle errors
	if resp.StatusCode >= 400 {
		return HandleOAuthError(fmt.Errorf("HTTP %d", resp.StatusCode), body)
	}

	// Parse token response
	tokenInfo, err := o.ParseTokenResponse(body)
	if err != nil {
		return err
	}

	// Validate scopes if requested
	if o.Scope != "" && tokenInfo.Scope != "" {
		if !o.IsScopeSubset(o.Scope, tokenInfo.Scope) {
			return &SpotifyOAuthError{
				ErrorType:        "invalid_scope",
				ErrorDescription: "Granted scopes do not include all requested scopes",
			}
		}
	}

	// Store token
	o.TokenInfo = tokenInfo

	// Save to cache
	if o.CacheHandler != nil {
		_ = o.CacheHandler.SaveTokenToCache(ctx, tokenInfo)
	}

	return nil
}

// RefreshToken refreshes the access token using refresh token
func (o *SpotifyOAuth) RefreshToken(ctx context.Context) error {
	if o.TokenInfo == nil || o.TokenInfo.RefreshToken == "" {
		return &SpotifyOAuthError{
			ErrorType:        "no_refresh_token",
			ErrorDescription: "No refresh token available",
		}
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", o.TokenInfo.RefreshToken)

	req, err := http.NewRequestWithContext(ctx, "POST", TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", o.GetAuthHeader())

	resp, err := o.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Handle errors
	if resp.StatusCode >= 400 {
		return HandleOAuthError(fmt.Errorf("HTTP %d", resp.StatusCode), body)
	}

	// Parse new token response
	newTokenInfo, err := o.ParseTokenResponse(body)
	if err != nil {
		return err
	}

	// Preserve refresh token if not in response
	if newTokenInfo.RefreshToken == "" {
		newTokenInfo.RefreshToken = o.TokenInfo.RefreshToken
	}

	// Update token info
	o.TokenInfo = newTokenInfo

	// Save to cache
	if o.CacheHandler != nil {
		_ = o.CacheHandler.SaveTokenToCache(ctx, o.TokenInfo)
	}

	return nil
}

// localServerState holds state for the local HTTP server
type localServerState struct {
	authCode string
	state    string
	err      error
	done     chan bool
}

// startLocalServer starts a local HTTP server to receive the OAuth callback
func (o *SpotifyOAuth) startLocalServer(ctx context.Context, port int) (string, error) {
	state := &localServerState{
		done: make(chan bool, 1),
	}

	server := &http.Server{
		Addr: fmt.Sprintf("127.0.0.1:%d", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Parse callback URL
			code, receivedState, err := ParseAuthResponseURL(r.URL.String())
			if err != nil {
				state.err = err
				state.done <- true
				o.sendCallbackResponse(w, false, err.Error())
				return
			}

			state.state = receivedState
			state.authCode = code
			state.done <- true

			// Send success response
			o.sendCallbackResponse(w, true, "")
		}),
	}

	// Start server in goroutine
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return "", fmt.Errorf("failed to start local server: %w", err)
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			state.err = fmt.Errorf("server error: %w", err)
			state.done <- true
		}
	}()

	// Wait for callback or context cancellation
	select {
	case <-state.done:
		// Shutdown server
		_ = server.Shutdown(ctx)
	case <-ctx.Done():
		_ = server.Shutdown(ctx)
		return "", ctx.Err()
	}

	// Check results in order
	if state.err != nil {
		return "", state.err
	}

	if o.State != "" && state.state != o.State {
		return "", &SpotifyStateError{
			SpotifyOAuthError: &SpotifyOAuthError{
				ErrorType:        "state_mismatch",
				ErrorDescription: "State parameter mismatch",
			},
			LocalState:  o.State,
			RemoteState: state.state,
		}
	}

	if state.authCode != "" {
		return state.authCode, nil
	}

	return "", &SpotifyOAuthError{
		ErrorType:        "no_code",
		ErrorDescription: "Server listening on localhost has not been accessed",
	}
}

// sendCallbackResponse sends an HTML response to the browser
func (o *SpotifyOAuth) sendCallbackResponse(w http.ResponseWriter, success bool, errorMsg string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	status := "successful"
	if !success {
		status = fmt.Sprintf("failed (%s)", html.EscapeString(errorMsg))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>Spotify Authorization</title></head>
<body>
	<h1>Authorization %s</h1>
	<p>You can close this window.</p>
	<button onclick="window.close();">Close Window</button>
	<script>window.close();</script>
</body>
</html>`, status)

	w.Write([]byte(html))
}

// openBrowserURL opens the given URL in the default browser
func openBrowserURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	return cmd.Run()
}

// SpotifyPKCE implements the PKCE OAuth2 flow
type SpotifyPKCE struct {
	*SpotifyAuthBase
	CodeVerifier  string
	CodeChallenge string
	State         string
	OpenBrowser   bool
	ShowDialog    bool
}

// NewSpotifyPKCE creates a new PKCE auth manager
// PKCE doesn't require client secret
func NewSpotifyPKCE(clientID, redirectURI, scope string) (*SpotifyPKCE, error) {
	// Ensure client ID (redirect URI and scope are optional)
	var err error
	clientID, err = ensureValue(clientID, "client_id", EnvClientID)
	if err != nil {
		return nil, err
	}

	// Redirect URI is optional for PKCE, but recommended
	redirectURI, _ = ensureValue(redirectURI, "redirect_uri", EnvRedirectURI)

	base := &SpotifyAuthBase{
		ClientID:        clientID,
		ClientSecret:    "", // PKCE doesn't use client secret
		RedirectURI:     redirectURI,
		HTTPClient:      newHTTPClient(5 * time.Second),
		RequestsTimeout: 5 * time.Second,
	}

	// Normalize scope if provided
	if scope != "" {
		base.Scope = NormalizeScope(scope)
	}

	return &SpotifyPKCE{
		SpotifyAuthBase: base,
		OpenBrowser:     true,
	}, nil
}

// GenerateCodeVerifier generates a new code verifier
// Length is between 43-128 characters (URL-safe base64)
func (p *SpotifyPKCE) GenerateCodeVerifier() (string, error) {
	// Generate random length between 33-96 bytes (which gives 44-128 base64 chars)
	// PKCE spec says verifier should be 43-128 characters
	// We'll generate 32-96 random bytes to get 43-128 base64 characters
	// Use a fixed length of 64 bytes for simplicity (gives ~86 base64 chars)
	length := 64

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate code verifier: %w", err)
	}

	// Encode as URL-safe base64 (no padding)
	verifier := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
	p.CodeVerifier = verifier
	return verifier, nil
}

// GenerateCodeChallenge generates code challenge from verifier using S256 method
func (p *SpotifyPKCE) GenerateCodeChallenge(verifier string) string {
	// SHA256 hash the verifier
	hash := sha256.Sum256([]byte(verifier))
	// Base64URL encode (no padding)
	challenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
	p.CodeChallenge = challenge
	return challenge
}

// GetAuthURL generates authorization URL with PKCE parameters
func (p *SpotifyPKCE) GetAuthURL(state string, showDialog bool) (string, error) {
	// Generate code verifier and challenge if not set
	if p.CodeVerifier == "" {
		if _, err := p.GenerateCodeVerifier(); err != nil {
			return "", err
		}
	}
	if p.CodeChallenge == "" {
		p.GenerateCodeChallenge(p.CodeVerifier)
	}

	params := url.Values{}
	params.Set("client_id", p.ClientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", p.RedirectURI)
	params.Set("code_challenge", p.CodeChallenge)
	params.Set("code_challenge_method", "S256")

	if p.Scope != "" {
		params.Set("scope", p.Scope)
	}

	// Use provided state or stored state
	useState := state
	if useState == "" {
		useState = p.State
	}
	if useState != "" {
		params.Set("state", useState)
		p.State = useState
	}

	if showDialog || p.ShowDialog {
		params.Set("show_dialog", "true")
	}

	return fmt.Sprintf("%s?%s", AuthURL, params.Encode()), nil
}

// GetAuthorizationCode performs the interactive authorization flow (same as SpotifyOAuth)
func (p *SpotifyPKCE) GetAuthorizationCode(ctx context.Context, openBrowser bool) (string, error) {
	// Generate state if not set
	if p.State == "" {
		state, err := GenerateRandomState()
		if err != nil {
			return "", fmt.Errorf("failed to generate state: %w", err)
		}
		p.State = state
	}

	// Get authorization URL
	authURL, err := p.GetAuthURL(p.State, p.ShowDialog)
	if err != nil {
		return "", err
	}

	// Parse redirect URI
	redirectURL, err := url.Parse(p.RedirectURI)
	if err != nil {
		return "", fmt.Errorf("invalid redirect URI: %w", err)
	}

	host, port := GetHostPort(redirectURL.Host)

	// Check for deprecated usage (same as SpotifyOAuth)
	if host == "localhost" {
		log.Printf("Warning: Using 'localhost' as a redirect URI is being deprecated. Use a loopback IP address such as 127.0.0.1 to ensure your app remains functional.")
	}

	if redirectURL.Scheme == "http" && host != "127.0.0.1" && host != "localhost" {
		log.Printf("Warning: Redirect URIs using HTTP are being deprecated. To ensure your app remains functional, use HTTPS instead.")
	}

	shouldStartServer := (openBrowser || p.OpenBrowser) &&
		(host == "127.0.0.1" || host == "localhost") &&
		redirectURL.Scheme == "http" &&
		port != nil

	if !shouldStartServer && port == nil && (host == "127.0.0.1" || host == "localhost") {
		log.Printf("Warning: Using '%s' as redirect URI without a port. Specify a port (e.g. '%s:8080') to allow automatic retrieval of authentication code instead of having to copy and paste the URL your browser is redirected to.", host, host)
	}

	// Open browser if requested
	if openBrowser || p.OpenBrowser {
		if err := openBrowserURL(authURL); err != nil {
			log.Printf("Warning: Failed to open browser: %v", err)
		}
	}

	// Start local server if conditions are met
	if shouldStartServer {
		return p.startLocalServer(ctx, *port)
	}

	return "", fmt.Errorf("manual authorization required. Visit: %s", authURL)
}

// ExchangeCode exchanges authorization code for tokens using code verifier (no client secret)
func (p *SpotifyPKCE) ExchangeCode(ctx context.Context, code string) error {
	if p.CodeVerifier == "" {
		return fmt.Errorf("code verifier not set")
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", p.RedirectURI)
	data.Set("client_id", p.ClientID) // PKCE includes client_id in body
	data.Set("code_verifier", p.CodeVerifier)

	req, err := http.NewRequestWithContext(ctx, "POST", TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// No Basic auth header for PKCE

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Handle errors
	if resp.StatusCode >= 400 {
		return HandleOAuthError(fmt.Errorf("HTTP %d", resp.StatusCode), body)
	}

	// Parse token response
	tokenInfo, err := p.ParseTokenResponse(body)
	if err != nil {
		return err
	}

	// Validate scopes if requested
	if p.Scope != "" && tokenInfo.Scope != "" {
		if !p.IsScopeSubset(p.Scope, tokenInfo.Scope) {
			return &SpotifyOAuthError{
				ErrorType:        "invalid_scope",
				ErrorDescription: "Granted scopes do not include all requested scopes",
			}
		}
	}

	// Store token
	p.TokenInfo = tokenInfo

	// Save to cache
	if p.CacheHandler != nil {
		_ = p.CacheHandler.SaveTokenToCache(ctx, tokenInfo)
	}

	return nil
}

// GetAccessToken retrieves or refreshes the access token
func (p *SpotifyPKCE) GetAccessToken(ctx context.Context) (string, error) {
	// Check cache first
	if p.CacheHandler != nil {
		cachedToken, err := p.CacheHandler.GetCachedToken(ctx)
		if err == nil && cachedToken != nil {
			if !p.IsTokenExpired(cachedToken) {
				p.TokenInfo = cachedToken
				return cachedToken.AccessToken, nil
			}
			if cachedToken.RefreshToken != "" {
				if err := p.RefreshToken(ctx); err == nil {
					return p.TokenInfo.AccessToken, nil
				}
			}
		}
	}

	// Check in-memory token
	if p.TokenInfo != nil {
		if !p.IsTokenExpired(p.TokenInfo) {
			return p.TokenInfo.AccessToken, nil
		}
		if p.TokenInfo.RefreshToken != "" {
			if err := p.RefreshToken(ctx); err == nil {
				return p.TokenInfo.AccessToken, nil
			}
		}
	}

	return "", &SpotifyOAuthError{
		ErrorType:        "no_token",
		ErrorDescription: "No access token available. User must authorize first.",
	}
}

// GetCachedToken returns the cached token info
func (p *SpotifyPKCE) GetCachedToken(ctx context.Context) (*TokenInfo, error) {
	if p.CacheHandler != nil {
		cachedToken, err := p.CacheHandler.GetCachedToken(ctx)
		if err == nil && cachedToken != nil {
			p.TokenInfo = cachedToken
			return cachedToken, nil
		}
	}

	if p.TokenInfo == nil {
		return nil, fmt.Errorf("no token cached")
	}
	return p.TokenInfo, nil
}

// RefreshToken refreshes the access token using refresh token
// For PKCE, include client_id in payload (no Basic auth header)
func (p *SpotifyPKCE) RefreshToken(ctx context.Context) error {
	if p.TokenInfo == nil || p.TokenInfo.RefreshToken == "" {
		return &SpotifyOAuthError{
			ErrorType:        "no_refresh_token",
			ErrorDescription: "No refresh token available",
		}
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", p.TokenInfo.RefreshToken)
	data.Set("client_id", p.ClientID) // PKCE includes client_id in body

	req, err := http.NewRequestWithContext(ctx, "POST", TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// No Basic auth header for PKCE

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Handle errors
	if resp.StatusCode >= 400 {
		return HandleOAuthError(fmt.Errorf("HTTP %d", resp.StatusCode), body)
	}

	// Parse new token response
	newTokenInfo, err := p.ParseTokenResponse(body)
	if err != nil {
		return err
	}

	// Preserve refresh token if not in response
	if newTokenInfo.RefreshToken == "" {
		newTokenInfo.RefreshToken = p.TokenInfo.RefreshToken
	}

	// Update token info
	p.TokenInfo = newTokenInfo

	// Save to cache
	if p.CacheHandler != nil {
		_ = p.CacheHandler.SaveTokenToCache(ctx, p.TokenInfo)
	}

	return nil
}

// startLocalServer starts a local HTTP server for PKCE (same as SpotifyOAuth)
func (p *SpotifyPKCE) startLocalServer(ctx context.Context, port int) (string, error) {
	state := &localServerState{
		done: make(chan bool, 1),
	}

	server := &http.Server{
		Addr: fmt.Sprintf("127.0.0.1:%d", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			code, receivedState, err := ParseAuthResponseURL(r.URL.String())
			if err != nil {
				state.err = err
				state.done <- true
				p.sendCallbackResponse(w, false, err.Error())
				return
			}

			state.state = receivedState
			state.authCode = code
			state.done <- true

			p.sendCallbackResponse(w, true, "")
		}),
	}

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return "", fmt.Errorf("failed to start local server: %w", err)
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			state.err = fmt.Errorf("server error: %w", err)
			state.done <- true
		}
	}()

	select {
	case <-state.done:
		_ = server.Shutdown(ctx)
	case <-ctx.Done():
		_ = server.Shutdown(ctx)
		return "", ctx.Err()
	}

	// For PKCE, check state before error (slightly different order)
	if p.State != "" && state.state != p.State {
		return "", &SpotifyStateError{
			SpotifyOAuthError: &SpotifyOAuthError{
				ErrorType:        "state_mismatch",
				ErrorDescription: "State parameter mismatch",
			},
			LocalState:  p.State,
			RemoteState: state.state,
		}
	}

	if state.err != nil {
		return "", state.err
	}

	if state.authCode != "" {
		return state.authCode, nil
	}

	return "", &SpotifyOAuthError{
		ErrorType:        "no_code",
		ErrorDescription: "Server listening on localhost has not been accessed",
	}
}

// sendCallbackResponse sends HTML response (shared method)
func (p *SpotifyPKCE) sendCallbackResponse(w http.ResponseWriter, success bool, errorMsg string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	status := "successful"
	if !success {
		status = fmt.Sprintf("failed (%s)", html.EscapeString(errorMsg))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>Spotify Authorization</title></head>
<body>
	<h1>Authorization %s</h1>
	<p>You can close this window.</p>
	<button onclick="window.close();">Close Window</button>
	<script>window.close();</script>
</body>
</html>`, status)

	w.Write([]byte(html))
}

// SpotifyImplicitGrant implements the Implicit Grant OAuth2 flow (deprecated)
type SpotifyImplicitGrant struct {
	*SpotifyAuthBase
	State       string
	OpenBrowser bool
	ShowDialog  bool
}

// NewSpotifyImplicitGrant creates a new Implicit Grant auth manager
// Implicit Grant doesn't require client secret
func NewSpotifyImplicitGrant(clientID, redirectURI, scope string) (*SpotifyImplicitGrant, error) {
	// Ensure client ID
	var err error
	clientID, err = ensureValue(clientID, "client_id", EnvClientID)
	if err != nil {
		return nil, err
	}

	redirectURI, _ = ensureValue(redirectURI, "redirect_uri", EnvRedirectURI)

	base := &SpotifyAuthBase{
		ClientID:        clientID,
		ClientSecret:    "", // Implicit Grant doesn't use client secret
		RedirectURI:     redirectURI,
		HTTPClient:      newHTTPClient(5 * time.Second),
		RequestsTimeout: 5 * time.Second,
	}

	if scope != "" {
		base.Scope = NormalizeScope(scope)
	}

	log.Printf("Warning: Implicit Grant flow is deprecated. Use PKCE flow instead. Tokens expire after 1 hour and cannot be refreshed.")

	return &SpotifyImplicitGrant{
		SpotifyAuthBase: base,
		OpenBrowser:     true,
	}, nil
}

// GetAuthURL generates authorization URL with response_type=token
func (i *SpotifyImplicitGrant) GetAuthURL(state string, showDialog bool) (string, error) {
	params := url.Values{}
	params.Set("client_id", i.ClientID)
	params.Set("response_type", "token") // Implicit Grant uses "token" not "code"
	params.Set("redirect_uri", i.RedirectURI)

	if i.Scope != "" {
		params.Set("scope", i.Scope)
	}

	useState := state
	if useState == "" {
		useState = i.State
	}
	if useState != "" {
		params.Set("state", useState)
		i.State = useState
	}

	if showDialog || i.ShowDialog {
		params.Set("show_dialog", "true")
	}

	return fmt.Sprintf("%s?%s", AuthURL, params.Encode()), nil
}

// ParseTokenFromURL extracts token from URL fragment
// Format: #access_token=...&expires_in=...&scope=...&state=...
func (i *SpotifyImplicitGrant) ParseTokenFromURL(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Parse fragment (not query parameters)
	fragment := parsedURL.Fragment
	if fragment == "" {
		return fmt.Errorf("no fragment in URL")
	}

	// Parse fragment as query string
	fragmentValues, err := url.ParseQuery(fragment)
	if err != nil {
		return fmt.Errorf("invalid fragment: %w", err)
	}

	// Check for error
	if errorParam := fragmentValues.Get("error"); errorParam != "" {
		errorDesc := fragmentValues.Get("error_description")
		return &SpotifyOAuthError{
			ErrorType:        errorParam,
			ErrorDescription: errorDesc,
		}
	}

	// Extract token information
	accessToken := fragmentValues.Get("access_token")
	if accessToken == "" {
		return fmt.Errorf("no access_token in fragment")
	}

	expiresInStr := fragmentValues.Get("expires_in")
	expiresIn := 3600 // Default 1 hour
	if expiresInStr != "" {
		if ei, err := strconv.Atoi(expiresInStr); err == nil {
			expiresIn = ei
		}
	}

	scope := fragmentValues.Get("scope")
	receivedState := fragmentValues.Get("state")

	// Validate state
	if i.State != "" && receivedState != i.State {
		return &SpotifyStateError{
			SpotifyOAuthError: &SpotifyOAuthError{
				ErrorType:        "state_mismatch",
				ErrorDescription: "State parameter mismatch",
			},
			LocalState:  i.State,
			RemoteState: receivedState,
		}
	}

	// Create token info
	tokenInfo := &TokenInfo{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		Scope:       scope,
		// No refresh token in Implicit Grant
	}

	// Calculate expires_at
	tokenInfo = i.AddCustomValuesToTokenInfo(tokenInfo)

	i.TokenInfo = tokenInfo

	// Save to cache
	if i.CacheHandler != nil {
		_ = i.CacheHandler.SaveTokenToCache(context.Background(), tokenInfo)
	}

	return nil
}

// GetAccessToken retrieves the access token
func (i *SpotifyImplicitGrant) GetAccessToken(ctx context.Context) (string, error) {
	// Check cache first
	if i.CacheHandler != nil {
		cachedToken, err := i.CacheHandler.GetCachedToken(ctx)
		if err == nil && cachedToken != nil {
			if !i.IsTokenExpired(cachedToken) {
				i.TokenInfo = cachedToken
				return cachedToken.AccessToken, nil
			}
			// Token expired - no refresh token available
			return "", &SpotifyOAuthError{
				ErrorType:        "token_expired",
				ErrorDescription: "Token expired and cannot be refreshed (Implicit Grant flow)",
			}
		}
	}

	// Check in-memory token
	if i.TokenInfo != nil {
		if !i.IsTokenExpired(i.TokenInfo) {
			return i.TokenInfo.AccessToken, nil
		}
		// Token expired
		return "", &SpotifyOAuthError{
			ErrorType:        "token_expired",
			ErrorDescription: "Token expired and cannot be refreshed (Implicit Grant flow)",
		}
	}

	return "", &SpotifyOAuthError{
		ErrorType:        "no_token",
		ErrorDescription: "No access token available. User must authorize first.",
	}
}

// GetCachedToken returns the cached token info
func (i *SpotifyImplicitGrant) GetCachedToken(ctx context.Context) (*TokenInfo, error) {
	if i.CacheHandler != nil {
		cachedToken, err := i.CacheHandler.GetCachedToken(ctx)
		if err == nil && cachedToken != nil {
			i.TokenInfo = cachedToken
			return cachedToken, nil
		}
	}

	if i.TokenInfo == nil {
		return nil, fmt.Errorf("no token cached")
	}
	return i.TokenInfo, nil
}

// RefreshToken is not supported in Implicit Grant flow
func (i *SpotifyImplicitGrant) RefreshToken(ctx context.Context) error {
	return &SpotifyOAuthError{
		ErrorType:        "not_supported",
		ErrorDescription: "Refresh token is not supported in Implicit Grant flow",
	}
}

// GetAuthorizationCode performs the interactive authorization flow for Implicit Grant
func (i *SpotifyImplicitGrant) GetAuthorizationCode(ctx context.Context, openBrowser bool) (string, error) {
	// Generate state if not set
	if i.State == "" {
		state, err := GenerateRandomState()
		if err != nil {
			return "", fmt.Errorf("failed to generate state: %w", err)
		}
		i.State = state
	}

	// Get authorization URL
	authURL, err := i.GetAuthURL(i.State, i.ShowDialog)
	if err != nil {
		return "", err
	}

	// Parse redirect URI
	redirectURL, err := url.Parse(i.RedirectURI)
	if err != nil {
		return "", fmt.Errorf("invalid redirect URI: %w", err)
	}

	host, port := GetHostPort(redirectURL.Host)

	// Check for deprecated usage
	if host == "localhost" {
		log.Printf("Warning: Using 'localhost' as a redirect URI is being deprecated. Use a loopback IP address such as 127.0.0.1 to ensure your app remains functional.")
	}

	if redirectURL.Scheme == "http" && host != "127.0.0.1" && host != "localhost" {
		log.Printf("Warning: Redirect URIs using HTTP are being deprecated. To ensure your app remains functional, use HTTPS instead.")
	}

	shouldStartServer := (openBrowser || i.OpenBrowser) &&
		(host == "127.0.0.1" || host == "localhost") &&
		redirectURL.Scheme == "http" &&
		port != nil

	if !shouldStartServer && port == nil && (host == "127.0.0.1" || host == "localhost") {
		log.Printf("Warning: Using '%s' as redirect URI without a port. Specify a port (e.g. '%s:8080') to allow automatic retrieval of authentication code instead of having to copy and paste the URL your browser is redirected to.", host, host)
	}

	// Open browser if requested
	if openBrowser || i.OpenBrowser {
		if err := openBrowserURL(authURL); err != nil {
			log.Printf("Warning: Failed to open browser: %v", err)
		}
	}

	// Start local server if conditions are met
	if shouldStartServer {
		return i.startLocalServer(ctx, *port)
	}

	return "", fmt.Errorf("manual authorization required. Visit: %s", authURL)
}

// startLocalServer starts a local HTTP server for Implicit Grant (parses fragment)
func (i *SpotifyImplicitGrant) startLocalServer(ctx context.Context, port int) (string, error) {
	state := &localServerState{
		done: make(chan bool, 1),
	}

	server := &http.Server{
		Addr: fmt.Sprintf("127.0.0.1:%d", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For Implicit Grant, token is in fragment, not query
			// Fragments are not sent to server, so we serve a page that extracts it via JavaScript
			if r.URL.Query().Get("fragment") != "" {
				// Fragment was extracted by JavaScript and sent as query param
				fragment := r.URL.Query().Get("fragment")
				fullURL := r.URL.Scheme + "://" + r.Host + r.URL.Path + "#" + fragment

				err := i.ParseTokenFromURL(fullURL)
				if err != nil {
					state.err = err
					state.done <- true
					i.sendCallbackResponse(w, false, err.Error())
					return
				}

				state.done <- true
				i.sendCallbackResponse(w, true, "")
			} else {
				// Serve page that extracts fragment and sends it back
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusOK)
				html := `<!DOCTYPE html>
<html>
<head><title>Spotify Authorization</title></head>
<body>
	<script>
		// Extract fragment and send it to server
		var fragment = window.location.hash.substring(1);
		if (fragment) {
			window.location.href = window.location.pathname + '?fragment=' + encodeURIComponent(fragment);
		} else {
			document.body.innerHTML = '<h1>No token found in URL fragment</h1>';
		}
	</script>
</body>
</html>`
				w.Write([]byte(html))
			}
		}),
	}

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return "", fmt.Errorf("failed to start local server: %w", err)
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			state.err = fmt.Errorf("server error: %w", err)
			state.done <- true
		}
	}()

	select {
	case <-state.done:
		_ = server.Shutdown(ctx)
	case <-ctx.Done():
		_ = server.Shutdown(ctx)
		return "", ctx.Err()
	}

	if state.err != nil {
		return "", state.err
	}

	// Return empty string (token is already parsed and stored)
	return "", nil
}

// sendCallbackResponse sends HTML response
func (i *SpotifyImplicitGrant) sendCallbackResponse(w http.ResponseWriter, success bool, errorMsg string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	status := "successful"
	if !success {
		status = fmt.Sprintf("failed (%s)", html.EscapeString(errorMsg))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>Spotify Authorization</title></head>
<body>
	<h1>Authorization %s</h1>
	<p>You can close this window.</p>
	<button onclick="window.close();">Close Window</button>
	<script>window.close();</script>
</body>
</html>`, status)

	w.Write([]byte(html))
}
