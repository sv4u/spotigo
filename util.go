package spotigo

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// GetHostPort parses host and port from a network location string (host:port format)
// Returns (host, port) where port is nil if not specified
func GetHostPort(netloc string) (string, *int) {
	if netloc == "" {
		return "", nil
	}

	// Check if ":" is in the string
	if idx := strings.Index(netloc, ":"); idx != -1 {
		host := netloc[:idx]
		portStr := netloc[idx+1:]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			// Invalid port format, return host only
			return netloc, nil
		}
		return host, &port
	}

	// No port specified
	return netloc, nil
}

// ParseAuthResponseURL parses the authorization response URL to extract code and state
// Returns (code, state, error)
func ParseAuthResponseURL(responseURL string) (string, string, error) {
	parsedURL, err := url.Parse(responseURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL: %w", err)
	}

	// Extract code and state from query parameters
	code := parsedURL.Query().Get("code")
	state := parsedURL.Query().Get("state")

	// Check for error in query parameters
	if errorParam := parsedURL.Query().Get("error"); errorParam != "" {
		errorDesc := parsedURL.Query().Get("error_description")
		return "", "", &SpotifyOAuthError{
			ErrorType:        errorParam,
			ErrorDescription: errorDesc,
		}
	}

	return code, state, nil
}

// GenerateRandomState generates a random state string for CSRF protection
func GenerateRandomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Valid entity types for Spotify URIs/URLs
var validEntityTypes = map[string]bool{
	"track":     true,
	"artist":    true,
	"album":     true,
	"playlist":  true,
	"show":      true,
	"episode":   true,
	"audiobook": true,
	"user":      true,
}

// Spotify URI pattern: spotify:track:ID or spotify:user:username:playlist:ID
var spotifyURIPattern = regexp.MustCompile(`^spotify:(?:(?P<type>track|artist|album|playlist|show|episode|audiobook):(?P<id>[0-9A-Za-z]+)|user:(?P<username>[0-9A-Za-z]+):playlist:(?P<playlistid>[0-9A-Za-z]+))$`)

// Spotify URL pattern: https://open.spotify.com/track/ID (with optional intl-XX/ or intl-XX-YY/)
// Handles query parameters, fragments, and trailing slashes
var spotifyURLPattern = regexp.MustCompile(`^https?://(?:(?:www|open)\.)?spotify\.com/(?:(?:intl-[a-z]{2}(?:-[A-Z]{2})?/)?)?(?P<type>track|artist|album|playlist|show|episode|audiobook|user)/(?P<id>[0-9A-Za-z]+)(?:[/?#].*)?$`)

// Base62 pattern for raw IDs
var base62Pattern = regexp.MustCompile(`^[0-9A-Za-z]+$`)

// GetID extracts a Spotify ID from URI, URL, or raw ID
// entityType is optional and used for validation
func GetID(uri string, entityType ...string) (string, error) {
	var expectedType string
	if len(entityType) > 0 {
		expectedType = entityType[0]
	}

	// Check if URI format
	if strings.HasPrefix(uri, "spotify:") {
		return parseURI(uri, expectedType)
	}

	// Check if URL format
	if strings.Contains(uri, "open.spotify.com") || strings.Contains(uri, "spotify.com") {
		return parseURL(uri, expectedType)
	}

	// Assume raw ID
	if expectedType != "" {
		if err := validateEntityType("", expectedType); err != nil {
			return "", err
		}
	}

	// Validate base62
	if !isValidBase62(uri) {
		return "", fmt.Errorf("invalid base62 ID: %s", uri)
	}

	return uri, nil
}

// GetURI converts an ID to Spotify URI format
func GetURI(id, entityType string) (string, error) {
	if err := validateEntityType(entityType, ""); err != nil {
		return "", err
	}

	if !isValidBase62(id) {
		return "", fmt.Errorf("invalid base62 ID: %s", id)
	}

	return fmt.Sprintf("spotify:%s:%s", entityType, id), nil
}

// IsURI checks if a string is a valid Spotify URI
func IsURI(uri string) bool {
	return spotifyURIPattern.MatchString(uri)
}

// parseURI parses a Spotify URI and extracts the ID
func parseURI(uri string, expectedType string) (string, error) {
	matches := spotifyURIPattern.FindStringSubmatch(uri)
	if matches == nil {
		return "", fmt.Errorf("invalid Spotify URI format: %s", uri)
	}

	// Extract type and ID from named groups
	typeIdx := spotifyURIPattern.SubexpIndex("type")
	idIdx := spotifyURIPattern.SubexpIndex("id")
	playlistIDIdx := spotifyURIPattern.SubexpIndex("playlistid")

	var entityType, id string
	if typeIdx >= 0 && idIdx >= 0 && matches[typeIdx] != "" && matches[idIdx] != "" {
		entityType = matches[typeIdx]
		id = matches[idIdx]
	} else if playlistIDIdx >= 0 && matches[playlistIDIdx] != "" {
		// User playlist format: spotify:user:username:playlist:ID
		entityType = "playlist"
		id = matches[playlistIDIdx]
	} else {
		return "", fmt.Errorf("invalid Spotify URI format: %s", uri)
	}

	// Validate entity type if expected
	if expectedType != "" && entityType != expectedType {
		return "", fmt.Errorf("entity type mismatch: expected %s, got %s", expectedType, entityType)
	}

	return id, nil
}

// parseURL parses a Spotify URL and extracts the ID
func parseURL(urlStr string, expectedType string) (string, error) {
	matches := spotifyURLPattern.FindStringSubmatch(urlStr)
	if matches == nil {
		return "", fmt.Errorf("invalid Spotify URL format: %s", urlStr)
	}

	typeIdx := spotifyURLPattern.SubexpIndex("type")
	idIdx := spotifyURLPattern.SubexpIndex("id")

	if typeIdx < 0 || idIdx < 0 {
		return "", fmt.Errorf("invalid Spotify URL format: %s", urlStr)
	}

	entityType := matches[typeIdx]
	id := matches[idIdx]

	// Validate entity type if expected
	if expectedType != "" && entityType != expectedType {
		return "", fmt.Errorf("entity type mismatch: expected %s, got %s", expectedType, entityType)
	}

	return id, nil
}

// validateEntityType validates that an entity type is supported
func validateEntityType(actual, expected string) error {
	if expected != "" {
		if !validEntityTypes[expected] {
			return fmt.Errorf("unsupported entity type: %s", expected)
		}
	}
	if actual != "" {
		if !validEntityTypes[actual] {
			return fmt.Errorf("unsupported entity type: %s", actual)
		}
		if expected != "" && actual != expected {
			return fmt.Errorf("entity type mismatch: expected %s, got %s", expected, actual)
		}
	}
	return nil
}

// isValidBase62 checks if a string is a valid base62 ID
func isValidBase62(id string) bool {
	if len(id) == 0 {
		return false
	}
	// Spotify IDs are typically 22 characters, but can vary
	// Just check that it's base62 characters
	return base62Pattern.MatchString(id)
}

// SupportedCountryCodes is the list of supported country codes (96 countries as of latest API)
// TODO: Consider fetching from Spotify's /markets endpoint for accuracy
var SupportedCountryCodes = map[string]bool{
	"AD": true, "AE": true, "AG": true, "AL": true, "AM": true, "AO": true, "AR": true, "AT": true, "AU": true, "AZ": true,
	"BA": true, "BB": true, "BD": true, "BE": true, "BF": true, "BG": true, "BH": true, "BI": true, "BJ": true, "BN": true,
	"BO": true, "BR": true, "BS": true, "BT": true, "BW": true, "BY": true, "BZ": true, "CA": true, "CH": true, "CI": true,
	"CL": true, "CM": true, "CO": true, "CR": true, "CV": true, "CW": true, "CY": true, "CZ": true, "DE": true, "DJ": true,
	"DK": true, "DM": true, "DO": true, "DZ": true, "EC": true, "EE": true, "EG": true, "ES": true, "FI": true, "FJ": true,
	"FM": true, "FR": true, "GA": true, "GB": true, "GD": true, "GE": true, "GH": true, "GM": true, "GN": true, "GQ": true,
	"GR": true, "GT": true, "GW": true, "GY": true, "HK": true, "HN": true, "HR": true, "HT": true, "HU": true, "ID": true,
	"IE": true, "IL": true, "IN": true, "IS": true, "IT": true, "JM": true, "JO": true, "JP": true, "KE": true, "KG": true,
	"KH": true, "KI": true, "KM": true, "KN": true, "KR": true, "KW": true, "KZ": true, "LA": true, "LB": true, "LC": true,
	"LI": true, "LK": true, "LR": true, "LS": true, "LT": true, "LU": true, "LV": true, "MA": true, "MD": true, "ME": true,
	"MG": true, "MH": true, "MK": true, "ML": true, "MN": true, "MO": true, "MR": true, "MT": true, "MU": true, "MV": true,
	"MW": true, "MX": true, "MY": true, "MZ": true, "NA": true, "NE": true, "NG": true, "NI": true, "NL": true, "NO": true,
	"NP": true, "NR": true, "NZ": true, "OM": true, "PA": true, "PE": true, "PG": true, "PH": true, "PK": true, "PL": true,
	"PS": true, "PT": true, "PW": true, "PY": true, "QA": true, "RO": true, "RS": true, "RW": true, "SA": true, "SB": true,
	"SC": true, "SE": true, "SG": true, "SI": true, "SK": true, "SL": true, "SM": true, "SN": true, "SR": true, "ST": true,
	"SV": true, "SZ": true, "TD": true, "TG": true, "TH": true, "TL": true, "TN": true, "TO": true, "TR": true, "TT": true,
	"TV": true, "TW": true, "TZ": true, "UA": true, "UG": true, "US": true, "UY": true, "UZ": true, "VC": true, "VE": true,
	"VN": true, "VU": true, "WS": true, "XK": true, "ZA": true, "ZM": true, "ZW": true,
}

// ValidateCountryCode validates an ISO 3166-1 alpha-2 country code
// Returns true if the code is in Spotify's supported countries list
func ValidateCountryCode(code string) bool {
	if len(code) != 2 {
		return false
	}
	upperCode := strings.ToUpper(code)
	return SupportedCountryCodes[upperCode]
}
