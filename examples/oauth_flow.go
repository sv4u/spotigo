package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sv4u/spotigo"
)

// This example demonstrates a complete OAuth authorization flow with
// token caching and error handling.
//
// Prerequisites:
//   - Set SPOTIGO_CLIENT_ID environment variable
//   - Set SPOTIGO_CLIENT_SECRET environment variable
//   - Set SPOTIGO_REDIRECT_URI environment variable
//   - Add redirect URI to your Spotify app settings
//
// Usage:
//   go run examples/oauth_flow.go
func main() {
	clientID := os.Getenv("SPOTIGO_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIGO_CLIENT_SECRET")
	redirectURI := os.Getenv("SPOTIGO_REDIRECT_URI")

	if clientID == "" || clientSecret == "" {
		log.Fatal("SPOTIGO_CLIENT_ID and SPOTIGO_CLIENT_SECRET must be set")
	}

	if redirectURI == "" {
		log.Fatal("SPOTIGO_REDIRECT_URI must be set")
	}

	// Create OAuth manager with scopes
	// You can customize the scopes based on what your application needs
	scopes := "user-read-private user-read-email user-library-read playlist-read-private"
	auth, err := spotigo.NewSpotifyOAuth(
		clientID,
		clientSecret,
		redirectURI,
		scopes,
	)
	if err != nil {
		log.Fatalf("Failed to create OAuth: %v", err)
	}

	// Use file cache to persist tokens across sessions
	cache, err := spotigo.NewFileCacheHandler("", "")
	if err != nil {
		log.Fatalf("Failed to create cache handler: %v", err)
	}
	auth.CacheHandler = cache

	ctx := context.Background()

	// Check for cached token first
	token, err := auth.GetCachedToken(ctx)
	if err != nil {
		log.Fatalf("Failed to check cache: %v", err)
	}

	if token != nil && !auth.IsTokenExpired(token) {
		fmt.Println("Using cached token.")
	} else {
		fmt.Println("No valid cached token. Starting OAuth flow...")

		// Get authorization code
		// The second parameter (true) opens the browser automatically
		// Set to false for headless environments
		code, err := auth.GetAuthorizationCode(ctx, true)
		if err != nil {
			log.Fatalf("Failed to get authorization code: %v", err)
		}

		fmt.Println("Authorization code received. Exchanging for tokens...")

		// Exchange authorization code for access and refresh tokens
		if err := auth.ExchangeCode(ctx, code); err != nil {
			log.Fatalf("Failed to exchange code: %v", err)
		}

		fmt.Println("Token exchange successful!")
	}

	// Create client with the authenticated auth manager
	client, err := spotigo.NewClient(auth)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Test the connection by getting user profile
	user, err := client.CurrentUser(ctx)
	if err != nil {
		log.Fatalf("Failed to get user profile: %v", err)
	}

	fmt.Printf("\nSuccessfully authenticated as: %s\n", user.DisplayName)
	fmt.Printf("User ID: %s\n", user.ID)
	fmt.Printf("Product: %s\n", user.Product)

	// Example: Get user's playlists
	fmt.Println("\nFetching user playlists...")
	playlists, err := client.CurrentUserPlaylists(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to get playlists: %v", err)
	}

	fmt.Printf("Found %d playlists:\n", playlists.Total)
	for i, playlist := range playlists.Items {
		if i >= 5 { // Show first 5
			break
		}
		fmt.Printf("  - %s (%d tracks)\n", playlist.Name, playlist.Tracks.Total)
	}

	if playlists.Total > 5 {
		fmt.Printf("  ... and %d more\n", playlists.Total-5)
	}
}
