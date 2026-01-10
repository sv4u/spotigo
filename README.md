# Spotigo

[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Spotigo is a Go client library for the Spotify Web API, providing full access to Spotify's music catalog, user data, and playback controls with type safety and idiomatic Go patterns.

## Features

- **Complete Spotify Web API coverage** - 120+ endpoints implemented
- **Type-safe Go structs** - All API responses are strongly-typed structs
- **OAuth2 authentication flows** - Client Credentials, Authorization Code, PKCE, and Implicit Grant
- **Token caching** - File-based and in-memory caching with secure defaults
- **Automatic retry** - Exponential backoff with configurable retry policies
- **Rate limiting support** - Automatic handling of rate limit responses
- **Context support** - Full `context.Context` support for cancellation and timeouts
- **Comprehensive error handling** - Typed errors matching Spotify API responses
- **Zero external dependencies** - Uses only Go standard library (except for testing)

## Installation

```bash
go get github.com/sv4u/spotigo
```

Requires Go 1.23 or later.

## Prerequisites

Before using Spotigo, you'll need:

1. **Go 1.23 or later** - Download from [golang.org](https://golang.org/dl/)
2. **Spotify Developer Account** - Sign up at [developer.spotify.com](https://developer.spotify.com/)
3. **Spotify App** - Create an app in the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard) to get:
   - Client ID
   - Client Secret (for server-side applications)
   - Redirect URI (for OAuth flows)

Once you have your credentials, set them as environment variables:
```bash
export SPOTIPY_CLIENT_ID="your_client_id"
export SPOTIPY_CLIENT_SECRET="your_client_secret"
export SPOTIPY_REDIRECT_URI="http://localhost:8080/callback"  # For OAuth flows
```

## Quick Start

### Client Credentials Flow (No User Authentication)

Perfect for accessing public Spotify data without user authentication:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sv4u/spotigo"
)

func main() {
	// Get credentials from environment variables
	clientID := os.Getenv("SPOTIPY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIPY_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("SPOTIPY_CLIENT_ID and SPOTIPY_CLIENT_SECRET must be set")
	}

	// Create authentication manager
	auth, err := spotigo.NewClientCredentials(clientID, clientSecret)
	if err != nil {
		log.Fatalf("Failed to create auth: %v", err)
	}

	// Create client
	client, err := spotigo.NewClient(auth)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Search for tracks
	results, err := client.Search(ctx, "weezer", "track", &spotigo.SearchOptions{
		Limit: 20,
	})
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	// Display results
	fmt.Println("Search Results:")
	for i, track := range results.Tracks.Items {
		if track != nil {
			artistName := "Unknown"
			if len(track.Artists) > 0 && track.Artists[0] != nil {
				artistName = track.Artists[0].Name
			}
			fmt.Printf("%d. %s by %s\n", i+1, track.Name, artistName)
		}
	}
}
```

### OAuth Flow (User Authentication)

For accessing user-specific data like playlists, saved tracks, and playback control:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sv4u/spotigo"
)

func main() {
	clientID := os.Getenv("SPOTIPY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIPY_CLIENT_SECRET")
	redirectURI := os.Getenv("SPOTIPY_REDIRECT_URI")

	if clientID == "" || clientSecret == "" {
		log.Fatal("SPOTIPY_CLIENT_ID and SPOTIPY_CLIENT_SECRET must be set")
	}

	// Create OAuth manager with required scopes
	auth, err := spotigo.NewSpotifyOAuth(
		clientID,
		clientSecret,
		redirectURI,
		"user-read-private user-read-email user-library-read",
	)
	if err != nil {
		log.Fatalf("Failed to create OAuth: %v", err)
	}

	ctx := context.Background()

	// Get authorization code (opens browser automatically)
	code, err := auth.GetAuthorizationCode(ctx, true)
	if err != nil {
		log.Fatalf("Failed to get authorization code: %v", err)
	}

	// Exchange code for tokens
	if err := auth.ExchangeCode(ctx, code); err != nil {
		log.Fatalf("Failed to exchange code: %v", err)
	}

	// Create client
	client, err := spotigo.NewClient(auth)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Get current user profile
	user, err := client.CurrentUser(ctx)
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}

	fmt.Printf("Logged in as: %s (%s)\n", user.DisplayName, user.ID)

	// Get saved tracks
	tracks, err := client.CurrentUserSavedTracks(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to get saved tracks: %v", err)
	}

	fmt.Println("\nYour Saved Tracks:")
	for _, item := range tracks.Items {
		if item != nil && item.Track != nil {
			artistName := "Unknown"
			if len(item.Track.Artists) > 0 && item.Track.Artists[0] != nil {
				artistName = item.Track.Artists[0].Name
			}
			fmt.Printf("  %s – %s\n", artistName, item.Track.Name)
		}
	}
}
```

## Authentication

Spotigo supports all OAuth2 flows required by the Spotify Web API:

### Client Credentials Flow

Use for accessing public data without user authentication:

```go
auth, err := spotigo.NewClientCredentials(clientID, clientSecret)
client, err := spotigo.NewClient(auth)
```

### Authorization Code Flow

Use for accessing user-specific data:

```go
auth, err := spotigo.NewSpotifyOAuth(
	clientID,
	clientSecret,
	redirectURI,
	"user-read-private user-read-email",
)
code, err := auth.GetAuthorizationCode(ctx, true)
err = auth.ExchangeCode(ctx, code)
client, err := spotigo.NewClient(auth)
```

### PKCE Flow

Use for public clients (mobile apps, SPAs) without client secret:

```go
auth, err := spotigo.NewSpotifyPKCE(
	clientID,
	redirectURI,
	"user-read-private",
)
code, err := auth.GetAuthorizationCode(ctx, true)
err = auth.ExchangeCode(ctx, code)
client, err := spotigo.NewClient(auth)
```

### Environment Variables

Spotigo supports environment variables for configuration:

- `SPOTIPY_CLIENT_ID` - Your Spotify app client ID
- `SPOTIPY_CLIENT_SECRET` - Your Spotify app client secret
- `SPOTIPY_REDIRECT_URI` - OAuth redirect URI
- `SPOTIPY_CLIENT_USERNAME` - Username for token caching

## Client Configuration

### Basic Client

```go
client, err := spotigo.NewClient(auth)
```

### Client with Options

```go
import (
	"net/http"
	"time"

	"github.com/sv4u/spotigo"
)

// Custom HTTP client
httpClient := &http.Client{
	Timeout: 30 * time.Second,
}

// Custom cache handler (set on auth manager, not client)
cache := spotigo.NewFileCacheHandler("/path/to/cache", "username")
auth.CacheHandler = cache

// Custom retry configuration
retryConfig := &spotigo.RetryConfig{
	MaxRetries:     3,
	StatusRetries: 3,
	StatusForcelist: []int{429, 500, 502, 503, 504},
	BackoffFactor:  0.3,
}

// Create client with options
client, err := spotigo.NewClient(
	auth,
	spotigo.WithHTTPClient(httpClient),
	spotigo.WithRetryConfig(retryConfig),
	spotigo.WithLanguage("en"),
	spotigo.WithRequestTimeout(10*time.Second),
)
```

## API Usage Examples

### Search

```go
import (
	"context"
	"fmt"
	"log"

	"github.com/sv4u/spotigo"
)

ctx := context.Background()

// Search for tracks
results, err := client.Search(ctx, "Blinding Lights", "track", &spotigo.SearchOptions{
	Limit:  10,
	Market: "US",
})
if err != nil {
	log.Fatal(err)
}

for _, track := range results.Tracks.Items {
	if track != nil {
		fmt.Println(track.Name)
	}
}
```

### Get Track by ID, URI, or URL

Spotigo accepts multiple input formats and automatically parses them:

```go
import (
	"context"
	"github.com/sv4u/spotigo"
)

// All of these work:
track1, _ := client.Track(ctx, "4iV5W9uYEdYUVa79Axb7Rh")                    // Raw ID
track2, _ := client.Track(ctx, "spotify:track:4iV5W9uYEdYUVa79Axb7Rh")    // URI
track3, _ := client.Track(ctx, "https://open.spotify.com/track/4iV5W9uYEdYUVa79Axb7Rh") // URL
```

### Pagination

```go
import (
	"context"
	"fmt"
	"log"

	"github.com/sv4u/spotigo"
)

// Get first page
tracks, err := client.AlbumTracks(ctx, albumID, nil)
if err != nil {
	log.Fatal(err)
}

// Iterate through all pages
for tracks != nil {
	for _, item := range tracks.Items {
		if item != nil && item.Track != nil {
			fmt.Println(item.Track.Name)
		}
	}

	// Get next page
	if tracks.Next != "" {
		tracks, err = tracks.Next()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		break
	}
}
```

### Error Handling

```go
import (
	"context"
	"fmt"

	"github.com/sv4u/spotigo"
)

track, err := client.Track(ctx, trackID)
if err != nil {
	// Check error type
	if spotifyErr, ok := err.(*spotigo.SpotifyError); ok {
		switch spotifyErr.HTTPStatus {
		case 404:
			fmt.Println("Track not found")
		case 401:
			fmt.Println("Authentication required")
		case 429:
			fmt.Println("Rate limit exceeded")
			if delay, ok := spotifyErr.RetryAfter(); ok {
				fmt.Printf("Retry after: %v\n", delay)
			}
		default:
			fmt.Printf("API error: %v\n", spotifyErr)
		}
	} else {
		fmt.Printf("Unexpected error: %v\n", err)
	}
	return
}

// Use track
fmt.Println(track.Name)
```

### Context for Timeouts and Cancellation

```go
import (
	"context"
	"fmt"
	"time"

	"github.com/sv4u/spotigo"
)

// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Use context in API call
track, err := client.Track(ctx, trackID)
if err != nil {
	if err == context.DeadlineExceeded {
		fmt.Println("Request timed out")
	} else if err == context.Canceled {
		fmt.Println("Request canceled")
	} else {
		fmt.Printf("Error: %v\n", err)
	}
}
```

## Type Safety

All API responses are strongly-typed Go structs:

```go
import (
	"context"
	"fmt"
	"log"

	"github.com/sv4u/spotigo"
)

// Track is a typed struct, not a map
track, err := client.Track(ctx, trackID)
if err != nil {
	log.Fatal(err)
}

// Direct field access with type safety
fmt.Println(track.Name)           // string
fmt.Println(track.DurationMs)     // int
fmt.Println(track.Popularity)     // int
fmt.Println(track.Explicit)      // bool

// Nested structs are also typed
if len(track.Artists) > 0 {
	artist := track.Artists[0]
	fmt.Println(artist.Name)      // string
	fmt.Println(artist.ID)        // string
}

// Optional fields use pointers
if track.Album != nil {
	fmt.Println(track.Album.Name)
}
```

## Token Caching

Token caching helps avoid unnecessary re-authentication. You can configure caching when setting up your authentication manager.

### File Cache (Default)

Tokens are automatically cached to a file when you set a cache handler:

```go
// Cache to default location (.cache-username)
cache := spotigo.NewFileCacheHandler("", "username")
auth.CacheHandler = cache
```

**Note:** The cache handler must be set on the `auth` manager (before creating the client), not on the client itself.

### Memory Cache

For short-lived applications or testing (tokens are lost when the program exits):

```go
cache := spotigo.NewMemoryCacheHandler()
auth.CacheHandler = cache
```

### Custom Cache Path

```go
cache := spotigo.NewFileCacheHandler("/path/to/cache.json", "")
auth.CacheHandler = cache
```

## Retry Logic

Spotigo automatically retries failed requests with exponential backoff:

```go
import (
	"github.com/sv4u/spotigo"
)

retryConfig := &spotigo.RetryConfig{
	MaxRetries:     3,              // Maximum retry attempts
	StatusRetries: 3,               // Retries for status codes
	StatusForcelist: []int{429, 500, 502, 503, 504}, // Retryable status codes
	BackoffFactor:  0.3,            // Backoff multiplier
	RetryAfterHeader: true,         // Respect Retry-After header
}

client, err := spotigo.NewClient(auth, spotigo.WithRetryConfig(retryConfig))
```

## Examples

See the [examples](./examples/) directory for complete, runnable examples:

- `basic_search.go` - Search without authentication
- `oauth_flow.go` - Complete OAuth authorization flow
- `user_profile.go` - Get user info and saved content

Each example is a standalone program. Run them individually:

```bash
go run examples/basic_search.go
go run examples/oauth_flow.go
go run examples/user_profile.go
```

**Note:** Examples are separate programs and cannot be built together. They are excluded from `go test ./...` runs. See [examples/README.md](./examples/README.md) for more details.

## Documentation

- [GoDoc](https://pkg.go.dev/github.com/sv4u/spotigo) - Full API documentation
- [Spotify Web API Reference](https://developer.spotify.com/documentation/web-api) - Official Spotify API docs
- [Spotipy Documentation](https://spotipy.readthedocs.org/) - Original Python library (for reference)

## Troubleshooting

### 401 Unauthorized

- Check that your client ID and secret are correct
- Verify required scopes are included in your OAuth flow
- Ensure tokens haven't expired (caching should handle this automatically)

### Rate Limiting

Spotigo automatically handles rate limits with retries. If you're hitting limits frequently:

- Reduce request frequency
- Use pagination to fetch data in batches
- Consider caching responses

### Token Cache Issues

If you're seeing "incorrect user" errors:

- Clear the token cache file
- Use `show_dialog=true` in OAuth flow to force re-authentication
- Check that `SPOTIPY_CLIENT_USERNAME` matches your Spotify username

### Search Not Finding Tracks

- Use the `Market` parameter to specify a country code
- Some tracks may not be available in all markets
- Check track availability using the `AvailableMarkets` field

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](./LICENSE) for details.

## Attribution

This project is a Go rewrite of [Spotipy](https://github.com/spotipy-dev/spotipy), a lightweight Python library for the Spotify Web API. Original Spotipy library by [Paul Lamere](https://github.com/plamere).
