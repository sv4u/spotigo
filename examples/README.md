# Examples

This directory contains example programs demonstrating how to use Spotigo.

All examples use the module path `github.com/sv4u/spotigo`.

## Running Examples

Each example is a standalone program. Run them individually using `go run`:

```bash
# Search example (no authentication required)
go run examples/basic_search.go

# OAuth flow example
go run examples/oauth_flow.go

# User profile example
go run examples/user_profile.go
```

## Prerequisites

Before running the examples, you'll need:

1. **Spotify Developer Account** - Sign up at [developer.spotify.com](https://developer.spotify.com/)
2. **Spotify App** - Create an app in the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
3. **Environment Variables** - Set the following before running examples:

   - `SPOTIGO_CLIENT_ID` - Your Spotify app client ID
   - `SPOTIGO_CLIENT_SECRET` - Your Spotify app client secret
   - `SPOTIGO_REDIRECT_URI` - OAuth redirect URI (required for OAuth examples, e.g., `http://localhost:8080/callback`)
   - `SPOTIGO_CLIENT_USERNAME` - Username for token caching (optional)

   Example:
   ```bash
   export SPOTIGO_CLIENT_ID="your_client_id"
   export SPOTIGO_CLIENT_SECRET="your_client_secret"
   export SPOTIGO_REDIRECT_URI="http://localhost:8080/callback"
   ```

## Note

These examples are separate programs and cannot be built together. They are meant to be run individually with `go run`.
